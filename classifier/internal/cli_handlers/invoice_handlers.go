package cli_handlers

import (
	"bufio"
	"classifier/internal/models"
	"classifier/internal/repository"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var allowedInvoiceTypes = map[string]struct{}{
	"incoming": {},
	"outgoing": {},
	"return":   {},
}

var allowedInvoiceStatuses = map[string]struct{}{
	"draft":     {},
	"confirmed": {},
	"paid":      {},
	"shipped":   {},
	"cancelled": {},
}

var allowedCurrencies = map[string]struct{}{
	"RUB": {},
	"USD": {},
	"EUR": {},
}

func invoiceMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Накладные ---")
		fmt.Println("1. Создать накладную")
		fmt.Println("2. Список накладных")
		fmt.Println("3. Выбрать накладную для работы")
		fmt.Println("4. Назад")
		fmt.Print("Выбор: ")

		switch readLine(reader) {
		case "1":
			createInvoice(repo, reader)
		case "2":
			listInvoices(repo)
		case "3":
			selectInvoice(repo, reader)
		case "4":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func selectInvoice(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID накладной: ")
	id, ok := readPositiveInt(reader)
	if !ok {
		fmt.Println("Неверный ID.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	invoice, err := repo.GetInvoice(ctx, id)
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			fmt.Println("Накладная не найдена.")
			return
		default:
			fmt.Printf("Ошибка получения накладной: %v\n", err)
			return
		}
	}

	invoiceWorkingMenu(repo, reader, invoice.ID)
}

func invoiceWorkingMenu(repo repository.Repository, reader *bufio.Reader, invoiceID int) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		invoice, err := repo.GetInvoice(ctx, invoiceID)
		cancel()
		if err != nil {
			fmt.Printf("Ошибка получения накладной: %v\n", err)
			return
		}

		fmt.Printf("\n--- Работа с накладной ID=%d (%s) ---\n", invoice.ID, invoice.InvoiceNumber)
		fmt.Println("1. Показать полную информацию")
		fmt.Println("2. Добавить позицию")
		fmt.Println("3. Редактировать позицию")
		fmt.Println("4. Удалить позицию")
		fmt.Println("5. Редактировать реквизиты накладной")
		fmt.Println("6. Удалить накладную")
		fmt.Println("7. Вернуться к списку накладных")
		fmt.Print("Выбор: ")

		switch readLine(reader) {
		case "1":
			showInvoiceDetails(repo, invoiceID)
		case "2":
			addInvoiceItemTo(repo, reader, invoiceID)
		case "3":
			updateInvoiceItemFor(repo, reader, invoiceID)
		case "4":
			deleteInvoiceItemFor(repo, reader, invoiceID)
		case "5":
			updateInvoiceDetails(repo, reader, invoiceID)
		case "6":
			if confirmDeleteInvoice(repo, reader, invoiceID) {
				return // после удаления накладной возвращаемся в предыдущее меню
			}
		case "7":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func showInvoiceDetails(repo repository.Repository, invoiceID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	invoice, err := repo.GetInvoice(ctx, invoiceID)
	if err != nil {
		fmt.Printf("Ошибка получения накладной: %v\n", err)
		return
	}
	printInvoice(invoice)

	items, err := repo.GetInvoiceItems(ctx, invoiceID)
	if err != nil {
		fmt.Printf("Ошибка получения позиций: %v\n", err)
		return
	}
	if len(items) == 0 {
		fmt.Println("Позиции отсутствуют.")
	} else {
		fmt.Println("\nПозиции накладной:")
		for _, item := range items {
			fmt.Printf("  ID=%d | продукт=%d | кол-во=%.3f | цена=%.2f | скидка=%s | итого=%.2f\n",
				item.ID, item.ProductID, item.Quantity, item.UnitPrice,
				floatPtrOrDash(item.DiscountPercent), item.TotalLine)
		}
	}
}

func addInvoiceItemTo(repo repository.Repository, reader *bufio.Reader, invoiceID int) {
	fmt.Print("ID продукта: ")
	productID, ok := readPositiveInt(reader)
	if !ok {
		fmt.Println("Неверный ID продукта.")
		return
	}

	fmt.Print("Количество (>0): ")
	quantity, ok := readPositiveFloat(reader)
	if !ok || quantity <= 0 {
		fmt.Println("Неверное количество.")
		return
	}

	fmt.Print("Цена за единицу (>=0): ")
	unitPrice, ok := readPositiveOrZeroFloat(reader)
	if !ok {
		fmt.Println("Неверная цена.")
		return
	}

	discountPercent, ok := readOptionalFloatPtr(reader, "Скидка % (пусто = 0): ")
	if !ok {
		fmt.Println("Неверное значение скидки.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	item, err := repo.AddInvoiceItem(ctx, models.CreateInvoiceItemRequest{
		InvoiceID:       invoiceID,
		ProductID:       productID,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		DiscountPercent: discountPercent,
	})
	if err != nil {
		switch err {
		case repository.ErrInvoiceNotDraft:
			fmt.Println("Нельзя изменять позиции: накладная не в статусе draft.")
		case repository.ErrNotFound:
			fmt.Println("Накладная или продукт не найден(ы).")
		default:
			fmt.Printf("Ошибка добавления позиции: %v\n", err)
		}
		return
	}
	fmt.Printf("Позиция добавлена. ID=%d, итого=%.2f\n", item.ID, item.TotalLine)
}

func updateInvoiceItemFor(repo repository.Repository, reader *bufio.Reader, invoiceID int) {
	fmt.Print("ID позиции для обновления: ")
	itemID, ok := readPositiveInt(reader)
	if !ok {
		fmt.Println("Неверный ID позиции.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	items, err := repo.GetInvoiceItems(ctx, invoiceID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	found := false
	for _, it := range items {
		if it.ID == itemID {
			found = true
			break
		}
	}
	if !found {
		fmt.Println("Позиция не принадлежит этой накладной.")
		return
	}

	fmt.Print("Новое количество (>0): ")
	quantity, ok := readPositiveFloat(reader)
	if !ok || quantity <= 0 {
		fmt.Println("Неверное количество.")
		return
	}

	fmt.Print("Новая цена за единицу (>=0): ")
	unitPrice, ok := readPositiveOrZeroFloat(reader)
	if !ok {
		fmt.Println("Неверная цена.")
		return
	}

	discountPercent, ok := readOptionalFloatPtr(reader, "Скидка % (пусто = 0): ")
	if !ok {
		fmt.Println("Неверное значение скидки.")
		return
	}

	err = repo.UpdateInvoiceItem(ctx, models.UpdateInvoiceItemRequest{
		ID:              itemID,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		DiscountPercent: discountPercent,
	})
	if err != nil {
		switch err {
		case repository.ErrInvoiceNotDraft:
			fmt.Println("Нельзя изменять позиции: накладная не в статусе draft.")
		case repository.ErrNotFound:
			fmt.Println("Позиция не найдена.")
		default:
			fmt.Printf("Ошибка обновления позиции: %v\n", err)
		}
		return
	}
	fmt.Println("Позиция обновлена.")
}

func deleteInvoiceItemFor(repo repository.Repository, reader *bufio.Reader, invoiceID int) {
	fmt.Print("ID позиции для удаления: ")
	itemID, ok := readPositiveInt(reader)
	if !ok {
		fmt.Println("Неверный ID позиции.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	items, err := repo.GetInvoiceItems(ctx, invoiceID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	found := false
	for _, it := range items {
		if it.ID == itemID {
			found = true
			break
		}
	}
	if !found {
		fmt.Println("Позиция не принадлежит этой накладной.")
		return
	}

	fmt.Print("Подтвердите удаление позиции (yes/no): ")
	if readLine(reader) != "yes" {
		fmt.Println("Удаление отменено.")
		return
	}

	err = repo.DeleteInvoiceItem(ctx, itemID)
	if err != nil {
		switch err {
		case repository.ErrInvoiceNotDraft:
			fmt.Println("Нельзя удалять позиции: накладная не в статусе draft.")
		case repository.ErrNotFound:
			fmt.Println("Позиция не найдена.")
		default:
			fmt.Printf("Ошибка удаления позиции: %v\n", err)
		}
		return
	}
	fmt.Println("Позиция удалена.")
}

func updateInvoiceDetails(repo repository.Repository, reader *bufio.Reader, invoiceID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	current, err := repo.GetInvoice(ctx, invoiceID)
	if err != nil {
		fmt.Printf("Ошибка получения накладной: %v\n", err)
		return
	}

	fmt.Printf("Номер (%s): ", current.InvoiceNumber)
	invoiceNumber := readLine(reader)
	if invoiceNumber == "" {
		invoiceNumber = current.InvoiceNumber
	}

	invoiceDate, ok := readDateWithDefault(reader,
		fmt.Sprintf("Дата (%s) [YYYY-MM-DD]: ", current.InvoiceDate.Format("2006-01-02")),
		current.InvoiceDate,
	)
	if !ok {
		fmt.Println("Неверный формат даты.")
		return
	}

	fmt.Printf("Тип (%s) [incoming/outgoing/return]: ", current.InvoiceType)
	invoiceType := strings.ToLower(readLine(reader))
	if invoiceType == "" {
		invoiceType = current.InvoiceType
	}
	if _, exists := allowedInvoiceTypes[invoiceType]; !exists {
		fmt.Println("Неверный тип накладной.")
		return
	}

	fmt.Printf("Статус (%s) [draft/confirmed/paid/shipped/cancelled]: ", current.Status)
	status := strings.ToLower(readLine(reader))
	if status == "" {
		status = current.Status
	}
	if _, exists := allowedInvoiceStatuses[status]; !exists {
		fmt.Println("Неверный статус накладной.")
		return
	}

	fmt.Printf("ID заказчика (%d): ", current.CustomerID)
	customerInput := readLine(reader)
	customerID := current.CustomerID
	if customerInput != "" {
		parsed, err := strconv.Atoi(customerInput)
		if err != nil || parsed <= 0 {
			fmt.Println("Неверный ID заказчика.")
			return
		}
		customerID = parsed
	}

	fmt.Printf("Валюта (%s) [RUB/USD/EUR]: ", current.Currency)
	currency := strings.ToUpper(readLine(reader))
	if currency == "" {
		currency = current.Currency
	}
	if _, exists := allowedCurrencies[currency]; !exists {
		fmt.Println("Неверная валюта.")
		return
	}

	discountTotal, ok := readOptionalFloatPtrForUpdate(reader, "Суммарная скидка", current.DiscountTotal)
	if !ok {
		fmt.Println("Неверное значение скидки.")
		return
	}

	taxRate, ok := readOptionalFloatPtrForUpdate(reader, "Ставка налога %", current.TaxRate)
	if !ok {
		fmt.Println("Неверное значение налога.")
		return
	}

	fmt.Printf("Комментарий (%s), пусто - оставить, '-' - очистить: ", strPtrOrDash(current.Comment))
	commentInput := readLine(reader)
	comment := current.Comment
	switch commentInput {
	case "":
	case "-":
		comment = nil
	default:
		comment = &commentInput
	}

	err = repo.UpdateInvoice(ctx, models.UpdateInvoiceRequest{
		ID:            current.ID,
		InvoiceNumber: invoiceNumber,
		InvoiceDate:   invoiceDate,
		InvoiceType:   invoiceType,
		Status:        status,
		CustomerID:    customerID,
		Currency:      currency,
		DiscountTotal: discountTotal,
		TaxRate:       taxRate,
		Comment:       comment,
	})
	if err != nil {
		switch err {
		case repository.ErrCustomerNotFound:
			fmt.Println("Заказчик не найден.")
		case repository.ErrNotFound:
			fmt.Println("Накладная не найдена.")
		default:
			fmt.Printf("Ошибка обновления накладной: %v\n", err)
		}
		return
	}
	fmt.Println("Накладная обновлена.")
}

func confirmDeleteInvoice(repo repository.Repository, reader *bufio.Reader, invoiceID int) bool {
	fmt.Print("Удалить накладную? (yes/no): ")
	if readLine(reader) != "yes" {
		fmt.Println("Удаление отменено.")
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := repo.DeleteInvoice(ctx, invoiceID)
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			fmt.Println("Накладная не найдена.")
		default:
			fmt.Printf("Ошибка удаления накладной: %v\n", err)
		}
		return false
	}
	fmt.Println("Накладная удалена.")
	return true
}

func createInvoice(repo repository.Repository, reader *bufio.Reader) {
	fmt.Println("Доступные заказчики:")
	listCustomers(repo)

	fmt.Print("Номер накладной: ")
	invoiceNumber := readLine(reader)
	if invoiceNumber == "" {
		fmt.Println("Номер накладной обязателен.")
		return
	}

	invoiceDate, ok := readDateWithDefault(reader, "Дата накладной (YYYY-MM-DD, пусто = сегодня): ", time.Now())
	if !ok {
		fmt.Println("Неверный формат даты.")
		return
	}

	fmt.Print("Тип (incoming/outgoing/return): ")
	invoiceType := strings.ToLower(readLine(reader))
	if _, exists := allowedInvoiceTypes[invoiceType]; !exists {
		fmt.Println("Неверный тип накладной.")
		return
	}

	fmt.Print("ID заказчика: ")
	customerID, ok := readPositiveInt(reader)
	if !ok {
		fmt.Println("Неверный ID заказчика.")
		return
	}

	fmt.Print("Валюта (RUB/USD/EUR, пусто = RUB): ")
	currency := strings.ToUpper(readLine(reader))
	if currency == "" {
		currency = "RUB"
	}
	if _, exists := allowedCurrencies[currency]; !exists {
		fmt.Println("Неверная валюта.")
		return
	}

	discountTotal, ok := readOptionalFloatPtr(reader, "Суммарная скидка (пусто = 0): ")
	if !ok {
		fmt.Println("Неверное значение скидки.")
		return
	}

	taxRate, ok := readOptionalFloatPtr(reader, "Ставка налога % (пусто = 0): ")
	if !ok {
		fmt.Println("Неверное значение налога.")
		return
	}

	fmt.Print("Комментарий (необязательно): ")
	commentInput := readLine(reader)
	var comment *string
	if commentInput != "" {
		comment = &commentInput
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	invoice, err := repo.CreateInvoice(ctx, models.CreateInvoiceRequest{
		InvoiceNumber: invoiceNumber,
		InvoiceDate:   invoiceDate,
		InvoiceType:   invoiceType,
		CustomerID:    customerID,
		Currency:      currency,
		DiscountTotal: discountTotal,
		TaxRate:       taxRate,
		Comment:       comment,
	})
	if err != nil {
		switch err {
		case repository.ErrCustomerNotFound:
			fmt.Println("Заказчик не найден.")
			return
		default:
			fmt.Printf("Ошибка создания накладной: %v\n", err)
			return
		}
	}
	fmt.Printf("Накладная создана. ID=%d, Number=%s, Status=%s\n", invoice.ID, invoice.InvoiceNumber, invoice.Status)
}

func listInvoices(repo repository.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	invoices, err := repo.GetAllInvoices(ctx)
	if err != nil {
		fmt.Printf("Ошибка получения списка накладных: %v\n", err)
		return
	}
	if len(invoices) == 0 {
		fmt.Println("Накладные отсутствуют.")
		return
	}
	fmt.Println("Список накладных:")
	for _, inv := range invoices {
		fmt.Printf("ID=%d | №%s | %s | %s | сумма=%.2f %s\n",
			inv.ID, inv.InvoiceNumber, inv.InvoiceDate.Format("2006-01-02"), inv.Status,
			inv.TotalAmount, inv.Currency)
	}
}

func printInvoice(invoice *models.Invoice) {
	fmt.Printf("\n--- Накладная ID=%d ---\n", invoice.ID)
	fmt.Printf("Номер: %s\n", invoice.InvoiceNumber)
	fmt.Printf("Дата: %s\n", invoice.InvoiceDate.Format("2006-01-02"))
	fmt.Printf("Тип: %s\n", invoice.InvoiceType)
	fmt.Printf("Статус: %s\n", invoice.Status)
	fmt.Printf("Заказчик ID: %d\n", invoice.CustomerID)
	fmt.Printf("Валюта: %s\n", invoice.Currency)
	fmt.Printf("Скидка: %s\n", floatPtrOrDash(invoice.DiscountTotal))
	fmt.Printf("Ставка НДС: %s%%\n", floatPtrOrDash(invoice.TaxRate))
	fmt.Printf("Сумма НДС: %s\n", floatPtrOrDash(invoice.TaxAmount))
	fmt.Printf("Итого: %.2f\n", invoice.TotalAmount)
	if invoice.Comment != nil {
		fmt.Printf("Комментарий: %s\n", *invoice.Comment)
	}
}

func readPositiveInt(reader *bufio.Reader) (int, bool) {
	value, err := strconv.Atoi(readLine(reader))
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func readPositiveFloat(reader *bufio.Reader) (float64, bool) {
	value, err := strconv.ParseFloat(readLine(reader), 64)
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func readPositiveOrZeroFloat(reader *bufio.Reader) (float64, bool) {
	value, err := strconv.ParseFloat(readLine(reader), 64)
	if err != nil || value < 0 {
		return 0, false
	}
	return value, true
}

func readOptionalFloatPtr(reader *bufio.Reader, prompt string) (*float64, bool) {
	fmt.Print(prompt)
	valueStr := readLine(reader)
	if valueStr == "" {
		return nil, true
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, false
	}
	return &value, true
}

func readOptionalFloatPtrForUpdate(reader *bufio.Reader, title string, current *float64) (*float64, bool) {
	fmt.Printf("%s (%s), пусто - оставить, '-' - очистить: ", title, floatPtrOrDash(current))
	input := readLine(reader)
	switch input {
	case "":
		return current, true
	case "-":
		return nil, true
	default:
		value, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return nil, false
		}
		return &value, true
	}
}

func readDateWithDefault(reader *bufio.Reader, prompt string, defaultValue time.Time) (time.Time, bool) {
	fmt.Print(prompt)
	input := readLine(reader)
	if input == "" {
		return defaultValue, true
	}

	parsed, err := time.Parse("2006-01-02", input)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func floatPtrOrDash(value *float64) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f", *value)
}

func floatPtrToValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}
