package dynamo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"frpg-backend/internal/domain"
)

// fakeDDB is an in-memory stand-in for the DynamoDB client (the parts the exercise
// store uses), so the marshal/unmarshal roundtrip is exercised without a real DB.
type fakeDDB struct {
	items map[string]map[string]types.AttributeValue
}

func newFakeDDB() *fakeDDB {
	return &fakeDDB{items: map[string]map[string]types.AttributeValue{}}
}

func (f *fakeDDB) PutItem(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	id := in.Item["exerciseId"].(*types.AttributeValueMemberS).Value
	f.items[id] = in.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (f *fakeDDB) GetItem(_ context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	id := in.Key["exerciseId"].(*types.AttributeValueMemberS).Value
	return &dynamodb.GetItemOutput{Item: f.items[id]}, nil
}

func (f *fakeDDB) Scan(_ context.Context, _ *dynamodb.ScanInput, _ ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	out := make([]map[string]types.AttributeValue, 0, len(f.items))
	for _, it := range f.items {
		out = append(out, it)
	}
	return &dynamodb.ScanOutput{Items: out}, nil
}

// A realistic generated item: nested contrast/prompt/origin + a choices list and
// an answer key inside the open Content/Answer maps.
func sampleExercise() domain.Exercise {
	return domain.Exercise{
		ID: "ex_a1_conj_604fd2aa", Skill: "reading", Format: "multiple_choice", Level: "A1",
		Contrast: domain.TargetContrast{SkillPoint: "conjugation", Lemma: "parler", Feature: "person"},
		Prompt:   domain.Prompt{Instructions: "Choisissez la bonne forme du verbe.", Text: "Tu ___ français."},
		Content: map[string]any{
			"choices": []map[string]any{
				{"id": "a", "text": "parle"}, {"id": "b", "text": "parles"},
			},
			"multiple": false,
		},
		Answer:    map[string]any{"correct": []string{"b"}},
		Source:    "generated",
		Origin:    domain.Origin{PromptVersion: "tmpl-present-er-v1", RetrievedRefs: []string{}, CreatedBy: "gen-cloze"},
		CreatedAt: "2026-07-10T00:00:00Z",
	}
}

// TestHandoffRoundtrip walks the whole contract: JSON (the jsonl line) →
// domain.Exercise → DynamoDB item → domain.Exercise, and checks the nested maps
// survive intact. This is what a live seed does, minus the network.
func TestHandoffRoundtrip(t *testing.T) {
	ctx := context.Background()

	// 1. jsonl hop: marshal to a line, read it back (what Python emits / Go imports).
	line, err := json.Marshal(sampleExercise())
	if err != nil {
		t.Fatalf("marshal jsonl: %v", err)
	}
	var e domain.Exercise
	if err := json.Unmarshal(line, &e); err != nil {
		t.Fatalf("unmarshal jsonl: %v", err)
	}

	// 2. dynamo hop: Put then Get through the store.
	store := &ExerciseStore{client: newFakeDDB(), table: "Exercises"}
	if err := store.Put(ctx, e); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := store.Get(ctx, e.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	// 3. nested content survived both hops.
	if got.Contrast.SkillPoint != "conjugation" || got.Contrast.Lemma != "parler" {
		t.Errorf("contrast lost: %+v", got.Contrast)
	}
	if got.Prompt.Text != "Tu ___ français." {
		t.Errorf("prompt lost: %q", got.Prompt.Text)
	}
	choices, ok := got.Content["choices"].([]any)
	if !ok || len(choices) != 2 {
		t.Fatalf("choices lost: %#v", got.Content["choices"])
	}
	correct, ok := got.Answer["correct"].([]any)
	if !ok || len(correct) != 1 || correct[0] != "b" {
		t.Fatalf("answer key lost: %#v", got.Answer["correct"])
	}
}

func TestExerciseStoreGetMissing(t *testing.T) {
	store := &ExerciseStore{client: newFakeDDB(), table: "Exercises"}
	if _, err := store.Get(context.Background(), "nope"); err != domain.ErrExerciseNotFound {
		t.Fatalf("want ErrExerciseNotFound, got %v", err)
	}
}

func TestExerciseStoreQueryLimit(t *testing.T) {
	ctx := context.Background()
	store := &ExerciseStore{client: newFakeDDB(), table: "Exercises"}
	for _, id := range []string{"e1", "e2", "e3"} {
		e := sampleExercise()
		e.ID = id
		_ = store.Put(ctx, e)
	}
	got, err := store.Query(ctx, "A1", "reading", 2)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want limit 2, got %d", len(got))
	}
}
