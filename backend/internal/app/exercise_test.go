package app

import (
	"context"
	"testing"

	"frpg-backend/internal/adapters/inmem"
	"frpg-backend/internal/domain"
)

func mcq(id string, correct []string) domain.Exercise {
	return domain.Exercise{
		ID: id, Skill: "reading", Format: "multiple_choice", Level: "A1",
		Content: map[string]any{"choices": []map[string]any{{"id": "a"}, {"id": "b"}, {"id": "c"}}},
		Answer:  map[string]any{"correct": correct},
	}
}

func TestExercisesNextAndGrade(t *testing.T) {
	ctx := context.Background()
	store := inmem.NewExerciseStore()
	_ = store.Put(ctx, mcq("ex1", []string{"b"}))
	x := NewExercises(store)

	got, err := x.Next(ctx, "A1", "reading")
	if err != nil || got.ID != "ex1" {
		t.Fatalf("next: %v, %+v", err, got)
	}

	// correct answer
	if ok, err := x.Grade(ctx, "ex1", []string{"b"}); err != nil || !ok {
		t.Errorf("grade correct: got ok=%v err=%v, want true", ok, err)
	}
	// wrong answer
	if ok, _ := x.Grade(ctx, "ex1", []string{"a"}); ok {
		t.Errorf("grade wrong: got true, want false")
	}
	// empty submission
	if ok, _ := x.Grade(ctx, "ex1", nil); ok {
		t.Errorf("grade empty: got true, want false")
	}
}

// Answers round-tripped through JSON/DynamoDB arrive as []interface{}, not
// []string — grading must handle both.
func TestGradeHandlesInterfaceSlice(t *testing.T) {
	e := domain.Exercise{Answer: map[string]any{"correct": []interface{}{"a", "c"}}}
	if !correctChoices(e.Answer, []string{"a", "c"}) {
		t.Error("want correct for matching multi-select")
	}
	if correctChoices(e.Answer, []string{"a"}) {
		t.Error("partial selection must not be correct")
	}
}

func TestNextEmptyStore(t *testing.T) {
	x := NewExercises(inmem.NewExerciseStore())
	if _, err := x.Next(context.Background(), "A1", "reading"); err != domain.ErrExerciseNotFound {
		t.Fatalf("want ErrExerciseNotFound, got %v", err)
	}
}
