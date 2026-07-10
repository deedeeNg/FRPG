// Package generate produces French exercises from a YAML level pack (grammar/skill
// taxonomy + curated lexicon). It is the v0 stand-in for the future Python pipeline
// (pull labeled corpus → parse → blank → lexicon distractors → template); it emits
// the same domain.Exercise shape, so swapping in the real pipeline changes nothing
// downstream.
//
// The *config and data* live in YAML (see content/levels/a1.yaml) — the source of
// truth for what a level teaches; this file holds only the *assembly procedure*.
// Each grammar point declares a `kind` selecting one of three data-driven builders,
// so adding a new skill pack is usually pure YAML (no code change):
//
//   - conjugation: blank a verb; distractors = other persons of the same lemma.
//   - agreement:   blank an agreeing word (article/demonstrative/possessive/…);
//     key chosen from keyByGender; distractors = the other options.
//   - lookup:      blank a slot with a fixed correct answer per row (e.g. the
//     preposition before a given country); distractors = other options.
//
// Distractors come from the lexicon — never an LLM — so every item is correct by
// construction.
package generate

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"frpg-backend/internal/domain"
)

// createdAt is fixed so generation is fully deterministic (same inputs ⇒ same
// items ⇒ same ids ⇒ idempotent import).
const createdAt = "2026-07-10T00:00:00Z"

// builder kinds (GrammarPoint.Kind).
const (
	kindConjugation = "conjugation"
	kindAgreement   = "agreement"
	kindLookup      = "lookup"
)

// LevelPack is a level's grammar/skill taxonomy + lexicon, loaded from YAML.
type LevelPack struct {
	Code    string         `yaml:"code"`   // CEFR level, e.g. "A1"
	Skill   string         `yaml:"skill"`  // delivery modality for the pack
	Format  string         `yaml:"format"` // answer shape for the pack
	Teaches []GrammarPoint `yaml:"teaches"`
}

// GrammarPoint is one skill contrast the level tests, plus the data to build it.
// Which sub-fields are populated depends on Kind.
type GrammarPoint struct {
	ID            string   `yaml:"id"`
	Kind          string   `yaml:"kind"`       // conjugation | agreement | lookup
	SkillPoint    string   `yaml:"skillPoint"` // ties to Exercise.Contrast.SkillPoint
	Feature       string   `yaml:"feature"`    // the axis distractors vary on
	Instructions  string   `yaml:"instructions"`
	PromptVersion string   `yaml:"promptVersion"`
	Template      string   `yaml:"template"`  // single template (conjugation)
	Templates     []string `yaml:"templates"` // multiple templates (agreement/lookup)
	Options       []string `yaml:"options"`   // fixed option set (agreement/lookup)

	// conjugation
	Persons []Person `yaml:"persons"`
	Verbs   []Verb   `yaml:"verbs"`

	// agreement
	KeyByGender map[string]string `yaml:"keyByGender"` // gender -> correct option
	Nouns       []Noun            `yaml:"nouns"`

	// lookup
	Rows []Row `yaml:"rows"`
}

type Person struct {
	Subject string `yaml:"subject"`
	Index   int    `yaml:"index"` // index into Verb.Forms (je=0 … ils=5)
}

type Verb struct {
	Lemma      string   `yaml:"lemma"`
	Complement string   `yaml:"complement"`
	Forms      []string `yaml:"forms"` // je,tu,il,nous,vous,ils
}

type Noun struct {
	Word   string `yaml:"word"`
	Gender string `yaml:"gender"` // "m" | "f"
}

type Row struct {
	Filler string `yaml:"filler"` // fills the template slot (e.g. a country name)
	Key    string `yaml:"key"`    // the correct option for this row
}

// LoadPack reads and parses a level pack YAML file.
func LoadPack(path string) (*LevelPack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p LevelPack
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &p, nil
}

// Generate returns n exercises spread as uniformly as possible across the pack's
// grammar points (n/len each, remainder to the earlier ones). Order and content
// are deterministic; n larger than a pool is capped.
func Generate(pack *LevelPack, n int) []domain.Exercise {
	pools := make([][]domain.Exercise, len(pack.Teaches))
	for i, gp := range pack.Teaches {
		pools[i] = gp.build(pack)
	}
	k := len(pools)
	if k == 0 {
		return nil
	}
	out := make([]domain.Exercise, 0, n)
	for i, pool := range pools {
		cnt := n / k
		if i < n%k {
			cnt++
		}
		if cnt > len(pool) {
			cnt = len(pool)
		}
		out = append(out, pool[:cnt]...)
	}
	return out
}

