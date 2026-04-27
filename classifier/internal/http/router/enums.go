package router

import (
	"classifier/internal/api_handlers"

	"github.com/gorilla/mux"
)

func registerEnumRoutes(r *mux.Router, handler *api_handlers.EnumHandler) {
	r.HandleFunc("/enums", handler.CreateEnum).Methods("POST")
	r.HandleFunc("/enums", handler.GetAllEnums).Methods("GET")
	r.HandleFunc("/enums/type/{type_node_id:[0-9]+}", handler.GetEnumsByTypeNode).Methods("GET")
	r.HandleFunc("/enums/{id:[0-9]+}", handler.GetEnum).Methods("GET")
	r.HandleFunc("/enums/{id:[0-9]+}", handler.UpdateEnum).Methods("PUT")
	r.HandleFunc("/enums/{id:[0-9]+}", handler.DeleteEnum).Methods("DELETE")

	r.HandleFunc("/enums/{enum_id:[0-9]+}/values", handler.CreateEnumValue).Methods("POST")
	r.HandleFunc("/enums/{enum_id:[0-9]+}/values", handler.GetEnumValues).Methods("GET")
	r.HandleFunc("/enums/values/{value_id:[0-9]+}", handler.GetEnumValue).Methods("GET")
	r.HandleFunc("/enums/values/{value_id:[0-9]+}", handler.UpdateEnumValue).Methods("PUT")
	r.HandleFunc("/enums/values/{value_id:[0-9]+}", handler.DeleteEnumValue).Methods("DELETE")
	r.HandleFunc("/enums/{enum_id:[0-9]+}/values/reorder", handler.ReorderEnumValues).Methods("POST")
}
