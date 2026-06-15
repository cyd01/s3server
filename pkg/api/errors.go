package api

import (
	"encoding/xml"
	"net/http"
)

type ErrorResponse struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message"`
}

func WriteError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	_ = xml.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
	})
}
