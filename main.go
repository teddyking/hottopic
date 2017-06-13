package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"

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
	topics.LoadOrStore(mapRequest.Topic, make(chan interface{}, 100))

	w.WriteHeader(http.StatusCreated)
}

func topicHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	topicName := params["topic"]

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var payload interface{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	topicMapEntry, ok := topics.Load(topicName)
	if !ok {
		http.Error(w, fmt.Sprintf("Topic %s does not exist", topicName), http.StatusInternalServerError)
		return
	}

	topic := topicMapEntry.(chan interface{})
	topic <- payload

	w.WriteHeader(http.StatusCreated)
}

func main() {
	port := os.Getenv("PORT")

	rtr := mux.NewRouter()

	rtr.HandleFunc("/map", mapHandler).Methods("POST")
	rtr.HandleFunc("/topic/{topic:[a-z]+}", topicHandler).Methods("POST") // TODO: update regex?

	http.Handle("/", rtr)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
