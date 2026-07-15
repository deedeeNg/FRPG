package domain

import (
	"context"
	"errors"
)

// ErrExerciseNotFound is returned by an ExerciseStore when no item matches an id.
var ErrExerciseNotFound = errors.New("exercise not found")

// Exercise is the single, format-agnostic learning item. One shape serves every
// skill (reading/listening/speaking/writing) and every answer format
// (multiple_choice/fill_blank/…) so adding a format is a new payload, never a new
// table or a schema migration.
//
// The design rests on three orthogonal axes kept in separate fields:
//
//   - Skill  = how the item is *delivered* (the stimulus modality).
//   - Format = how the answer is *collected & graded* (the discriminator for
//     Content/Answer).
//   - Contrast = the *pedagogical point* — what the item makes the learner think
//     about — which is what lets a generator/validator prove the question is
//     actually testing something (see TargetContrast).
//
// Prompt/Content/Answer follow the rule: Prompt = what you *perceive*, Content =
// what you *interact with*, Answer = the *secret*. Answer is graded server-side and
// MUST be stripped before an item is sent to the client (the ports layer does this;
// storage keeps it). Contrast, Prompt and Origin are value objects stored **inline**
// in this one item — they are not separate tables.
//
// Example stored item (DynamoDB JSON), an A2 article-gender reading MCQ:
//
//	{
//	  "exerciseId": "ex_a2_art_0007",
//	  "skill": "reading", "format": "multiple_choice", "level": "A2",
//	  "contrast": { "skillPoint": "article_gender", "lemma": "maison", "feature": "gender" },
//	  "prompt":  { "instructions": "Choisissez le bon article.", "text": "___ maison est grande.",
//	               "audioUrl": "", "imageUrl": "" },
//	  "content": { "choices": [ {"id":"a","text":"la"}, {"id":"b","text":"le"}, {"id":"c","text":"l'"} ],
//	               "multiple": false },
//	  "answer":  { "correct": ["a"] },
//	  "source":  "generated",
//	  "origin":  { "model": "", "promptVersion": "tmpl-articles-v1", "retrievedRefs": [],
//	               "reviewed": true, "createdBy": "gen-drills" },
//	  "createdAt": "2026-07-03T00:00:00Z"
//	}
//
// A *listening* variant of that same item differs only in Skill ("listening"),
// Contrast.SkillPoint ("listening_comprehension") and Prompt (empty Text, populated
// AudioURL) — Content/Answer/Format are identical. That diff is the whole
// flexibility argument.
type Exercise struct {
	// ID is the primary key, "ex_<...>" (e.g. "ex_a2_art_0007").
	ID string `dynamodbav:"exerciseId" json:"exerciseId"`

	// Skill is the delivery modality: one of the Skill* constants. Drives which
	// Prompt fields are meaningful (e.g. listening ⇒ AudioURL, not Text).
	Skill string `dynamodbav:"skill" json:"skill"`

	// Format is the answer shape and the discriminator for Content/Answer: one of
	// the Format* constants. A per-format Grader reads Answer to score.
	Format string `dynamodbav:"format" json:"format"`

	// Level is the CEFR level (Level* constants). It is a *contract*: every token in
	// the item must fit that level's grammar+vocab whitelist (enforced later by the
	// validator gauntlet, not here). It also selects which items a learner sees.
	Level string `dynamodbav:"level" json:"level"`

	// Contrast is the machine-checkable pedagogical intent — what makes this "a
	// question with a point" rather than "a question". Inline value object.
	Contrast TargetContrast `dynamodbav:"contrast" json:"contrast"`

	// Prompt is the stimulus the learner perceives. Inline value object; which of
	// its fields are filled *is* the modality.
	Prompt Prompt `dynamodbav:"prompt" json:"prompt"`

	// Content is the format-specific apparatus the learner manipulates (choices,
	// blank template…). Open map so a new Format needs no schema change. It is
	// client-safe and NEVER contains the key — the key lives in Answer.
	Content map[string]any `dynamodbav:"content" json:"content"`

	// Answer is the format-specific grading spec (correct ids, accepted strings…).
	// SERVER-ONLY: strip before sending an item to the client.
	Answer map[string]any `dynamodbav:"answer" json:"answer"`

	// Source is the coarse provenance bucket: one of the Source* constants.
	Source string `dynamodbav:"source" json:"source"`

	// Origin is fine-grained provenance (which model/template/refs produced it,
	// whether a human reviewed it). Inline value object; drives auditing and the
	// generation flywheel. Distinct from Source, which is only the bucket.
	Origin Origin `dynamodbav:"origin" json:"origin"`

	// CreatedAt is RFC3339.
	CreatedAt string `dynamodbav:"createdAt" json:"createdAt"`
}

