package cli_handlers

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

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
