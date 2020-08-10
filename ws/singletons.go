package ws

import (
	"net/http"

	"github.com/williamlian/simple-go-pubsub.git/pubsub"
)

var pubsubSingle pubsub.PubSub
var pubWsSingle http.Handler
var subWsSingle http.Handler

var _instantiated bool = false

// GetHandlers returns a publish http handler and the subscription websocker handler
func GetHandlers() (pub http.Handler, sub http.Handler) {
	if !_instantiated {
		pubsubSingle = pubsub.NewPubSub()
		pubWsSingle = &pubHandler{}
		subWsSingle = &subHandler{}
		_instantiated = true
	}
	return pubWsSingle, subWsSingle
}
