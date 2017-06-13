package main

import (
	"fmt"
	"net/http"
	"os"
)

type MapRequest struct {
	App   string `json:"app"`
	Topic string `json:"topic"`
}

func mapHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "sup")
}

func main() {
	port := os.Getenv("PORT")

	http.HandleFunc("/map", mapHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
