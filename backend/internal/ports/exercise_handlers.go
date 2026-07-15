package ports

import (
	"errors"
	"net/http"

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
// only whether it was correct.
func (s *Server) handleGradeExercise(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExerciseID string   `json:"exerciseId"`
		Selected   []string `json:"selected"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	ok, err := s.Exercises.Grade(r.Context(), req.ExerciseID, req.Selected)
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
