package app

import (
	"context"
	"testing"

	"frpg-backend/internal/adapters/inmem"
	"frpg-backend/internal/domain"
)

func mcq(id string, correct []string) domain.Exercise {
	return domain.Exercise{
		ID: id, Skill: "reading", Format: domain.FormatMultipleChoice, Level: "A1",
		Content: map[string]any{"choices": []map[string]any{{"id": "a"}, {"id": "b"}, {"id": "c"}}},
		Answer:  map[string]any{"correct": correct},
	}
}

func fillBlank(id string, accepted []string) domain.Exercise {
	return domain.Exercise{
		ID: id, Skill: "reading", Format: domain.FormatFillBlank, Level: "A1",
		Content: map[string]any{"template": "Tu ___ français.", "blanks": []map[string]any{{"id": "1"}}},
		Answer:  map[string]any{"accepted": map[string][]string{"1": accepted}},
	}
}

func TestExercisesNextAndGradeMultipleChoice(t *testing.T) {
	ctx := context.Background()
	store := inmem.NewExerciseStore()
	_ = store.Put(ctx, mcq("ex1", []string{"b"}))
	x := NewExercises(store)

	got, err := x.Next(ctx, "A1", "reading")
	if err != nil || got.ID != "ex1" {
		t.Fatalf("next: %v, %+v", err, got)
	}

	if ok, err := x.Grade(ctx, "ex1", Submission{Selected: []string{"b"}}); err != nil || !ok {
		t.Errorf("grade correct: got ok=%v err=%v, want true", ok, err)
	}
	if ok, _ := x.Grade(ctx, "ex1", Submission{Selected: []string{"a"}}); ok {
		t.Errorf("grade wrong: got true, want false")
	}
	if ok, _ := x.Grade(ctx, "ex1", Submission{}); ok {
		t.Errorf("grade empty: got true, want false")
	}
}

func TestGradeFillBlank(t *testing.T) {
	ctx := context.Background()
	store := inmem.NewExerciseStore()
	_ = store.Put(ctx, fillBlank("fb1", []string{"parles"}))
	x := NewExercises(store)

	cases := []struct {
		text string
		want bool
	}{
		{"parles", true},
		{"Parles", true},     // case-insensitive
		{"  parles  ", true}, // trims whitespace
		{"parle", false},     // wrong form
		{"parlés", false},    // accents matter
		{"", false},          // empty submission
	}
	for _, c := range cases {
		got, err := x.Grade(ctx, "fb1", Submission{Text: c.text})
		if err != nil {
			t.Fatalf("grade %q: %v", c.text, err)
		}
		if got != c.want {
			t.Errorf("grade %q = %v, want %v", c.text, got, c.want)
		}
	}
}

func TestGradeFillBlankMultipleAccepted(t *testing.T) {
	ctx := context.Background()
	store := inmem.NewExerciseStore()
	_ = store.Put(ctx, fillBlank("fb2", []string{"bonjour", "salut"}))
	x := NewExercises(store)

	for _, text := range []string{"bonjour", "salut"} {
		if ok, _ := x.Grade(ctx, "fb2", Submission{Text: text}); !ok {
			t.Errorf("grade %q: want true", text)
		}
	}
	if ok, _ := x.Grade(ctx, "fb2", Submission{Text: "coucou"}); ok {
		t.Error("grade unrelated word: want false")
	}
}

func TestGradeHandlesInterfaceSlice(t *testing.T) {
	e := domain.Exercise{Answer: map[string]any{"correct": []interface{}{"a", "c"}}}
	if !correctChoices(e.Answer, []string{"a", "c"}) {
		t.Error("want correct for matching multi-select")
	}
	if correctChoices(e.Answer, []string{"a"}) {
		t.Error("partial selection must not be correct")
	}
}

// answer["accepted"] round-trips through JSON/DynamoDB as
// map[string]interface{} -> []interface{}, not the map[string][]string a fresh
// build produces — grading must handle both.
func TestGradeFillBlankHandlesInterfaceMap(t *testing.T) {
	answer := map[string]any{
		"accepted": map[string]interface{}{"1": []interface{}{"bonjour", "salut"}},
	}
	if !correctFillBlank(answer, "salut") {
		t.Error("want correct for interface-map-shaped accepted list")
	}
	if correctFillBlank(answer, "coucou") {
		t.Error("want incorrect for a word not in the accepted list")
	}
}

func TestNextEmptyStore(t *testing.T) {
	x := NewExercises(inmem.NewExerciseStore())
	if _, err := x.Next(context.Background(), "A1", "reading"); err != domain.ErrExerciseNotFound {
		t.Fatalf("want ErrExerciseNotFound, got %v", err)
	}
}
