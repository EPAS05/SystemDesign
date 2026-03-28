package main

import (
	"bufio"
	"classifier/internal/models"
	"classifier/internal/repository"
	"context"
	"fmt"
	"strconv"
	"time"
)

func unitMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Операции с ЕИ ---")
		fmt.Println("1. Создать ЕИ")
		fmt.Println("2. Список всех ЕИ")
		fmt.Println("3. Информация об ЕИ")
		fmt.Println("4. Изменить ЕИ")
		fmt.Println("5. Удалить ЕИ")
		fmt.Println("6. Главное меню")
		fmt.Print("Выбор: ")

		choice := readLine(reader)

		switch choice {
		case "1":
			createUnit(repo, reader)
		case "2":
			listUnits(repo)
		case "3":
			getUnit(repo, reader)
		case "4":
			updateUnit(repo, reader)
		case "5":
			deleteUnit(repo, reader)
		case "6":
			return
		default:
			fmt.Println("Неправильный выбор. Попробуйте снова")
		}
	}
}

func createUnit(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Имя: ")
	name := readLine(reader)
	fmt.Print("Множитель (например, 0.001 для мм): ")
	multStr := readLine(reader)
	mult, err := strconv.ParseFloat(multStr, 64)
	if err != nil {
		fmt.Println("Неправильно.")
		return
	}
	req := models.CreateUnitRequest{
		Name:       name,
		Multiplier: mult,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	unit, err := repo.CreateUnit(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("ЕИ создана с ID: %d\n", unit.ID)
}

func listUnits(repo repository.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	units, err := repo.GetAllUnits(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(units) == 0 {
		fmt.Println("Не найдено ЕИ")
		return
	}
	fmt.Println("ЕИ:")
	for _, u := range units {
		fmt.Printf("ID: %d, Имя: %s, Множитель: %g\n", u.ID, u.Name, u.Multiplier)
	}
}

func getUnit(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID: ")
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неправильный ID.")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	unit, err := repo.GetUnit(ctx, id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("ID: %d, Имя: %s, Множитель: %g\n", unit.ID, unit.Name, unit.Multiplier)
}

func updateUnit(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID: ")
	id := readID(reader)
	if id == nil {
		return
	}
	ctxGet, cancelGet := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelGet()
	unit, err := repo.GetUnit(ctxGet, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Текущее имя: %s\n", unit.Name)
	fmt.Print("Новое имя (оставьте пустым, чтобы сохранить): ")
	name := readLine(reader)
	if name == "" {
		name = unit.Name
	}
	fmt.Printf("Текущий множитель: %g\n", unit.Multiplier)
	fmt.Print("Новый множитель (оставьте пустым, чтобы сохранить): ")
	multStr := readLine(reader)
	mult := unit.Multiplier
	if multStr != "" {
		mult, err = strconv.ParseFloat(multStr, 64)
		if err != nil {
			fmt.Println("Неправильный множитель.")
			return
		}
	}
	req := models.UpdateUnitRequest{
		ID:         *id,
		Name:       name,
		Multiplier: mult,
	}

	ctxUpdate, cancelUpdate := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelUpdate()
	err = repo.UpdateUnit(ctxUpdate, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("ЕИ изменен.")
}

func deleteUnit(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID ЕИ для удаления: ")
	id := readID(reader)
	if id == nil {
		return
	}
	fmt.Print("Уверены? (yes/no): ")
	confirm := readLine(reader)
	if confirm != "yes" {
		fmt.Println("Удаление отменено.")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := repo.DeleteUnit(ctx, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("ЕИ удалена.")
}
