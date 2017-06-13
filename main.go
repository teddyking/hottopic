package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/sync/syncmap"
)

var (
	appsToTopics map[string]string
	topics       *syncmap.Map
)

type MapRequest struct {
	App   string `json:"app"`
	Topic string `json:"topic"`
}

func init() {
	appsToTopics = make(map[string]string)
	topics = &syncmap.Map{}
}

func mapHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mapRequest := MapRequest{}
	err = json.Unmarshal(body, &mapRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appsToTopics[mapRequest.App] = mapRequest.Topic
	topics.LoadOrStore(mapRequest.Topic, make(chan MapRequest, 100))

	w.WriteHeader(http.StatusCreated)
}

func main() {
	port := os.Getenv("PORT")

	http.HandleFunc("/map", mapHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
