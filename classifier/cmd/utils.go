package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func getDBConnStr() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "classifier")
	password := getEnv("DB_PASSWORD", "secret")
	dbname := getEnv("DB_NAME", "classifier")
	sslmode := getEnv("DB_SSLMODE", "disable")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func readID(reader *bufio.Reader) *int {
	fmt.Print("Введите ID узла: ")
	idStr := readLine(reader)
	if idStr == "" {
		fmt.Println("ID не может быть пустым")
		return nil
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неправильный ID (Введите число).")
		return nil
	}
	return &id
}

func boolPtr(b bool) *bool {
	return &b
}
