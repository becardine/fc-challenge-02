package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Event struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Date         string `json:"date"`
	Price        int    `json:"price"`
	Rating       string `json:"rating"`
	ImageURL     string `json:"image_url"`
	CreatedAt    string `json:"created_at"`
	Location     string `json:"location"`
}

type Spot struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	EventID int    `json:"event_id"`
}

type Data struct {
	Events []Event `json:"events"`
	Spots  []Spot  `json:"spots"`
}

var data Data

func main() {
	loadData()

	r := mux.NewRouter()

	r.HandleFunc("/events", getEventsHandler).Methods("GET")
	r.HandleFunc("/events/{eventId}", getEventDetailsHandler).Methods("GET")
	r.HandleFunc("/events/{eventId}/spots", getEventSpotsHandler).Methods("GET")
	r.HandleFunc("/event/{eventId}/reserve", reserveSpotHandler).Methods("POST")

	fmt.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func loadData() {
	file, err := ioutil.ReadFile("data.json")
	if err != nil {
		log.Fatalf("Error reading data.json: %v", err)
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Fatalf("Error unmarshalling data.json: %v", err)
	}
}

func getEventsHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, err := json.Marshal(data.Events)
	if err != nil {
		http.Error(w, "Error encoding events to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func getEventDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["eventId"]

	for _, event := range data.Events {
		if fmt.Sprint(event.ID) == eventID {
			jsonResponse, err := json.Marshal(event)
			if err != nil {
				http.Error(w, "Error encoding event details to JSON", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			return
		}
	}

	http.NotFound(w, r)
}

func getEventSpotsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["eventId"]

	spots := []Spot{}
	for _, spot := range data.Spots {
		if fmt.Sprint(spot.EventID) == eventID {
			spots = append(spots, spot)
		}
	}

	jsonResponse, err := json.Marshal(spots)
	if err != nil {
		http.Error(w, "Error encoding event spots to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func reserveSpotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventIDStr := vars["eventId"]

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Spot string `json:"spot"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	spotFound := false
	var spotToUpdate Spot
	for i, spot := range data.Spots {
		if spot.EventID == eventID && spot.Name == requestBody.Spot {
			spotFound = true
			spotToUpdate = data.Spots[i]
			break
		}
	}

	if !spotFound {
		http.Error(w, "Spot not found", http.StatusBadRequest)
		return
	}

	if spotToUpdate.Status == "reserved" {
		http.Error(w, "Spot is already reserved", http.StatusBadRequest)
		return
	}

	spotToUpdate.Status = "reserved"

	for i, spot := range data.Spots {
		if spot.EventID == eventID && spot.Name == requestBody.Spot {
			data.Spots[i] = spotToUpdate
			break
		}
	}

	response := map[string]string{"message": "Spot reserved successfully"}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding reservation response to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
