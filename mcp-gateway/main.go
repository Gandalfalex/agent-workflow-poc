package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type JiraTrigger struct {
	IssueKey string `json:"issueKey"`
	Summary  string `json:"summary"`
	Body     string `json:"body"`
	Type     string `json:"type"`
}

func main() {
	http.HandleFunc("/mcp/jira", func(w http.ResponseWriter, r *http.Request) {
		var trigger JiraTrigger
		_ = json.NewDecoder(r.Body).Decode(&trigger)
		log.Printf("Received Jira issue %s (%s)", trigger.IssueKey, trigger.Type)
		w.WriteHeader(http.StatusAccepted)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}