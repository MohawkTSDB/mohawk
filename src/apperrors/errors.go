package apperrors

import (
	"fmt"
	"net/http"
)

var BadMetricIDErr = fmt.Errorf("Bad metrics ID")

// Error represent an error that occurred and its status code.
type Error struct {
	code    int
	message string
}

func (e Error) Error() string {
	return e.message
}

// JSON writes Error as JSON to http.ResponseWriter
func (e Error) JSON(w http.ResponseWriter) {
	w.WriteHeader(e.code)
	w.Write([]byte(fmt.Sprintf(`{"code":"%d","message":"%s"}`, e.code, e.message)))
}

// BadRequest creates new Error with status code 400 from error.
func BadRequest(err error) Error {
	return Error{code: http.StatusBadRequest, message: err.Error()}
}

// InternalError creates new Error with status code 500 from error.
func InternalError(err error) Error {
	return Error{code: http.StatusInternalServerError, message: err.Error()}
}
