// Character attribute stats shown in the home-page hexagon (radar) chart.
// The 6 axes are fixed; each is scored out of 100.
//
// For now each value is seeded with a random 0–100 number. These are held as
// plain variables so they can later be swapped for real per-user data pulled
// from the backend — replace `randomScore()` with the fetched value per key.

const MAX_SCORE = 100

function randomScore() {
  return Math.floor(Math.random() * (MAX_SCORE + 1))
}

// key -> i18n label key. Order defines the axis order around the hexagon.
export const attributes = [
  { key: 'listening', labelKey: 'attr.listening' },
  { key: 'reading', labelKey: 'attr.reading' },
  { key: 'writing', labelKey: 'attr.writing' },
  { key: 'speaking', labelKey: 'attr.speaking' },
  { key: 'vocabulary', labelKey: 'attr.vocabulary' },
  { key: 'grammar', labelKey: 'attr.grammar' },
]

export { MAX_SCORE }

// Seed values (0–100). Swap these for backend data later.
export const attributeScores = {
  listening: randomScore(),
  reading: randomScore(),
  writing: randomScore(),
  speaking: randomScore(),
  vocabulary: randomScore(),
  grammar: randomScore(),
}
