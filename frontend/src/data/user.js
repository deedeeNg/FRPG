// Player profile shown in the HUD (avatar chip + XP bar). Seeded for now with
// the prototype's values; swap `userProfile` for data fetched from the backend
// (GET /api/me currently returns only { userId, email } — extend it with level
// and XP, then feed the response in here).
//
//   name         -> avatar label ("<name> · Niveau <level>")
//   level        -> current level
//   currentXp    -> XP earned toward the next level
//   nextLevelXp  -> XP needed to reach the next level (bar fill = current / next)
export const userProfile = {
  name: 'Adventurer',
  level: 5,
  currentXp: 1240,
  nextLevelXp: 2000,
}
