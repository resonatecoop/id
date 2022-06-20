package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var realm = "go_oauth2_server"

// WriteJSON writes JSON response
func WriteJSON(w http.ResponseWriter, v interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// NoContent writes a 204 no content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error produces a JSON error response with the following structure:
// {"error":"some error message"}
func Error(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc_err := json.NewEncoder(w).Encode(map[string]string{"error": err})
	if enc_err != nil {
		http.Error(w, enc_err.Error(), http.StatusInternalServerError)
		return
	}
}

// UnauthorizedError has to contain WWW-Authenticate header
// See http://self-issued.info/docs/draft-ietf-oauth-v2-bearer.html#rfc.section.3
func UnauthorizedError(w http.ResponseWriter, err string) {
	// TODO - include error if the request contained an access token
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer realm=%s", realm))
	Error(w, err, http.StatusUnauthorized)
}
