package util

import (
	"encoding/json"
	"net/http"
)

// RespondWithError responds
func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"error": msg})
}

// RespondWithJSON responds
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

/**
  isJSONString("Platypus") = true
  isJSON("Platypus") = false

  isJSONString(Platypus) = false
  isJSON(Platypus) = false

  isJSONString({"id":"1"}) = false
  isJSON({"id":"1"}) = true
*/

// IsJSONString checks if string contains JSON
func IsJSONString(s string) bool {
	var js string
	return json.Unmarshal([]byte(s), &js) == nil

}

// IsJSON checks if object is JSON
func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
