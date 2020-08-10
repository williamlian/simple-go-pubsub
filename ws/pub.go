package ws

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type pubHandler struct {
}

/* ServeHTTP implements http.Handler
Publish handler takes the last path element as topic name and post body as data.*/
func (h *pubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	topic := paths[len(paths)-1]
	if topic == "" {
		log.Print("Error: empty topic")
		w.WriteHeader(400)
		fmt.Fprint(w, "Error: empty topic")
		return
	}
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}
	pubsubSingle.Publish(topic, string(bytes))
	w.WriteHeader(200)
	fmt.Fprintln(w, "OK")
}
