package router

import (
	"classifier/internal/api_handlers"

	"github.com/gorilla/mux"
)

func registerUnitRoutes(r *mux.Router, handler *api_handlers.UnitHandler) {
	r.HandleFunc("/units", handler.CreateUnit).Methods("POST")
	r.HandleFunc("/units", handler.GetAllUnits).Methods("GET")
	r.HandleFunc("/units/{id:[0-9]+}", handler.GetUnit).Methods("GET")
	r.HandleFunc("/units/{id:[0-9]+}", handler.UpdateUnit).Methods("PUT")
	r.HandleFunc("/nodes/{id:[0-9]+}/unit", handler.SetUnit).Methods("PUT")
	r.HandleFunc("/products/{id:[0-9]+}/unit", handler.SetDefaultUnit).Methods("PUT")
	r.HandleFunc("/units/{id:[0-9]+}", handler.DeleteUnit).Methods("DELETE")
}
