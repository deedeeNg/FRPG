// Package generate produces French exercises from a YAML level pack (grammar/skill
// taxonomy + curated lexicon). It is the v0 stand-in for the future Python pipeline
// (pull labeled corpus → parse → blank → lexicon distractors → template); it emits
// the same domain.Exercise shape, so swapping in the real pipeline changes nothing
// downstream.
//
// The *config and data* live in YAML (see content/levels/a1.yaml) — the source of
// truth for what a level teaches; this file holds only the *assembly procedure*.
// Each grammar point declares a `kind` selecting one of six data-driven builders,
// so adding a new skill pack is usually pure YAML (no code change):
//
//   - conjugation:        blank a verb; distractors = other persons of the same lemma.
//   - agreement:          blank an agreeing word (article/demonstrative/possessive/…);
//     key chosen from keyByGender; distractors = the other options.
//   - lookup:              blank a slot with a fixed correct answer per row (e.g. the
//     preposition before a given country); distractors = other options.
//   - adjective:           agree an adjective with a noun's gender.
//   - passe_compose_etre:  agree a past participle with a Person's gender+number
//     (the auxiliary "être" is shared; only the participle varies).
//   - clock:               ask for the French time phrase for a given clock face.
//
// Distractors come from the lexicon — never an LLM — so every item is correct by
// construction. Any grammar point can also emit a fill_blank sibling of each item
// (same stem/key, typed answer instead of choices) by setting `alsoFillBlank: true`.
package generate

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"frpg-backend/internal/domain"
)

// createdAt is fixed so generation is fully deterministic (same inputs ⇒ same
// items ⇒ same ids ⇒ idempotent import).
const createdAt = "2026-07-10T00:00:00Z"

// builder kinds (GrammarPoint.Kind).
const (
	kindConjugation      = "conjugation"
	kindAgreement        = "agreement"
	kindLookup           = "lookup"
	kindNumber           = "number"
	kindAdjective        = "adjective"
	kindPasseComposeEtre = "passe_compose_etre"
	kindClock            = "clock"
)

// LevelPack is a level's grammar/skill taxonomy + lexicon, loaded from YAML.
type LevelPack struct {
	Code    string         `yaml:"code"`   // CEFR level, e.g. "A1"
	Skill   string         `yaml:"skill"`  // delivery modality for the pack
	Format  string         `yaml:"format"` // default answer shape for the pack
	Teaches []GrammarPoint `yaml:"teaches"`
}

// GrammarPoint is one skill contrast the level tests, plus the data to build it.
// Which sub-fields are populated depends on Kind.
type GrammarPoint struct {
	ID            string   `yaml:"id"`
	Kind          string   `yaml:"kind"`
	SkillPoint    string   `yaml:"skillPoint"` // ties to Exercise.Contrast.SkillPoint
	Feature       string   `yaml:"feature"`    // the axis distractors vary on
	Instructions  string   `yaml:"instructions"`
	PromptVersion string   `yaml:"promptVersion"`
	Format        string   `yaml:"format"`        // overrides pack.Format when set
	AlsoFillBlank bool     `yaml:"alsoFillBlank"` // also emit a fill_blank sibling of every item
	Template      string   `yaml:"template"`      // single template (conjugation/passe_compose_etre)
	Templates     []string `yaml:"templates"`     // multiple templates (agreement/lookup)
	Options       []string `yaml:"options"`       // fixed option set (agreement/lookup)

	// conjugation / passe_compose_etre
	Persons []Person `yaml:"persons"`
	Verbs   []Verb   `yaml:"verbs"`
	Aux     []string `yaml:"aux"` // shared auxiliary conjugation (je..ils), passe_compose_etre only

	// agreement
	KeyByGender map[string]string `yaml:"keyByGender"` // gender -> correct option
	Nouns       []Noun            `yaml:"nouns"`

	// adjective: agree an adjective with each noun's gender.
	Adjectives []Adjective `yaml:"adjectives"`

	// lookup: each row is a vocab record (e.g. {name, article, prep}); FillerField
	// names the column that fills the template, KeyField names the correct answer.
	// This lets one shared vocab list drive several skills (e.g. countries feed both
	// the article and the preposition exercises).
	Rows        []map[string]string `yaml:"rows"`
	FillerField string              `yaml:"fillerField"`
	KeyField    string              `yaml:"keyField"`

	// number: mint the words for every integer in [Min, Max] and quiz them.
	Min int `yaml:"min"`
	Max int `yaml:"max"`

	// clock: no extra fields — every hour:minute (0-23 × 0-59) in both registers
	// is minted algorithmically (see frenchTime), same spirit as kindNumber.
}

