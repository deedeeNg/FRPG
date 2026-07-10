# Exercise Generation — Strategy

Companion to `PLAN.md` (phased build) and `PROPOSE.md` (options explored). This file
records the **decided** generation strategy so we don't re-litigate it.

Guiding rule: **keep it as simple as possible.** Ship the deterministic path first;
add AI only where it's actually needed.

---

## v0 in one line

Pull **already-labeled** French sentences from a public dataset, blank a word tied to a
target skill, build the wrong choices **deterministically** from a morphological
lexicon, fill a template, and import the finished items into the DB.
**No LLM. No CEFR classifier.**

Why no classifier: the dataset sentences already carry a CEFR label, so the level is
*inherited from the source* — there is nothing to classify. (The classifier only earns
its place once we generate *new* text with an LLM — see [Reminder](#reminder--parked-until-we-add-llm-generation).)

---

## Pipeline (offline, Python)

1. **Pull** `vekkt/french_CEFR` → keep each sentence + its level label; filter to A1/A2.
2. **Parse** (spaCy `fr_core_news_sm`) → POS + lemma; find the token for the target
   skill (a conjugated form of `parler`, the article before a noun, …).
3. **Blank** that token → the stem.
4. **Choices** from the **lexicon** (Verbiste / Lefff / UniMorph): the key + distractors
   are other inflections of the *same lemma* differing only on the target `Feature`.
5. **Fill** a template → a finished exercise.
6. **Write** one JSON object per line to `exercises.jsonl`.

Then, on the Go side: `cmd/import-exercises` reads the jsonl, unmarshals each line into
`domain.Exercise` (schema gate), and calls `ExerciseStore.Put` → DynamoDB.

```
french_CEFR ─┐
             ▼   (Python, offline)                         (Go)
  parse → blank → lexicon choices → template → exercises.jsonl → import → DynamoDB
```

---

## Locked decisions

- **Handoff:** Python → `exercises.jsonl` → Go importer → DynamoDB. **Go is the single
  schema authority** (the `Exercise` struct + its tags are the one contract).
- **IDs:** deterministic hash of the identifying inputs (template + lemma + level +
  skill + blanked form) → re-runs are **idempotent**, so `Put` is a plain upsert (no
  duplicate bank).
- **Distractors:** deterministic **lexicon** for grammar skills — **no LLM**.
- **LevelPack** (YAML, source of truth) has two axes:
  - `teaches` — the SkillPoints a level may *test* as an item's `contrast`. **Distinct
    per level** (A2 does not re-teach A1).
  - `assumes` — vocab + grammar a level may *use / assume as known*. **Cumulative**
    (A2 assumes A1: e.g. an A2 item writes *"la maison"* and assumes you know its
    gender, without gender being the tested point).
- **Validation — two gates, no overlap:** Python validates **pedagogy** (skill match,
  distractor sanity, level); the Go importer validates **schema only** and stays
  ignorant of French.
- **Repo layout:** `content/` (`levels/`, `templates/`, `lexicon/`) shared by both
  languages · `pipeline/` (Python) · `backend/` (Go).

---

## First batch (prove the loop)

reading · skills `article_gender` + `conjugation` · formats `multiple_choice` +
`fill_blank` · corpus + lexicon + template **only** (no audio, no LLM, no classifier).
This runs the whole loop end to end on the deterministic path before any AI is added.

---

## Example — one generated item (a line of `exercises.jsonl`)

A present-tense conjugation cloze, built from the A1 sentence *"Je parle français."*
(blank `parle`; distractors are other persons of `parler`):

```json
{
  "exerciseId": "ex_a1_conj_9f2c1a",
  "skill": "reading", "format": "multiple_choice", "level": "A1",
  "contrast": { "skillPoint": "conjugation", "lemma": "parler", "feature": "person" },
  "prompt": { "instructions": "Choisissez la bonne forme du verbe.",
              "text": "Je ___ français.", "audioUrl": "", "imageUrl": "" },
  "content": { "choices": [ {"id":"a","text":"parle"}, {"id":"b","text":"parles"},
                            {"id":"c","text":"parlons"}, {"id":"d","text":"parlez"} ],
               "multiple": false },
  "answer": { "correct": ["a"] },
  "source": "generated",
  "origin": { "model": "", "promptVersion": "tmpl-present-er-v1",
              "retrievedRefs": ["french_CEFR#12345"], "reviewed": false, "createdBy": "gen-cloze" },
  "createdAt": "2026-07-09T00:00:00Z"
}
```

Maps 1:1 onto `domain.Exercise` (see `internal/domain/exercise.go`). `answer` is stored
but **stripped before serving**; `origin.retrievedRefs` keeps the source sentence id for
audit.

---

## Reminder — parked until we add LLM generation

**Not part of v0.** Revisit only when we generate *new* text (dialogue, free-form)
rather than reuse labeled corpus:

- **HF CEFR classifier (Inference API)** — the level-validator gate. Needed because
  LLM-generated text has **no trusted label**. Pin a served French-CEFR model then;
  keep it offline (batch), never on the hot path.
- **LLM meaning items** — `vocab_meaning`, `verb_choice`, reading comprehension, and
  **`dialogue_sim`** (speaking) — "dialog or something entirely".
- **`match` format** (*associez*) — 47× in the Cahier; a natural early add after v0.
- **Listening / audio** — TTS + `Prompt.AudioURL`.
- **A2 content** — needs the *Édito A2* volume or the DELF A2 référentiel (Livre.pdf is
  A1 only).

---

## Before building the Python side (open checks)

- Verify `vekkt/french_CEFR`: **label granularity** (A1–C2, or coarser?) and **license**
  (may we redistribute derived items?).
- Pick the **lexicon** source (Verbiste vs Lefff vs UniMorph FR).
