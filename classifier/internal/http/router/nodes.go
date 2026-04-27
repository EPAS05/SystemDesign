package router

import (
	"classifier/internal/api_handlers"

	"github.com/gorilla/mux"
)

func registerNodeRoutes(r *mux.Router, handler *api_handlers.NodeHandler) {
	r.HandleFunc("/nodes", handler.CreateNode).Methods("POST")
	r.HandleFunc("/nodes/{id:[0-9]+}", handler.GetNode).Methods("GET")
	r.HandleFunc("/nodes/{id:[0-9]+}/children", handler.GetChildren).Methods("GET")
	r.HandleFunc("/nodes/{id:[0-9]+}/descendants", handler.GetAllDescendants).Methods("GET")
	r.HandleFunc("/nodes/{id:[0-9]+}/ancestors", handler.GetAllAncestors).Methods("GET")
	r.HandleFunc("/nodes/{id:[0-9]+}/parent", handler.GetParent).Methods("GET")
	r.HandleFunc("/nodes/{id:[0-9]+}/parent", handler.SetParent).Methods("PUT")
	r.HandleFunc("/nodes/{id:[0-9]+}/name", handler.SetName).Methods("PUT")
	r.HandleFunc("/nodes/{id:[0-9]+}/order", handler.SetNodeOrder).Methods("PUT")
	r.HandleFunc("/nodes/{id:[0-9]+}", handler.DeleteNode).Methods("DELETE")
	r.HandleFunc("/nodes/{id:[0-9]+}/terminal-descendants", handler.GetTerminalDescendants).Methods("GET")
}