// Person is one subject a conjugation-kind builder drills. Gender/Plural are only
// read by kindPasseComposeEtre (which participle form to append); every other
// conjugation-kind builder ignores them.
type Person struct {
	Subject string `yaml:"subject"`
	Index   int    `yaml:"index"`  // index into Verb.Forms / GrammarPoint.Aux (je=0 … ils=5)
	Gender  string `yaml:"gender"` // "m" | "f" — passe_compose_etre only
	Plural  bool   `yaml:"plural"` // passe_compose_etre only
}

type Verb struct {
	Lemma      string      `yaml:"lemma"`
	Complement string      `yaml:"complement"`
	Forms      []string    `yaml:"forms"`                // je,tu,il,nous,vous,ils — used by kindConjugation
	Participle *Participle `yaml:"participle,omitempty"` // used by kindPasseComposeEtre instead of Forms
}

// Participle holds a past participle's four agreement forms — same shape as
// Adjective, because a passé-composé-with-être participle agrees exactly like an
// adjective (e.g. "allé/allée/allés/allées").
type Participle struct {
	M   string `yaml:"m"`
	F   string `yaml:"f"`
	Mpl string `yaml:"mpl"`
	Fpl string `yaml:"fpl"`
}

type Noun struct {
	Word   string `yaml:"word"`
	Gender string `yaml:"gender"` // "m" | "f"
}

// Adjective holds the four agreement forms (some may coincide, e.g. "français").
type Adjective struct {
	Lemma string `yaml:"lemma"`
	M     string `yaml:"m"`
	F     string `yaml:"f"`
	Mpl   string `yaml:"mpl"`
	Fpl   string `yaml:"fpl"`
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
	case kindNumber:
		return gp.buildNumber(pack)
	case kindAdjective:
		return gp.buildAdjective(pack)
	case kindPasseComposeEtre:
		return gp.buildPasseComposeEtre(pack)
	case kindClock:
		return gp.buildClock(pack)
	default:
		return nil
	}
}

// buildAdjective agrees each adjective with each noun's gender; the key is the
// singular form for that gender, distractors are the adjective's other forms. The
// template exposes {article} (le/la, from the noun's gender) and {word}.
func (gp GrammarPoint) buildAdjective(pack *LevelPack) []domain.Exercise {
	var out []domain.Exercise
	for _, adj := range gp.Adjectives {
		for _, nn := range gp.Nouns {
			key := adj.M
			article := "le"
			if nn.Gender == "f" {
				key, article = adj.F, "la"
			}
			var options []string
			for _, f := range []string{adj.M, adj.F, adj.Mpl, adj.Fpl} {
				if f != "" {
					options = appendUnique(options, f)
				}
			}
			stem := strings.NewReplacer("{article}", article, "{word}", nn.Word).Replace(gp.Template)
			out = append(out, gp.makeItems(pack, adj.Lemma, stem, options, key)...)
		}
	}
	return out
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
			out = append(out, gp.makeItems(pack, v.Lemma, stem, options, key)...)
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
			out = append(out, gp.makeItems(pack, nn.Word, stem, gp.Options, key)...)
		}
	}
	return out
}

func (gp GrammarPoint) buildLookup(pack *LevelPack) []domain.Exercise {
	var out []domain.Exercise
	// templates outer, rows inner: a capped sample spans the whole row list.
	for _, tmpl := range gp.Templates {
		for _, r := range gp.Rows {
			filler, key := r[gp.FillerField], r[gp.KeyField]
			if filler == "" || key == "" {
				continue
			}
			stem := strings.ReplaceAll(tmpl, "{filler}", filler)
			out = append(out, gp.makeItems(pack, filler, stem, gp.Options, key)...)
		}
	}
	return out
}

// buildPasseComposeEtre agrees a past participle with each Person's gender+number
// (e.g. "Il est allé" / "Elle est allée" / "Ils sont allés"). The auxiliary "être"
// is shared across all verbs (GrammarPoint.Aux) since it never varies; only the
// participle changes per verb and per person. Distractors are the SAME verb's
// other three agreement forms (wrong gender/number) with the SAME auxiliary — the
// point being tested is agreement, not person/aux selection (that's covered by the
// plain conjugation entries).
func (gp GrammarPoint) buildPasseComposeEtre(pack *LevelPack) []domain.Exercise {
	var out []domain.Exercise
	for _, p := range gp.Persons {
		if p.Index < 0 || p.Index >= len(gp.Aux) {
			continue
		}
		aux := gp.Aux[p.Index]
		for _, v := range gp.Verbs {
			if v.Participle == nil {
				continue
			}
			key := aux + " " + participleForm(v.Participle, p.Gender, p.Plural)
			options := []string{
				aux + " " + v.Participle.M,
				aux + " " + v.Participle.F,
				aux + " " + v.Participle.Mpl,
				aux + " " + v.Participle.Fpl,
			}
			stem := strings.NewReplacer("{subject}", p.Subject, "{complement}", v.Complement).Replace(gp.Template)
			out = append(out, gp.makeItems(pack, v.Lemma, stem, options, key)...)
		}
	}
	return out
}