func (gp GrammarPoint) build(pack *LevelPack) []domain.Exercise {
	switch gp.Kind {
	case kindConjugation:
		return gp.buildConjugation(pack)
	case kindAgreement:
		return gp.buildAgreement(pack)
	case kindLookup:
		return gp.buildLookup(pack)
	default:
		return nil
	}
}

func (gp GrammarPoint) buildConjugation(pack *LevelPack) []domain.Exercise {
	var out []domain.Exercise
	// persons outer, verbs inner: a capped sample spans the whole verb list rather
	// than exhausting one verb first.
	for _, p := range gp.Persons {
		for _, v := range gp.Verbs {
			if p.Index < 0 || p.Index >= len(v.Forms) {
				continue
			}
			key := v.Forms[p.Index]
			// distractors: other persons' forms, stable order, deduped.
			options := []string{key}
			for i := 1; i < len(v.Forms) && len(options) < 4; i++ { // skip index 0 (je)
				if i == p.Index {
					continue
				}
				options = appendUnique(options, v.Forms[i])
			}
			stem := strings.NewReplacer("{subject}", p.Subject, "{complement}", v.Complement).Replace(gp.Template)
			out = append(out, gp.makeItem(pack, v.Lemma, stem, options, key))
		}
	}
	return out
}

func (gp GrammarPoint) buildAgreement(pack *LevelPack) []domain.Exercise {
	var out []domain.Exercise
	// templates outer, nouns inner: a capped sample spans the whole noun list.
	for _, tmpl := range gp.Templates {
		for _, nn := range gp.Nouns {
			key, ok := gp.KeyByGender[nn.Gender]
			if !ok {
				continue
			}
			stem := strings.ReplaceAll(tmpl, "{word}", nn.Word)
			out = append(out, gp.makeItem(pack, nn.Word, stem, gp.Options, key))
		}
	}
	return out
}

func (gp GrammarPoint) buildLookup(pack *LevelPack) []domain.Exercise {
	var out []domain.Exercise
	// templates outer, rows inner: a capped sample spans the whole row list.
	for _, tmpl := range gp.Templates {
		for _, r := range gp.Rows {
			stem := strings.ReplaceAll(tmpl, "{filler}", r.Filler)
			out = append(out, gp.makeItem(pack, r.Filler, stem, gp.Options, r.Key))
		}
	}
	return out
}

// makeItem assembles a multiple_choice item. Options are sorted so the key's
// position varies, choice ids are a, b, c…, and the answer records the key's id
// (never its text — Content stays free of the key).
func (gp GrammarPoint) makeItem(pack *LevelPack, lemma, stem string, options []string, keyText string) domain.Exercise {
	opts := dedupSorted(options)
	choices := make([]map[string]any, len(opts))
	keyID := ""
	for i, text := range opts {
		id := string(rune('a' + i))
		choices[i] = map[string]any{"id": id, "text": text}
		if text == keyText {
			keyID = id
		}
	}

	return domain.Exercise{
		ID:     exerciseID(pack.Code, gp, lemma, stem, keyText),
		Skill:  pack.Skill,
		Format: pack.Format,
		Level:  pack.Code,
		Contrast: domain.TargetContrast{
			SkillPoint: gp.SkillPoint,
			Lemma:      lemma,
			Feature:    gp.Feature,
		},
		Prompt: domain.Prompt{
			Instructions: gp.Instructions,
			Text:         stem,
		},
		Content: map[string]any{
			"choices":  choices,
			"multiple": false,
		},
		Answer: map[string]any{
			"correct": []string{keyID},
		},
		Source: domain.SourceGenerated,
		Origin: domain.Origin{
			PromptVersion: gp.PromptVersion,
			RetrievedRefs: []string{},
			Reviewed:      false,
			CreatedBy:     "gen-cloze",
		},
		CreatedAt: createdAt,
	}
}

// exerciseID is a deterministic hash of the item's identifying inputs, so a
// re-run produces the same id and import is idempotent. The grammar point id is
// the human-readable prefix.
func exerciseID(code string, gp GrammarPoint, lemma, stem, keyText string) string {
	canonical := strings.Join([]string{code, gp.SkillPoint, lemma, gp.Feature, stem, keyText}, "|")
	sum := sha1.Sum([]byte(canonical))
	return "ex_" + strings.ToLower(code) + "_" + gp.ID + "_" + hex.EncodeToString(sum[:])[:8]
}

func appendUnique(xs []string, x string) []string {
	for _, e := range xs {
		if e == x {
			return xs
		}
	}
	return append(xs, x)
}

func dedupSorted(xs []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}
