package utils

import (
  "encoding/json"
  "log"
  "net/http"
)

func WriteJSONError(w http.ResponseWriter, status int, message string) error {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  err := json.NewEncoder(w).Encode(map[string]string{"error": message})
  if err != nil {
    log.Printf("Failed to encode error response: %v", err)
  }
  return err
}