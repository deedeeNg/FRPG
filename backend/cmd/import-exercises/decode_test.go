package main

import (
	"bufio"
	"os"
	"testing"
)

// The importer accepts any valid Exercise JSON, whoever produced it. This runs the
// hand-authored (non-template) examples — a reading-comprehension item and a
// listening dialogue, the kind an LLM would emit — through the real decode path to
// prove the pipeline is not tied to the YAML template generator.
func TestDecodeAuthoredExamples(t *testing.T) {
	f, err := os.Open("../../content/exercises/authored_examples.jsonl")
	if err != nil {
		t.Fatalf("open authored examples: %v", err)
	}
	defer f.Close()

	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		e, err := decodeExercise(sc.Bytes())
		if err != nil {
			t.Fatalf("authored item failed to decode: %v", err)
		}
		// These items carry payloads no template produces: a passage/audio prompt
		// plus a free-text question inside the open Content map.
		if _, ok := e.Content["question"]; !ok {
			t.Errorf("%s: expected a free-text 'question' in content", e.ID)
		}
		n++
	}
	if n == 0 {
		t.Fatal("no authored examples found")
	}
}
