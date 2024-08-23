package send

import (
	"encoding/json"
	"fmt"
	"github.com/anmho/idempotent-rides/scope"
	"log/slog"
	"net/http"
)

type HTTPError struct {
	Cause   error  `json:"error,omitempty"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

var _ error = (*HTTPError)(nil)

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTPError status %d: %s caused by %s", e.Status, e.Message, e.Cause.Error())
}

func NewErrInternal(cause error) HTTPError {
	return HTTPError{
		Cause:   cause,
		Message: "internal server error",
		Status:  http.StatusInternalServerError,
	}
}

func Error(
	w http.ResponseWriter,
	err HTTPError) {
	// Log the original error
	scope.GetLogger().Error(
		err.Message,
		slog.Int("status", err.Status),
		slog.Any("cause", err.Cause),
	)

	// Strip the original error from the payload
	// before sending back to client

	err.Cause = nil
	// Need to set status code before writing, or it will be marked as superfluous
	if err.Status == 0 {
		err.Status = http.StatusInternalServerError
	}
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)

}
