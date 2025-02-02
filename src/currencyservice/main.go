// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
)

func main() {
	log.Print("starting server...")
	http.Handle("/convert", &convertHandler{currencies: currencies()})

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// ConvertRequest stores request data
type ConvertRequest struct {
	Value string `json:"value"`
}

// ConvertResponse stores response data
type ConvertResponse struct {
	Answer int `json:"answer"`
}

type convertHandler struct {
	currencies map[string]float64
}

func currencies() map[string]float64 {
	return map[string]float64{
		"JPY": 100,
		"USD": 0.82,
		"EUR": 0.74,
		"BRL": 3.96,
		"AUD": 1.09,
	}
}

// IsValid validates request data
func (s *ConvertRequest) IsValid() bool {
	if len(s.Value) < 4 {
		return false
	}
	if _, err := strconv.Atoi(s.Value[3:]); err != nil {
		return false
	}
	return true
}

func (h *convertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var convertRequest ConvertRequest
	err := json.NewDecoder(r.Body).Decode(&convertRequest)
	if err != nil {
		returnError(w, "Failed to decode data")
		return
	}
	if !convertRequest.IsValid() {
		returnError(w, fmt.Sprintf("Invalid format: %s", convertRequest.Value))
		return
	}
	if _, ok := h.currencies[convertRequest.Value[:3]]; !ok {
		returnError(w, fmt.Sprintf("Unknown currency: %s", convertRequest.Value[:3]))
		return
	}

	num, _ := strconv.Atoi(convertRequest.Value[3:])
	currency := h.currencies[convertRequest.Value[:3]]
	answer := math.Floor(h.currencies["JPY"] / currency * float64(num))

	convertResponse := ConvertResponse{Answer: int(answer)}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(convertResponse)
}

// ErrorJSON defines error message as JSON
type ErrorJSON struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func returnError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	errorJSON := ErrorJSON{
		Error:   "Bad Request",
		Message: message,
	}
	json.NewEncoder(w).Encode(errorJSON)
}