func participleForm(p *Participle, gender string, plural bool) string {
	switch {
	case gender == "f" && plural:
		return p.Fpl
	case gender == "f":
		return p.F
	case plural:
		return p.Mpl
	default:
		return p.M
	}
}

// buildNumber mints the French words for every integer in [Min, Max] and asks the
// learner to pick the correct spelling; distractors are nearby numbers. Items are
// emitted in a strided order so a capped sample spreads across the whole range.
func (gp GrammarPoint) buildNumber(pack *LevelPack) []domain.Exercise {
	lo, hi := gp.Min, gp.Max
	if lo < 1 {
		lo = 1
	}
	if hi < lo {
		return nil
	}
	const stride = 8
	var out []domain.Exercise
	for off := 0; off < stride; off++ {
		for num := lo + off; num <= hi; num += stride {
			key := frenchNumber(num)
			if key == "" {
				continue
			}
			// distractors: nearby numbers, distinct words, in range.
			options := []string{key}
			for _, d := range []int{num - 1, num + 1, num + 10, num - 10, num + 2, num - 2} {
				if len(options) >= 4 {
					break
				}
				if d < lo || d > hi {
					continue
				}
				if w := frenchNumber(d); w != "" && w != key {
					options = appendUnique(options, w)
				}
			}
			stem := strconv.Itoa(num)
			out = append(out, gp.makeItems(pack, stem, stem, options, key)...)
		}
	}
	return out
}

// numberUnits holds 0..16, from which every French number to 80 is composed.
var numberUnits = []string{
	"zéro", "un", "deux", "trois", "quatre", "cinq", "six", "sept", "huit", "neuf",
	"dix", "onze", "douze", "treize", "quatorze", "quinze", "seize",
}

// frenchNumber spells an integer in [0, 80] in French, "" if out of range. Handles
// the "et un" cases (21, 31…61, 71) and the vigesimal 70s/80 (soixante-dix, quatre-vingts).
func frenchNumber(n int) string {
	switch {
	case n < 0 || n > 80:
		return ""
	case n <= 16:
		return numberUnits[n]
	case n <= 19: // 17..19 = dix-sept…dix-neuf
		return "dix-" + numberUnits[n-10]
	case n <= 69:
		tens := map[int]string{2: "vingt", 3: "trente", 4: "quarante", 5: "cinquante", 6: "soixante"}[n/10]
		switch u := n % 10; {
		case u == 0:
			return tens
		case u == 1:
			return tens + " et un"
		default:
			return tens + "-" + numberUnits[u]
		}
	case n <= 79: // 70..79 = soixante + 10..19
		r := n - 60
		switch {
		case r == 11:
			return "soixante et onze"
		case r <= 16:
			return "soixante-" + numberUnits[r]
		default:
			return "soixante-dix-" + numberUnits[r-10]
		}
	default: // 80
		return "quatre-vingts"
	}
}

