package cli_handlers

import (
	"bufio"
	"classifier/internal/models"
	"classifier/internal/repository"
	"context"
	"fmt"
	"strconv"
	"time"
)

func documentMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Работа с документами ---")
		fmt.Println("1. Работа с заказчиками")
		fmt.Println("2. Накладная")
		fmt.Println("3. Назад")
		fmt.Print("Выбор: ")

		switch readLine(reader) {
		case "1":
			customerMenu(repo, reader)
		case "2":
			invoiceMenu(repo, reader)
		case "3":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func customerMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Работа с заказчиками ---")
		fmt.Println("1. Создать заказчика")
		fmt.Println("2. Получить заказчика по ID")
		fmt.Println("3. Список заказчиков")
		fmt.Println("4. Обновить заказчика")
		fmt.Println("5. Удалить заказчика")
		fmt.Println("6. Назад")
		fmt.Print("Выбор: ")

		switch readLine(reader) {
		case "1":
			createCustomer(repo, reader)
		case "2":
			getCustomer(repo, reader)
		case "3":
			listCustomers(repo)
		case "4":
			updateCustomer(repo, reader)
		case "5":
			deleteCustomer(repo, reader)
		case "6":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func createCustomer(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Название заказчика: ")
	name := readLine(reader)
	if name == "" {
		fmt.Println("Название обязательно.")
		return
	}

	fmt.Print("ИНН (необязательно): ")
	taxIDStr := readLine(reader)
	var taxID *string
	if taxIDStr != "" {
		taxID = &taxIDStr
	}

	fmt.Print("Адрес (необязательно): ")
	addressStr := readLine(reader)
	var address *string
	if addressStr != "" {
		address = &addressStr
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	customer, err := repo.CreateCustomer(ctx, models.CreateCustomerRequest{
		Name:    name,
		TaxID:   taxID,
		Address: address,
	})
	if err != nil {
		fmt.Printf("Ошибка создания заказчика: %v\n", err)
		return
	}

	fmt.Printf("Заказчик создан. ID=%d, Name=%s\n", customer.ID, customer.Name)
}

func getCustomer(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("ID заказчика: ")
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		fmt.Println("Неверный ID.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	customer, err := repo.GetCustomer(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			fmt.Println("Заказчик не найден.")
			return
		}
		fmt.Printf("Ошибка получения заказчика: %v\n", err)
		return
	}

	printCustomer(customer)
}

func listCustomers(repo repository.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	customers, err := repo.GetAllCustomers(ctx)
	if err != nil {
		fmt.Printf("Ошибка получения списка заказчиков: %v\n", err)
		return
	}
	if len(customers) == 0 {
		fmt.Println("Заказчики отсутствуют.")
		return
	}

	fmt.Println("Список заказчиков:")
	for _, customer := range customers {
		printCustomer(customer)
	}
}

func updateCustomer(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("ID заказчика для обновления: ")
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		fmt.Println("Неверный ID.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	current, err := repo.GetCustomer(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			fmt.Println("Заказчик не найден.")
			return
		}
		fmt.Printf("Ошибка получения заказчика: %v\n", err)
		return
	}

	fmt.Printf("Название (%s): ", current.Name)
	name := readLine(reader)
	if name == "" {
		name = current.Name
	}

	fmt.Printf("ИНН (%s), пусто - оставить, '-' - очистить: ", strPtrOrDash(current.TaxID))
	taxIDInput := readLine(reader)
	taxID := current.TaxID
	switch taxIDInput {
	case "":
	case "-":
		taxID = nil
	default:
		taxID = &taxIDInput
	}

	fmt.Printf("Адрес (%s), пусто - оставить, '-' - очистить: ", strPtrOrDash(current.Address))
	addressInput := readLine(reader)
	address := current.Address
	switch addressInput {
	case "":
	case "-":
		address = nil
	default:
		address = &addressInput
	}

	customer, err := repo.UpdateCustomer(ctx, models.UpdateCustomerRequest{
		ID:      id,
		Name:    name,
		TaxID:   taxID,
		Address: address,
	})
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			fmt.Println("Заказчик не найден.")
			return
		}
		fmt.Printf("Ошибка обновления заказчика: %v\n", err)
		return
	}

	fmt.Printf("Заказчик обновлён. ID=%d, Name=%s\n", customer.ID, customer.Name)
}

func deleteCustomer(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("ID заказчика для удаления: ")
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		fmt.Println("Неверный ID.")
		return
	}

	fmt.Print("Подтвердите удаление (yes/no): ")
	if readLine(reader) != "yes" {
		fmt.Println("Удаление отменено.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = repo.DeleteCustomer(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			fmt.Println("Заказчик не найден.")
			return
		}
		fmt.Printf("Ошибка удаления заказчика: %v\n", err)
		return
	}

	fmt.Println("Заказчик удалён.")
}

func printCustomer(customer *models.Customer) {
	fmt.Printf("ID=%d | Name=%s | INN=%s | Address=%s\n",
		customer.ID,
		customer.Name,
		strPtrOrDash(customer.TaxID),
		strPtrOrDash(customer.Address),
	)
}

func strPtrOrDash(value *string) string {
	if value == nil || *value == "" {
		return "-"
	}
	return *value
}
