package app

import (
	"context"
	"math/rand"

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

// Grade fetches the stored exercise and checks the submitted choice ids against
// its answer key. Covers multiple_choice / multi_select: the selected set must
// equal the correct set.
func (x *Exercises) Grade(ctx context.Context, id string, selected []string) (bool, error) {
	e, err := x.store.Get(ctx, id)
	if err != nil {
		return false, err
	}
	return correctChoices(e.Answer, selected), nil
}

func correctChoices(answer map[string]any, selected []string) bool {
	want := idSet(answer["correct"])
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

// idSet extracts choice ids from an answer value that may be a []string (freshly
// built) or []interface{} of strings (round-tripped through DynamoDB/JSON).
func idSet(v any) map[string]bool {
	out := map[string]bool{}
	switch xs := v.(type) {
	case []string:
		for _, s := range xs {
			out[s] = true
		}
	case []interface{}:
		for _, x := range xs {
			if s, ok := x.(string); ok {
				out[s] = true
			}
		}
	}
	return out
}