// buildClock mints EVERY hour:minute (0-23 × 0-59) in BOTH French time
// registers — l'heure officielle (24h: "vingt-trois heures trente-six") and
// everyday 12h speech with a time-of-day qualifier ("onze heures trente-six du
// soir") — via frenchTime, the same "algorithm, not a curated list" approach as
// buildNumber. Distractors are nearby times (±15/30/45 min, wrapping) in the SAME
// register, so they're plausible near-miss readings of the same clock face.
//
// Emission order interleaves register and a minute-stride (informal times first,
// minutes spread via a stride coprime-ish with 60) so a capped sample still spans
// a wide variety of hours/minutes/registers rather than exhausting 00:00-00:59
// before ever reaching another hour.
// buildClock emits in TWO phases so a capped sample — not just the full
// 24×60×2 pool — reliably shows both the special phrases and arbitrary times:
//
//   - Phase 1: the four special minute markers (:00/:15/:30/:45 — et quart, et
//     demie, moins le quart, midi/minuit) across every hour, hour order shuffled
//     via a stride so matin/après-midi/soir all appear within the first few
//     hours rather than only after exhausting 0-11.
//   - Phase 2: every OTHER ("arbitrary") minute, via a coprime modular walk over
//     the full day — proves "any time" coverage (e.g. 23:36) once the cap grows
//     past phase 1.
//
// formal/informal are interleaved per minute (not split into separate halves),
// so even a small capped sample sees both registers together.
func (gp GrammarPoint) buildClock(pack *LevelPack) []domain.Exercise {
	const totalMinutes = 24 * 60
	var out []domain.Exercise
	seen := map[int]bool{} // minute-of-day already emitted

	emit := func(hour, minute int) {
		t := hour*60 + minute
		if seen[t] {
			return
		}
		seen[t] = true
		for _, formal := range []bool{false, true} {
			phrase := frenchTime(hour, minute, formal)
			options := []string{phrase}
			for _, delta := range []int{15, -15, 30, -30, 45, -45} {
				if len(options) >= 4 {
					break
				}
				dh, dm := addMinutes(hour, minute, delta)
				if w := frenchTime(dh, dm, formal); w != phrase {
					options = appendUnique(options, w)
				}
			}
			lemma := fmt.Sprintf("%02d:%02d", hour, minute)
			// stem is empty: the stimulus is the rendered clock face
			// (Content.clock), not text — Prompt.Text would otherwise just
			// duplicate Instructions.
			items := gp.makeItems(pack, lemma, "", options, phrase)
			for j := range items {
				items[j].Content["clock"] = map[string]any{"hour": hour, "minute": minute, "formal": formal}
			}
			out = append(out, items...)
		}
	}

	// Phase 1: special minutes, hours shuffled (stride 5, coprime with 24) so
	// morning/afternoon/evening all show up early.
	hour := 0
	for i := 0; i < 24; i++ {
		for _, minute := range []int{0, 15, 30, 45} {
			emit(hour, minute)
		}
		hour = (hour + 5) % 24
	}

	// Phase 2: every remaining minute, via a coprime (with 1440) modular walk —
	// 37 visits every minute-of-day exactly once before repeating.
	t := 0
	for i := 0; i < totalMinutes; i++ {
		emit(t/60, t%60)
		t = (t + 37) % totalMinutes
	}

	return out
}

// addMinutes wraps hour:minute by delta minutes across a 24h day.
func addMinutes(hour, minute, delta int) (int, int) {
	total := ((hour*60+minute+delta)%1440 + 1440) % 1440
	return total / 60, total % 60
}

// frenchTime spells hour:minute (24h input) in French. formal=true gives l'heure
// officielle (always "H heures MM", literal digits, no "et quart"/"moins");
// formal=false gives everyday 12h speech (et quart/et demie/moins le quart for
// :15/:30/:45, midi/minuit by name, a du matin/de l'après-midi/du soir qualifier
// to disambiguate the 12h wrap, and every OTHER minute stated directly — that's
// the A1-taught rule: only 0/15/30/45 get special phrasing, e.g. 23:36 is "onze
// heures trente-six du soir", not some "moins vingt-quatre" construction).
func frenchTime(hour, minute int, formal bool) string {
	hour = ((hour % 24) + 24) % 24
	minute = ((minute % 60) + 60) % 60
	if formal {
		return frenchTimeFormal(hour, minute)
	}
	return frenchTimeInformal(hour, minute)
}

func frenchTimeFormal(hour, minute int) string {
	unit := "heures"
	if hour == 0 || hour == 1 {
		unit = "heure"
	}
	label := frenchHourWord(hour) + " " + unit
	if minute == 0 {
		return label
	}
	return label + " " + frenchNumber(minute)
}

func frenchTimeInformal(hour, minute int) string {
	// midi/minuit take the minute directly ("minuit vingt"), not "zéro heure
	// vingt" / "douze heures vingt" — handle both reference hours up front.
	if hour == 0 || hour == 12 {
		base := "minuit"
		if hour == 12 {
			base = "midi"
		}
		switch minute {
		case 0:
			return base
		case 15:
			return base + " et quart"
		case 30:
			return base + " et demi"
		case 45:
			nextHour := (hour + 1) % 24
			nih, nq := informalHourAndQualifier(nextHour)
			return frenchHourPhrase(nih) + " moins le quart " + nq
		default:
			return base + " " + frenchNumber(minute)
		}
	}

	ihour, qualifier := informalHourAndQualifier(hour)
	switch minute {
	case 0:
		return frenchHourPhrase(ihour) + " " + qualifier
	case 15:
		return frenchHourPhrase(ihour) + " et quart " + qualifier
	case 30:
		return frenchHourPhrase(ihour) + " et demie " + qualifier
	case 45:
		nextHour := (hour + 1) % 24
		if nextHour == 0 {
			return "minuit moins le quart"
		}
		if nextHour == 12 {
			return "midi moins le quart"
		}
		nih, nq := informalHourAndQualifier(nextHour)
		return frenchHourPhrase(nih) + " moins le quart " + nq
	default:
		return frenchHourPhrase(ihour) + " " + frenchNumber(minute) + " " + qualifier
	}
}

