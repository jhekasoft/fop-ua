package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jhekasoft/fop-ua/fopua"
)

// our main function
func main() {
	addr := ":8010"

	router := mux.NewRouter()
	router.HandleFunc("/calendar", getCalendar).Methods("GET")

	log.Printf("Starting server. Address: %s.", addr)

	log.Fatal(http.ListenAndServe(addr, router))
}

func getCalendar(w http.ResponseWriter, r *http.Request) {
	dataDir := "../../data"

	// Group
	groupParam := r.URL.Query().Get("group")
	parsedGroup, err := strconv.ParseInt(groupParam, 10, 0)
	if err != nil {
		parsedGroup = 1
	}
	group := int(parsedGroup)
	if group > 3 || group < 1 {
		group = 3
	}

	// With PDV
	withPdvParam := r.URL.Query().Get("with_pdv")
	withPdv := false
	if withPdvParam != "" && withPdvParam != "0" {
		withPdv = true
	}

	data, err := fopua.GetFopSingleCalendar(dataDir, group, withPdv)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(calenarResponse{
		Data:  data,
		Group: group,
	})
}

type calenarResponse struct {
	Group int           `json:"group,omitempty"`
	Data  []fopua.Month `json:"data,omitempty"`
}
