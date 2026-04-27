package router

import (
	"classifier/internal/api_handlers"

	"github.com/gorilla/mux"
)

func registerParameterRoutes(r *mux.Router, handler *api_handlers.ParameterHandler) {
	r.HandleFunc("/nodes/{node_id:[0-9]+}/parameter-definitions", handler.CreateParameterDefinition).Methods("POST")
	r.HandleFunc("/nodes/{node_id:[0-9]+}/parameter-definitions", handler.GetParameterDefinitionsForClass).Methods("GET")
	r.HandleFunc("/parameter-definitions/{id:[0-9]+}", handler.GetParameterDefinition).Methods("GET")
	r.HandleFunc("/parameter-definitions/{id:[0-9]+}", handler.UpdateParameterDefinition).Methods("PUT")
	r.HandleFunc("/parameter-definitions/{id:[0-9]+}", handler.DeleteParameterDefinition).Methods("DELETE")
	r.HandleFunc("/parameter-definitions/{id:[0-9]+}/constraints", handler.GetParameterConstraints).Methods("GET")

	r.HandleFunc("/products/{product_id:[0-9]+}/parameter-values", handler.SetParameterValue).Methods("POST")
	r.HandleFunc("/products/{product_id:[0-9]+}/parameter-values", handler.GetParameterValuesForProduct).Methods("GET")
	r.HandleFunc("/parameter-values/{id:[0-9]+}", handler.UpdateParameterValue).Methods("PUT")
	r.HandleFunc("/parameter-values/{id:[0-9]+}", handler.DeleteParameterValue).Methods("DELETE")
	r.HandleFunc("/nodes/{node_id:[0-9]+}/products/search", handler.FindProductsByParameters).Methods("POST")
}
