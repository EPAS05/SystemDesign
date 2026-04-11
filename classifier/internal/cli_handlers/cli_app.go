package cli_handlers

import (
	"bufio"
	"classifier/internal/repository"
	"fmt"
)

func StartCLI(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n=== Главное меню ===")
		fmt.Println("1. Операции с узлами")
		fmt.Println("2. Операции с продуктами")
		fmt.Println("3. Операции с ЕИ")
		fmt.Println("4. Операции с перечислениями")
		fmt.Println("5. Параметры")
		fmt.Println("6. Выход")
		fmt.Print("Выбор: ")

		choice := readLine(reader)
		switch choice {
		case "1":
			nodeMenu(repo, reader)
		case "2":
			productMenu(repo, reader)
		case "3":
			unitMenu(repo, reader)
		case "4":
			enumMenu(repo, reader)
		case "5":
			paramMenu(repo, reader)
		case "6":
			fmt.Println("Завершение работы!")
			return
		default:
			fmt.Println("Неправильный выбор. Попробуйте снова")
		}
	}
}
