package main

import (
	"classifier/internal/api_handlers"
	"classifier/internal/db"
	"classifier/internal/repository"
	"classifier/internal/utils"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	connStr := utils.GetDBConnStr()
	database, err := db.NewConnection(connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	repo := repository.NewPostgresRepository(database)

	nodeHandler := &api_handlers.NodeHandler{Repo: repo}
	unitHandler := &api_handlers.UnitHandler{Repo: repo}
	enumHandler := &api_handlers.EnumHandler{Repo: repo}
	productHandler := &api_handlers.ProductHandler{Repo: repo, NodeRepo: repo}
	parameterHandler := &api_handlers.ParameterHandler{
		ParamRepo:   repo,
		NodeRepo:    repo,
		EnumRepo:    repo,
		UnitRepo:    repo,
		ProductRepo: repo,
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/nodes", nodeHandler.CreateNode).Methods("POST")
	api.HandleFunc("/nodes/{id:[0-9]+}", nodeHandler.GetNode).Methods("GET")
	api.HandleFunc("/nodes/{id:[0-9]+}/children", nodeHandler.GetChildren).Methods("GET")
	api.HandleFunc("/nodes/{id:[0-9]+}/descendants", nodeHandler.GetAllDescendants).Methods("GET")
	api.HandleFunc("/nodes/{id:[0-9]+}/ancestors", nodeHandler.GetAllAncestors).Methods("GET")
	api.HandleFunc("/nodes/{id:[0-9]+}/parent", nodeHandler.GetParent).Methods("GET")
	api.HandleFunc("/nodes/{id:[0-9]+}/parent", nodeHandler.SetParent).Methods("PUT")
	api.HandleFunc("/nodes/{id:[0-9]+}/name", nodeHandler.SetName).Methods("PUT")
	api.HandleFunc("/nodes/{id:[0-9]+}/order", nodeHandler.SetNodeOrder).Methods("PUT")
	api.HandleFunc("/nodes/{id:[0-9]+}", nodeHandler.DeleteNode).Methods("DELETE")
	api.HandleFunc("/nodes/{id:[0-9]+}/terminal-descendants", nodeHandler.GetTerminalDescendants).Methods("GET")

	api.HandleFunc("/units", unitHandler.CreateUnit).Methods("POST")
	api.HandleFunc("/units", unitHandler.GetAllUnits).Methods("GET")
	api.HandleFunc("/units/{id:[0-9]+}", unitHandler.GetUnit).Methods("GET")
	api.HandleFunc("/units/{id:[0-9]+}", unitHandler.UpdateUnit).Methods("PUT")
	api.HandleFunc("/units/{id:[0-9]+}", unitHandler.DeleteUnit).Methods("DELETE")

	api.HandleFunc("/enums", enumHandler.CreateEnum).Methods("POST")
	api.HandleFunc("/enums", enumHandler.GetAllEnums).Methods("GET")
	api.HandleFunc("/enums/type/{type_node_id:[0-9]+}", enumHandler.GetEnumsByTypeNode).Methods("GET")
	api.HandleFunc("/enums/{id:[0-9]+}", enumHandler.GetEnum).Methods("GET")
	api.HandleFunc("/enums/{id:[0-9]+}", enumHandler.UpdateEnum).Methods("PUT")
	api.HandleFunc("/enums/{id:[0-9]+}", enumHandler.DeleteEnum).Methods("DELETE")

	api.HandleFunc("/enums/{enum_id:[0-9]+}/values", enumHandler.CreateEnumValue).Methods("POST")
	api.HandleFunc("/enums/{enum_id:[0-9]+}/values", enumHandler.GetEnumValues).Methods("GET")
	api.HandleFunc("/enums/values/{value_id:[0-9]+}", enumHandler.GetEnumValue).Methods("GET")
	api.HandleFunc("/enums/values/{value_id:[0-9]+}", enumHandler.UpdateEnumValue).Methods("PUT")
	api.HandleFunc("/enums/values/{value_id:[0-9]+}", enumHandler.DeleteEnumValue).Methods("DELETE")
	api.HandleFunc("/enums/{enum_id:[0-9]+}/values/reorder", enumHandler.ReorderEnumValues).Methods("POST")

	api.HandleFunc("/products", productHandler.CreateProduct).Methods("POST")
	api.HandleFunc("/products/{id:[0-9]+}", productHandler.GetProduct).Methods("GET")
	api.HandleFunc("/products/{id:[0-9]+}", productHandler.UpdateProduct).Methods("PUT")
	api.HandleFunc("/products/{id:[0-9]+}", productHandler.DeleteProduct).Methods("DELETE")
	api.HandleFunc("/nodes/{node_id:[0-9]+}/products", productHandler.GetProductsByClass).Methods("GET")

	api.HandleFunc("/nodes/{node_id:[0-9]+}/parameter-definitions", parameterHandler.CreateParameterDefinition).Methods("POST")
	api.HandleFunc("/nodes/{node_id:[0-9]+}/parameter-definitions", parameterHandler.GetParameterDefinitionsForClass).Methods("GET")
	api.HandleFunc("/parameter-definitions/{id:[0-9]+}", parameterHandler.GetParameterDefinition).Methods("GET")
	api.HandleFunc("/parameter-definitions/{id:[0-9]+}", parameterHandler.UpdateParameterDefinition).Methods("PUT")
	api.HandleFunc("/parameter-definitions/{id:[0-9]+}", parameterHandler.DeleteParameterDefinition).Methods("DELETE")
	api.HandleFunc("/parameter-definitions/{id:[0-9]+}/constraints", parameterHandler.GetParameterConstraints).Methods("GET")

	api.HandleFunc("/products/{product_id:[0-9]+}/parameter-values", parameterHandler.SetParameterValue).Methods("POST")
	api.HandleFunc("/products/{product_id:[0-9]+}/parameter-values", parameterHandler.GetParameterValuesForProduct).Methods("GET")
	api.HandleFunc("/parameter-values/{id:[0-9]+}", parameterHandler.UpdateParameterValue).Methods("PUT")
	api.HandleFunc("/parameter-values/{id:[0-9]+}", parameterHandler.DeleteParameterValue).Methods("DELETE")

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
