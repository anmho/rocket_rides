package send

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		Error(w, HTTPError{
			Cause:   err,
			Message: "internal server error",
			Status:  http.StatusInternalServerError,
		})
	}
}

func Read[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decoding json: %w", err)
	}
	return v, nil
}
