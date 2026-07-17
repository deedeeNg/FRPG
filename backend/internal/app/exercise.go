package app

import (
	"context"
	"math/rand"
	"strings"

	"frpg-backend/internal/domain"
)

// Exercises is the use case for serving and grading exercises. It orchestrates the
// domain.ExerciseStore port; the delivery layer strips the answer key before any
// item reaches the client.
type Exercises struct {
	store domain.ExerciseStore
}

func NewExercises(store domain.ExerciseStore) *Exercises {
	return &Exercises{store: store}
}

// Next returns one exercise for (level, skill), chosen at random from a sample.
// The returned item still carries its Answer — the ports layer removes it.
func (x *Exercises) Next(ctx context.Context, level, skill string) (domain.Exercise, error) {
	items, err := x.store.Query(ctx, level, skill, 30)
	if err != nil {
		return domain.Exercise{}, err
	}
	if len(items) == 0 {
		return domain.Exercise{}, domain.ErrExerciseNotFound
	}
	return items[rand.Intn(len(items))], nil
}

// Submission is a learner's attempt at one exercise. Exactly one field is
// populated, matching the exercise's Format: Selected (choice ids) for
// multiple_choice, Text (typed answer) for fill_blank.
type Submission struct {
	Selected []string
	Text     string
}

// Grade fetches the stored exercise and checks the submission against its answer
// key, dispatching on the exercise's own Format — the caller doesn't need to know
// which shape of grading a given exercise expects.
func (x *Exercises) Grade(ctx context.Context, id string, sub Submission) (bool, error) {
	e, err := x.store.Get(ctx, id)
	if err != nil {
		return false, err
	}
	if e.Format == domain.FormatFillBlank {
		return correctFillBlank(e.Answer, sub.Text), nil
	}
	return correctChoices(e.Answer, sub.Selected), nil
}

func correctChoices(answer map[string]any, selected []string) bool {
	want := stringSet(answer["correct"])
	if len(want) == 0 || len(want) != len(selected) {
		return false
	}
	for _, id := range selected {
		if !want[id] {
			return false
		}
	}
	return true
}

// correctFillBlank compares a typed answer against the single blank's accepted
// list, case/whitespace-insensitive but NOT accent-insensitive (correct accents
// are part of what A1 spelling should teach).
func correctFillBlank(answer map[string]any, submitted string) bool {
	norm := normalizeAnswer(submitted)
	if norm == "" {
		return false
	}
	for _, accepted := range acceptedAnswers(answer["accepted"]) {
		if normalizeAnswer(accepted) == norm {
			return true
		}
	}
	return false
}

func normalizeAnswer(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// acceptedAnswers reads answer["accepted"]["1"] (our generator only ever produces
// one blank, id "1"). The map value is map[string][]string when freshly built in
// Go, but decodes as map[string]interface{} -> []interface{} after a JSON/
// DynamoDB round-trip — both shapes are handled.
func acceptedAnswers(v any) []string {
	switch m := v.(type) {
	case map[string][]string:
		return m["1"]
	case map[string]interface{}:
		return stringSlice(m["1"])
	}
	return nil
}

// stringSet extracts a set of strings from a value that may be a []string
// (freshly built) or []interface{} of strings (round-tripped through
// DynamoDB/JSON).
func stringSet(v any) map[string]bool {
	out := map[string]bool{}
	for _, s := range stringSlice(v) {
		out[s] = true
	}
	return out
}

func stringSlice(v any) []string {
	switch xs := v.(type) {
	case []string:
		return xs
	case []interface{}:
		out := make([]string, 0, len(xs))
		for _, x := range xs {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
