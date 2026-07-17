package ports

import (
	"errors"
	"net/http"

	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

// handleNextExercise returns one exercise to attempt, with the answer key
// stripped — grading happens server-side (handleGradeExercise), so the correct
// answer never reaches the client.
func (s *Server) handleNextExercise(w http.ResponseWriter, r *http.Request) {
	e, err := s.Exercises.Next(r.Context(), "A1", "reading")
	if errors.Is(err, domain.ErrExerciseNotFound) {
		writeError(w, http.StatusNotFound, "no exercises available")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load exercise")
		return
	}
	e.Answer = nil // never send the answer key to the client
	writeJSON(w, http.StatusOK, e)
}

// handleGradeExercise checks a submitted answer against the stored key and returns
// only whether it was correct. Selected (choice ids) is used for multiple_choice
// exercises, Text (typed answer) for fill_blank — the exercise's own Format
// decides which one app.Exercises.Grade reads; the client just sends whichever it
// collected.
func (s *Server) handleGradeExercise(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExerciseID string   `json:"exerciseId"`
		Selected   []string `json:"selected"`
		Text       string   `json:"text"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	sub := app.Submission{Selected: req.Selected, Text: req.Text}
	ok, err := s.Exercises.Grade(r.Context(), req.ExerciseID, sub)
	if errors.Is(err, domain.ErrExerciseNotFound) {
		writeError(w, http.StatusNotFound, "exercise not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not grade")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"correct": ok})
}
