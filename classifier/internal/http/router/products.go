package router

import (
	"classifier/internal/api_handlers"

	"github.com/gorilla/mux"
)

func registerProductRoutes(r *mux.Router, handler *api_handlers.ProductHandler) {
	r.HandleFunc("/products", handler.CreateProduct).Methods("POST")
	r.HandleFunc("/products/{id:[0-9]+}", handler.GetProduct).Methods("GET")
	r.HandleFunc("/products/{id:[0-9]+}", handler.UpdateProduct).Methods("PUT")
	r.HandleFunc("/products/{id:[0-9]+}", handler.DeleteProduct).Methods("DELETE")
	r.HandleFunc("/nodes/{node_id:[0-9]+}/products", handler.GetProductsByClass).Methods("GET")
}
