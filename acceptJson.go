package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type ChirpParam struct {
	Body string `json:"body"`
}

func jsonHandlerForChirp(w http.ResponseWriter, r *http.Request) ChirpParam {

	decoder := json.NewDecoder(r.Body)
	data := ChirpParam{}
	err := decoder.Decode(&data)
	if err != nil {
		// If errors occur it should respond with appropriate HTTP status codes and a JSON body
		log.Printf("Error decoding JSON: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return ChirpParam{}
	}

	if len(data.Body) > 140 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "Chirp is too long"}`))
		return ChirpParam{}
	}

	data.Body = CheckForBadWords(data.Body)

	return data
}

func CheckForBadWords(body string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	bodySet := strings.Split(body, " ")
	for i, word := range bodySet {
		loweCase := strings.ToLower(word)
		if _, ok := badWords[loweCase]; ok {
			bodySet[i] = "****"
		}
	}
	body = strings.Join(bodySet, " ")
	return body
}
