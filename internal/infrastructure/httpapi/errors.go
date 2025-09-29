package httpapi

import (
    "encoding/json"
    "net/http"
)

type apiErrorBody struct {
    Error apiError `json:"error"`
}

type apiError struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

func writeError(w http.ResponseWriter, status int, code string, message string, details interface{}) {
    if code == "" { code = http.StatusText(status) }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(apiErrorBody{Error: apiError{Code: code, Message: message, Details: details}})
}