// informalHourAndQualifier maps a 24h hour (never 0 or 12 — callers special-case
// those as minuit/midi) to its 12h form and time-of-day qualifier.
func informalHourAndQualifier(h int) (int, string) {
	switch {
	case h < 12:
		return h, "du matin"
	case h < 18:
		return h - 12, "de l'après-midi"
	default:
		return h - 12, "du soir"
	}
}

// frenchHourPhrase renders a 1-12 hour count with the correctly agreed "heure(s)".
func frenchHourPhrase(ihour int) string {
	if ihour == 1 {
		return "une heure"
	}
	return frenchHourWord(ihour) + " heures"
}

// frenchHourWord is frenchNumber with "un" feminized to "une" ("heure" is
// feminine: "une heure", "vingt et une heures" — never "un heure(s)").
func frenchHourWord(n int) string {
	w := frenchNumber(n)
	if w == "un" {
		return "une"
	}
	return strings.Replace(w, " un", " une", 1)
}

// makeItems assembles one exercise, format(s) per the grammar point's Format
// (falling back to the pack default). When AlsoFillBlank is set, it returns BOTH
// a multiple_choice item and a fill_blank sibling built from the same stem/key —
// same underlying question, two answer shapes, so no data is duplicated to get both.
func (gp GrammarPoint) makeItems(pack *LevelPack, lemma, stem string, options []string, keyText string) []domain.Exercise {
	format := gp.Format
	if format == "" {
		format = pack.Format
	}
	items := []domain.Exercise{gp.makeItem(pack, lemma, stem, options, keyText, format)}
	if gp.AlsoFillBlank && format != domain.FormatFillBlank {
		items = append(items, gp.makeItem(pack, lemma, stem, options, keyText, domain.FormatFillBlank))
	}
	return items
}

// makeItem assembles one exercise in the given format.
//
//   - multiple_choice: options are sorted so the key's position varies, choice ids
//     are a, b, c…, and the answer records the key's id (never its text — Content
//     stays free of the key).
//   - fill_blank: no choices/distractors needed — Content carries the blanked
//     template, Answer carries the accepted typed text(s). Prompt.Text stays empty
//     (the blanked sentence lives in Content.template instead), matching the
//     listening fill_blank shape already documented in PLAN.md.
func (gp GrammarPoint) makeItem(pack *LevelPack, lemma, stem string, options []string, keyText, format string) domain.Exercise {
	e := domain.Exercise{
		ID:     exerciseID(pack.Code, gp, lemma, stem, keyText, format),
		Skill:  pack.Skill,
		Format: format,
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
		Source: domain.SourceGenerated,
		Origin: domain.Origin{
			PromptVersion: gp.PromptVersion,
			RetrievedRefs: []string{},
			Reviewed:      false,
			CreatedBy:     "gen-cloze",
		},
		CreatedAt: createdAt,
	}

	if format == domain.FormatFillBlank {
		e.Prompt.Text = ""
		e.Content = map[string]any{
			"template": stem,
			"blanks":   []map[string]any{{"id": "1"}},
		}
		e.Answer = map[string]any{
			"accepted": map[string][]string{"1": {keyText}},
		}
		return e
	}

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
	e.Content = map[string]any{"choices": choices, "multiple": false}
	e.Answer = map[string]any{"correct": []string{keyID}}
	return e
}

// exerciseID is a deterministic hash of the item's identifying inputs, so a
// re-run produces the same id and import is idempotent. Format is part of the
// hash so a multiple_choice item and its fill_blank sibling get distinct ids even
// though they share the same stem/key. The grammar point id is the human-readable
// prefix.
func exerciseID(code string, gp GrammarPoint, lemma, stem, keyText, format string) string {
	canonical := strings.Join([]string{code, gp.SkillPoint, lemma, gp.Feature, stem, keyText, format}, "|")
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
