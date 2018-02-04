package api

import (
	"net/http"
)

// MetadataClient is a helper client used to augment the metadata of the user
// and channel extracted from the registered token.
type MetadataClient interface {
	GetUser(string) string
	GetChannel(string) string
	IsIM(string) string
}

// Server is used to provide API access
type Server struct {
	metadata MetadataClient
}

// NewServer returns a new API Server that will use the provided metadata client
func NewServer(client MetadataClient) Server {
	return Server{
		metadata: client,
	}
}

// Listen starts listening
func (s Server) Listen(path, address string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {

	})
	http.ListenAndServe(address, nil)
}
