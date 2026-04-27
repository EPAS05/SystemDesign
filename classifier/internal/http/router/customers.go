package router

import (
	"classifier/internal/api_handlers"

	"github.com/gorilla/mux"
)

func registerCustomerRoutes(r *mux.Router, handler *api_handlers.CustomerHandler) {
	r.HandleFunc("/documents/customers", handler.CreateCustomer).Methods("POST")
	r.HandleFunc("/documents/customers", handler.GetAllCustomers).Methods("GET")
	r.HandleFunc("/documents/customers/{id:[0-9]+}", handler.GetCustomer).Methods("GET")
	r.HandleFunc("/documents/customers/{id:[0-9]+}", handler.UpdateCustomer).Methods("PUT")
	r.HandleFunc("/documents/customers/{id:[0-9]+}", handler.DeleteCustomer).Methods("DELETE")
}
