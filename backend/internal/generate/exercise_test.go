package generate

import (
	"testing"

	"frpg-backend/internal/domain"
)

// the real level pack, relative to this package dir (backend/internal/generate).
const packPath = "../../content/levels/a1.yaml"

func loadPack(t *testing.T) *LevelPack {
	t.Helper()
	p, err := LoadPack(packPath)
	if err != nil {
		t.Fatalf("load %s: %v", packPath, err)
	}
	return p
}

func TestGenerateCountAndUniformSkillDistribution(t *testing.T) {
	pack := loadPack(t)
	exs := Generate(pack, 51)
	if len(exs) < 50 {
		t.Fatalf("want >= 50 exercises, got %d", len(exs))
	}

	dist := map[string]int{}
	for _, e := range exs {
		dist[e.Contrast.SkillPoint]++
	}
	// One skill point per grammar point, spread uniformly (counts differ by <= 1).
	if len(dist) != len(pack.Teaches) {
		t.Fatalf("want %d skill points, got %d (%v)", len(pack.Teaches), len(dist), dist)
	}
	lo, hi := 1<<30, 0
	for _, n := range dist {
		if n < lo {
			lo = n
		}
		if n > hi {
			hi = n
		}
	}
	if hi-lo > 1 {
		t.Errorf("distribution not uniform: spread %d..%d (%v)", lo, hi, dist)
	}
}

func TestGenerateItemsAreValid(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range Generate(loadPack(t), 51) {
		if seen[e.ID] {
			t.Errorf("duplicate id %q", e.ID)
		}
		seen[e.ID] = true

		if e.Level != domain.LevelA1 || e.Skill != domain.SkillReading || e.Format != domain.FormatMultipleChoice {
			t.Errorf("%s: unexpected level/skill/format: %q/%q/%q", e.ID, e.Level, e.Skill, e.Format)
		}
		if e.Prompt.Text == "" || e.Prompt.Instructions == "" {
			t.Errorf("%s: empty prompt", e.ID)
		}
		if e.Contrast.SkillPoint == "" || e.Contrast.Lemma == "" || e.Contrast.Feature == "" {
			t.Errorf("%s: incomplete contrast: %+v", e.ID, e.Contrast)
		}

		// The key id must exist and point at a real choice; content must not leak it.
		choices, ok := e.Content["choices"].([]map[string]any)
		if !ok || len(choices) < 2 {
			t.Fatalf("%s: bad choices %#v", e.ID, e.Content["choices"])
		}
		correct, ok := e.Answer["correct"].([]string)
		if !ok || len(correct) != 1 {
			t.Fatalf("%s: bad answer %#v", e.ID, e.Answer["correct"])
		}
		ids := map[string]bool{}
		for _, c := range choices {
			ids[c["id"].(string)] = true
		}
		if !ids[correct[0]] {
			t.Errorf("%s: correct id %q not among choices", e.ID, correct[0])
		}
	}
}

// At a sufficient n, every verb in the pack must surface (a capped sample must
// span the whole lexicon, not just the first verbs).
func TestGenerateCoversEveryVerb(t *testing.T) {
	pack := loadPack(t)
	var wantVerbs int
	for _, gp := range pack.Teaches {
		if gp.Kind == kindConjugation {
			wantVerbs = len(gp.Verbs)
		}
	}
	if wantVerbs == 0 {
		t.Skip("no conjugation pack")
	}
	seen := map[string]bool{}
	for _, e := range Generate(pack, 140) {
		if e.Contrast.SkillPoint == "conjugation" {
			seen[e.Contrast.Lemma] = true
		}
	}
	if len(seen) != wantVerbs {
		t.Errorf("want all %d verbs to surface, got %d (%v)", wantVerbs, len(seen), seen)
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	a := Generate(loadPack(t), 51)
	b := Generate(loadPack(t), 51)
	if len(a) != len(b) {
		t.Fatalf("length differs: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Prompt.Text != b[i].Prompt.Text {
			t.Fatalf("item %d differs between runs: %q vs %q", i, a[i].ID, b[i].ID)
		}
	}
}
