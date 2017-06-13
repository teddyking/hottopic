package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/common/log"

	"golang.org/x/sync/syncmap"
)

var (
	appsToTopics *syncmap.Map
	topicsToApps *syncmap.Map
	topics       *syncmap.Map
)

type MapRequest struct {
	App   string `json:"app"`
	Topic string `json:"topic"`
}

func init() {
	appsToTopics = &syncmap.Map{}
	topicsToApps = &syncmap.Map{}
	topics = &syncmap.Map{}
}

func registerAppToTopic(w http.ResponseWriter, r *http.Request) {
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

	appsToTopics.Store(mapRequest.App, mapRequest.Topic)
	topicsToApps.Store(mapRequest.Topic, mapRequest.App)
	topics.LoadOrStore(mapRequest.Topic, make(chan interface{}, 100))
	log.Infof("app '%s' registered on topic '%s'", mapRequest.App, mapRequest.Topic)

	w.WriteHeader(http.StatusCreated)
}

func postToTopic(w http.ResponseWriter, r *http.Request) {
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

	log.Infof("Wrote new value to topic '%s'", topicName)

	w.WriteHeader(http.StatusCreated)

	logTopics()
}

func readFromTopic(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appName := params["app"]

	topicName, ok := appsToTopics.Load(appName)
	if !ok {
		http.Error(w, fmt.Sprintf("App %s has not been registered to a Topic", appName), http.StatusInternalServerError)
		return
	}

	topicMapEntry, ok := topics.Load(topicName)
	if !ok {
		http.Error(w, fmt.Sprintf("Topic %s does not exist", topicName), http.StatusInternalServerError)
		return
	}

	topic := topicMapEntry.(chan interface{})
	payload := <-topic
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshaling payload - %s", err.Error()), http.StatusInternalServerError)
		return
	}

	log.Infof("Data read from topic '%s'", topicName)
	_, err = w.Write(payloadBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error writing payload to http response - %s", err.Error()), http.StatusInternalServerError)
		return
	}

	logTopics()
}

func logTopics() {
	topics.Range(func(key, value interface{}) bool {
		topicName := key.(string)
		topic := value.(chan interface{})
		fmt.Printf("%s: %d\n", topicName, len(topic))
		return true
	})
}

func autoscale() {
	for {
		topicsToApps.Range(func(key, value interface{}) bool {
			topicName := key.(string)
			appName := value.(string)

			topic, _ := topics.Load(topicName)

			numberOfInstances := len(topic.(chan interface{}))

			scale(appName, numberOfInstances)

			return true
		})

		time.Sleep(time.Second * 2)
	}
}

func main() {
	port := os.Getenv("PORT")

	go autoscale()

	rtr := mux.NewRouter()

	rtr.HandleFunc("/map", registerAppToTopic).Methods("POST")
	rtr.HandleFunc("/topic/{topic:[a-zA-Z0-9]+}", postToTopic).Methods("POST") // TODO: update regex?
	rtr.HandleFunc("/map/{app:[a-zA-Z0-9]+}", readFromTopic).Methods("GET")    // TODO: update regex?

	http.Handle("/", rtr)
	_ = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
