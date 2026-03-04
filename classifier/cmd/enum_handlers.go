package main

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

func enumMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Операции с перечислениями ---")
		fmt.Println("1. Создать перечисление")
		fmt.Println("2. Список всех перечислений")
		fmt.Println("3. Выбрать перечисление (работа с его значениями)")
		fmt.Println("4. Назад в главное меню")
		fmt.Print("Выбор: ")
		choice := readLine(reader)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		switch choice {
		case "1":
			createEnum(ctx, repo, reader)
		case "2":
			listEnums(ctx, repo)
		case "3":
			selectEnum(ctx, repo, reader)
		case "4":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func createEnum(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите название перечисления: ")
	name := readLine(reader)
	if name == "" {
		fmt.Println("Название обязательно")
		return
	}
	fmt.Print("Введите описание (необязательно): ")
	desc := readLine(reader)
	var description *string
	if desc != "" {
		description = &desc
	}
	req := models.CreateEnumRequest{
		Name:        name,
		Description: description,
	}
	enum, err := repo.CreateEnum(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	fmt.Printf("Перечисление создано с ID: %d\n", enum.ID)
}

func listEnums(ctx context.Context, repo repository.Repository) {
	enums, err := repo.GetAllEnums(ctx)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	if len(enums) == 0 {
		fmt.Println("Нет перечислений.")
		return
	}
	fmt.Println("Список перечислений:")
	for _, e := range enums {
		desc := ""
		if e.Description != nil {
			desc = fmt.Sprintf(" (%s)", *e.Description)
		}
		fmt.Printf("ID: %d, Название: %s%s\n", e.ID, e.Name, desc)
	}
}

func selectEnum(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID перечисления: ")
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}
	enum, err := repo.GetEnum(ctx, id)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	enumValuesMenu(ctx, repo, reader, enum)
}

func enumValuesMenu(ctx context.Context, repo repository.Repository, reader *bufio.Reader, enum *models.Enum) {
	descStr := ""
	if enum.Description != nil {
		descStr = *enum.Description
	}
	for {
		fmt.Printf("\n--- Перечисление: %s ( %s ,ID: %d) ---\n", enum.Name, descStr, enum.ID)
		fmt.Println("1. Добавить значение")
		fmt.Println("2. Список значений")
		fmt.Println("3. Редактировать значение")
		fmt.Println("4. Удалить значение")
		fmt.Println("5. Изменить порядок значений")
		fmt.Println("6. Редактировать себя")
		fmt.Println("7. Вернуться к списку перечислений")
		fmt.Print("Выбор: ")
		choice := readLine(reader)

		switch choice {
		case "1":
			createEnumValue(ctx, repo, reader, enum.ID)
		case "2":
			listEnumValues(ctx, repo, enum.ID)
		case "3":
			updateEnumValue(ctx, repo, reader, enum.ID)
		case "4":
			deleteEnumValue(ctx, repo, reader, enum.ID)
		case "5":
			reorderEnumValues(ctx, repo, reader, enum.ID)
		case "6":
			updateEnum(ctx, repo, reader, enum)
		case "7":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func createEnumValue(ctx context.Context, repo repository.Repository, reader *bufio.Reader, enumID int) {
	fmt.Print("Введите значение: ")
	value := readLine(reader)
	fmt.Print("Порядок (оставьте пустым для авто): ")
	orderStr := readLine(reader)
	var sortOrder *int
	if orderStr != "" {
		order, err := strconv.Atoi(orderStr)
		if err != nil {
			fmt.Println("Неверный порядок.")
			return
		}
		sortOrder = &order
	}
	req := models.CreateEnumValueRequest{
		EnumID:    enumID,
		Value:     value,
		SortOrder: sortOrder,
	}
	ev, err := repo.CreateEnumValue(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	fmt.Printf("Значение создано с ID: %d\n", ev.ID)
}

func listEnumValues(ctx context.Context, repo repository.Repository, enumID int) {
	values, err := repo.GetEnumValues(ctx, enumID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	if len(values) == 0 {
		fmt.Println("Нет значений.")
		return
	}
	fmt.Println("Значения:")
	for _, v := range values {
		fmt.Printf("ID: %d, Значение: %s, Порядок: %d\n", v.ID, v.Value, v.SortOrder)
	}
}

func updateEnum(ctx context.Context, repo repository.Repository, reader *bufio.Reader, enum *models.Enum) {
	fmt.Print("Введите имя (пусто для старого): ")
	name := readLine(reader)
	if name == "" {
		name = enum.Name
	}
	fmt.Print("Введите описание (пусто для старого): ")
	newDesc := readLine(reader)
	var descPtr *string
	if newDesc == "" {
		descPtr = enum.Description
	} else {
		descPtr = &newDesc
	}
	req := models.UpdateEnumRequest{
		ID:          enum.ID,
		Name:        name,
		Description: descPtr,
	}

	err := repo.UpdateEnum(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	fmt.Println("Значение обновлено")
}

func updateEnumValue(ctx context.Context, repo repository.Repository, reader *bufio.Reader, enumID int) {
	fmt.Println("Все значения:")
	listEnumValues(ctx, repo, enumID)
	fmt.Println("\nВведите ID для замены:")
	valueID := readID(reader)
	value, err := repo.GetEnumValue(ctx, *valueID)
	if err != nil {
		fmt.Printf("Ошибка получения значения: %s", err)
		return
	}
	fmt.Print("Введите новое значение (оставьте пустым для сохранения текущего): ")
	newVal := readLine(reader)
	if newVal == "" {
		newVal = value.Value
	}
	req := models.UpdateEnumValueRequest{
		ID:    *valueID,
		Value: newVal,
	}
	err = repo.UpdateEnumValue(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка обновления значения: %v\n", err)
		return
	}

	fmt.Println("Значение обновлено.")
}

func deleteEnumValue(ctx context.Context, repo repository.Repository, reader *bufio.Reader, enumID int) {
	fmt.Println("Все значения:")
	listEnumValues(ctx, repo, enumID)

	fmt.Println("\nВведите ID значения для удаления:")
	valueID := readID(reader)
	if valueID == nil {
		return
	}
	value, err := repo.GetEnumValue(ctx, *valueID)
	if err != nil {
		fmt.Printf("Ошибка получения значения: %v\n", err)
		return
	}
	if value.EnumID != enumID {
		fmt.Println("Значение не принадлежит указанному перечислению.")
		return
	}
	err = repo.DeleteEnumValue(ctx, *valueID)
	if err != nil {
		fmt.Printf("Ошибка удаления значения: %v\n", err)
		return
	}
	fmt.Println("Значение удалено.")
}

func reorderEnumValues(ctx context.Context, repo repository.Repository, reader *bufio.Reader, enumID int) {
	fmt.Println("Текущий порядок значений:")
	listEnumValues(ctx, repo, enumID)

	fmt.Println("\nВведите новый порядок (ID значений через запятую в нужной последовательности):")
	input := readLine(reader)
	parts := strings.Split(input, ",")
	var valueIDs []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.Atoi(p)
		if err != nil {
			fmt.Println("Неверный ID:", p)
			return
		}
		valueIDs = append(valueIDs, id)
	}
	if len(valueIDs) == 0 {
		fmt.Println("Не введено ни одного ID.")
		return
	}
	values, err := repo.GetEnumValues(ctx, enumID)
	if err != nil {
		fmt.Printf("Ошибка получения значений: %v\n", err)
		return
	}
	existingIDs := make(map[int]bool)
	for _, v := range values {
		existingIDs[v.ID] = true
	}
	for _, id := range valueIDs {
		if !existingIDs[id] {
			fmt.Printf("ID %d не принадлежит текущему перечислению.\n", id)
			return
		}
	}
	req := models.ReorderEnumValuesRequest{
		EnumID:   enumID,
		ValueIDs: valueIDs,
	}
	err = repo.ReorderEnumValues(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка изменения порядка: %v\n", err)
		return
	}
	fmt.Println("Порядок обновлён.")
}