// TargetContrast is what the item makes the learner think about — the field that
// turns "a question" into "a question with a checkable point". Without it, a
// fill-blank meant to test verb conjugation could offer a noun as a distractor and
// nothing would flag it. Stored inline in Exercise (not a table).
type TargetContrast struct {
	// SkillPoint is the pedagogical category (one of the SkillPoint* constants). It
	// selects which validator runs, which mastery bucket an attempt updates, and
	// which templates are eligible.
	SkillPoint string `dynamodbav:"skillPoint" json:"skillPoint"`

	// Lemma is the specific word under test — MAY be empty. Set for word-level
	// grammar drills so the morphology gate can assert "all choices are inflections
	// of this word"; empty for comprehension items (nothing single-word to check).
	Lemma string `dynamodbav:"lemma" json:"lemma"`

	// Feature is the axis the answer varies on (e.g. "person", "gender") — MAY be
	// empty. When set, distractors must differ from the key ONLY on this axis.
	Feature string `dynamodbav:"feature" json:"feature"`
}

// Prompt is the stimulus the learner perceives before answering. It is multi-field
// precisely so one struct serves every skill: which fields are filled is the
// modality. Load-bearing rule: a listening item sets AudioURL and leaves Text empty
// (showing the transcript would defeat listening). Stored inline in Exercise.
type Prompt struct {
	Instructions string `dynamodbav:"instructions" json:"instructions"` // the task line; always set
	Text         string `dynamodbav:"text" json:"text"`                 // written stimulus; reading/writing
	AudioURL     string `dynamodbav:"audioUrl" json:"audioUrl"`         // audio asset pointer; listening
	ImageURL     string `dynamodbav:"imageUrl" json:"imageUrl"`         // picture; picture-description tasks
}

// Origin is fine-grained provenance for one item: enough to audit it and to power
// the generation flywheel (regenerate everything from a bad prompt version, prefer
// human-reviewed items, etc.). Stored inline in Exercise.
type Origin struct {
	Model         string   `dynamodbav:"model" json:"model"`                 // generating model id; empty for template/human
	PromptVersion string   `dynamodbav:"promptVersion" json:"promptVersion"` // template or prompt version, e.g. "tmpl-articles-v1"
	RetrievedRefs []string `dynamodbav:"retrievedRefs" json:"retrievedRefs"` // source refs (corpus ids/urls) if retrieval-grounded
	Reviewed      bool     `dynamodbav:"reviewed" json:"reviewed"`           // a human approved this item
	CreatedBy     string   `dynamodbav:"createdBy" json:"createdBy"`         // pipeline/tool/user that produced it
}

// Skill values (delivery modality).
const (
	SkillReading   = "reading"
	SkillListening = "listening"
	SkillSpeaking  = "speaking"
	SkillWriting   = "writing"
)

// Format values (answer shape / Content+Answer discriminator).
const (
	FormatMultipleChoice = "multiple_choice"
	FormatFillBlank      = "fill_blank"
	FormatDialogueSim    = "dialogue_sim"
)

// Level values (CEFR). v0 targets A1/A2; B1/B2 defined for later.
const (
	LevelA1 = "A1"
	LevelA2 = "A2"
	LevelB1 = "B1"
	LevelB2 = "B2"
)

// SkillPoint values (TargetContrast.SkillPoint). Extend as new drills are added.
const (
	SkillPointConjugation            = "conjugation"
	SkillPointArticleGender          = "article_gender"
	SkillPointAgreement              = "agreement"
	SkillPointPreposition            = "preposition"
	SkillPointNegation               = "negation"
	SkillPointVerbChoice             = "verb_choice"
	SkillPointVocabMeaning           = "vocab_meaning"
	SkillPointReadingComprehension   = "reading_comprehension"
	SkillPointListeningComprehension = "listening_comprehension"
)

// Source values (coarse provenance bucket).
const (
	SourceGenerated = "generated" // produced by a template/model pipeline
	SourceAuthored  = "authored"  // written by a human
)

// ExerciseStore is the driven port for exercise storage: the write side lets a
// generator push items in, the read side serves them. Adapters implement it
// (InMemory for tests/dev, Dynamo for production), mirroring the User Repository.
type ExerciseStore interface {
	// Get returns one item by id, or ErrExerciseNotFound.
	Get(ctx context.Context, id string) (Exercise, error)

	// Query samples up to limit items for a (level, skill) pair — the "give me N
	// A2 reading items" access pattern, served from a GSI in the Dynamo adapter.
	Query(ctx context.Context, level, skill string, limit int) ([]Exercise, error)

	// Put stores (inserts or overwrites) one item — the push path a generator uses.
	Put(ctx context.Context, e Exercise) error
}
