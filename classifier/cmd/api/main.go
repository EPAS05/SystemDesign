package main

import (
	"classifier/internal/api_handlers"
	"classifier/internal/db"
	router "classifier/internal/http/router"
	"classifier/internal/repository"
	"classifier/internal/utils"
	"fmt"
	"log"
	"net/http"
	"os"
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
	productHandler := &api_handlers.ProductHandler{Repo: repo, NodeRepo: repo, UnitRepo: repo}
	customerHandler := &api_handlers.CustomerHandler{Repo: repo}
	parameterHandler := &api_handlers.ParameterHandler{
		ParamRepo:   repo,
		NodeRepo:    repo,
		EnumRepo:    repo,
		UnitRepo:    repo,
		ProductRepo: repo,
	}

	r := router.New(router.Handlers{
		Node:      nodeHandler,
		Unit:      unitHandler,
		Enum:      enumHandler,
		Product:   productHandler,
		Customer:  customerHandler,
		Parameter: parameterHandler,
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
