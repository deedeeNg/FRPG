package generate

import (
	"strings"
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

func TestGenerateCountAndUniformDistribution(t *testing.T) {
	pack := loadPack(t)
	n := 5 * len(pack.Teaches) // divides evenly regardless of how many pools exist
	exs := Generate(pack, n)
	if len(exs) < 50 {
		t.Fatalf("want >= 50 exercises, got %d", len(exs))
	}

	// Group by PromptVersion, not SkillPoint: several teaches entries may
	// deliberately share one SkillPoint (e.g. passe_compose_avoir/etre both feed
	// the "passe_compose" mastery bucket), but PromptVersion is unique per pool —
	// that's the axis Generate() actually spreads evenly across.
	dist := map[string]int{}
	for _, e := range exs {
		dist[e.Origin.PromptVersion]++
	}
	if len(dist) != len(pack.Teaches) {
		t.Fatalf("want %d pools, got %d (%v)", len(pack.Teaches), len(dist), dist)
	}
	lo, hi := 1<<30, 0
	for _, c := range dist {
		if c < lo {
			lo = c
		}
		if c > hi {
			hi = c
		}
	}
	if hi-lo > 1 {
		t.Errorf("distribution not uniform: spread %d..%d (%v)", lo, hi, dist)
	}
}

func TestGenerateItemsAreValid(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range Generate(loadPack(t), 300) {
		if seen[e.ID] {
			t.Errorf("duplicate id %q", e.ID)
		}
		seen[e.ID] = true

		if e.Level != domain.LevelA1 || e.Skill != domain.SkillReading {
			t.Errorf("%s: unexpected level/skill: %q/%q", e.ID, e.Level, e.Skill)
		}
		// Prompt.Text is deliberately empty for clock items (stimulus is
		// Content.clock) and for fill_blank items (stimulus is Content.template).
		if e.Prompt.Text == "" && e.Contrast.SkillPoint != "telling_time" && e.Format != domain.FormatFillBlank {
			t.Errorf("%s: empty prompt text", e.ID)
		}
		if e.Prompt.Instructions == "" {
			t.Errorf("%s: empty prompt instructions", e.ID)
		}
		if e.Contrast.SkillPoint == "" || e.Contrast.Lemma == "" || e.Contrast.Feature == "" {
			t.Errorf("%s: incomplete contrast: %+v", e.ID, e.Contrast)
		}

		switch e.Format {
		case domain.FormatMultipleChoice:
			// The key id must exist and point at a real choice; content must not
			// leak it.
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
		case domain.FormatFillBlank:
			tmpl, ok := e.Content["template"].(string)
			if !ok || !strings.Contains(tmpl, "___") {
				t.Fatalf("%s: bad fill_blank template %#v", e.ID, e.Content["template"])
			}
			accepted, ok := e.Answer["accepted"].(map[string][]string)
			if !ok || len(accepted["1"]) == 0 {
				t.Fatalf("%s: bad fill_blank accepted %#v", e.ID, e.Answer["accepted"])
			}
		default:
			t.Errorf("%s: unexpected format %q", e.ID, e.Format)
		}
	}
}

// At a sufficient n, every verb in every conjugation-kind pool must surface (a
// capped sample must span the whole lexicon, not just the first verbs) — checked
// per pool via PromptVersion, since several pools can share a SkillPoint (e.g.
// passe_compose_avoir/etre) or even reuse a lemma under a different SkillPoint
// (e.g. "partir" appears in both futur_proche_verbs and passe_compose_etre).
func TestGenerateCoversEveryVerb(t *testing.T) {
	pack := loadPack(t)
	// generous: every pool's every verb gets a real chance, incl. pools doubled
	// by alsoFillBlank (MC + fill_blank siblings interleaved per item).
	exs := Generate(pack, 60*len(pack.Teaches))

	for _, gp := range pack.Teaches {
		if gp.Kind != kindConjugation {
			continue
		}
		want := map[string]bool{}
		for _, v := range gp.Verbs {
			want[v.Lemma] = true
		}
		seen := map[string]bool{}
		for _, e := range exs {
			if e.Origin.PromptVersion == gp.PromptVersion {
				seen[e.Contrast.Lemma] = true
			}
		}
		for lemma := range want {
			if !seen[lemma] {
				t.Errorf("%s: verb %q never surfaced", gp.ID, lemma)
			}
		}
	}
}

func TestFrenchNumber(t *testing.T) {
	cases := map[int]string{
		1: "un", 7: "sept", 16: "seize", 17: "dix-sept", 20: "vingt",
		21: "vingt et un", 22: "vingt-deux", 31: "trente et un", 42: "quarante-deux",
		60: "soixante", 61: "soixante et un", 69: "soixante-neuf", 70: "soixante-dix",
		71: "soixante et onze", 72: "soixante-douze", 76: "soixante-seize",
		77: "soixante-dix-sept", 79: "soixante-dix-neuf", 80: "quatre-vingts",
	}
	for n, want := range cases {
		if got := frenchNumber(n); got != want {
			t.Errorf("frenchNumber(%d) = %q, want %q", n, got, want)
		}
	}
	if frenchNumber(81) != "" || frenchNumber(0) != "zéro" {
		t.Errorf("range guard wrong: 81=%q 0=%q", frenchNumber(81), frenchNumber(0))
	}
}

func TestFrenchTimeFormal(t *testing.T) {
	cases := map[[2]int]string{
		{0, 0}: "zéro heure", {1, 0}: "une heure", {9, 5}: "neuf heures cinq",
		{14, 0}: "quatorze heures", {14, 15}: "quatorze heures quinze",
		{21, 15}: "vingt et une heures quinze", {23, 36}: "vingt-trois heures trente-six",
	}
	for hm, want := range cases {
		if got := frenchTime(hm[0], hm[1], true); got != want {
			t.Errorf("frenchTime(%d,%d,formal) = %q, want %q", hm[0], hm[1], got, want)
		}
	}
}

func TestFrenchTimeInformal(t *testing.T) {
	cases := map[[2]int]string{
		{0, 0}: "minuit", {12, 0}: "midi",
		{3, 0}: "trois heures du matin", {3, 15}: "trois heures et quart du matin",
		{3, 30}: "trois heures et demie du matin", {3, 45}: "quatre heures moins le quart du matin",
		{1, 0}: "une heure du matin", {13, 0}: "une heure de l'après-midi",
		{0, 36}: "minuit trente-six", {12, 36}: "midi trente-six",
		{0, 45}:  "une heure moins le quart du matin",
		{11, 45}: "midi moins le quart", {23, 45}: "minuit moins le quart",
		// the exact case from the ask: 23:36 informal = "11:36 pm"
		{23, 36}: "onze heures trente-six du soir",
		{15, 36}: "trois heures trente-six de l'après-midi",
	}
	for hm, want := range cases {
		if got := frenchTime(hm[0], hm[1], false); got != want {
			t.Errorf("frenchTime(%d,%d,informal) = %q, want %q", hm[0], hm[1], got, want)
		}
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
