package main

import (
	"log"
	"net/http"

	"github.com/williamlian/simple-go-pubsub.git/ws"
)

func main() {
	pub, sub := ws.GetHandlers()
	http.Handle("/publish/", pub)
	http.Handle("/subscribe/", sub)
	http.HandleFunc("/sample", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./sample/sample.html")
	})
	log.Fatal(http.ListenAndServe(":7240", nil))
}
