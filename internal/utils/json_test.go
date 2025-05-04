package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteJSONError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		expectedBody   string
		expectedStatus int
		encodeErr      bool
	}{
		{
			name:           "success",
			status:         http.StatusBadRequest,
			message:        "bad request",
			expectedBody:   `{"error":"bad request"}`,
			expectedStatus: http.StatusBadRequest,
			encodeErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := WriteJSONError(w, tt.status, tt.message)

			assert.NoError(t, err, "WriteJSONError should not return an error")
			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type header mismatch")
			assert.JSONEq(t, tt.expectedBody, w.Body.String(), "Response body mismatch")
		})
	}
}
