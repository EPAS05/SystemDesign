package main

import (
	"bufio"
	"classifier/internal/cli_handlers"
	"classifier/internal/db"
	"classifier/internal/repository"
	"classifier/internal/utils"
	"fmt"
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
	reader := bufio.NewReader(os.Stdin)

	cli_handlers.StartCLI(repo, reader)
}
