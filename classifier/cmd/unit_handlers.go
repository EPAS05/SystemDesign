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
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		switch choice {
		case "1":
			createUnit(ctx, repo, reader)
		case "2":
			listUnits(ctx, repo)
		case "3":
			getUnit(ctx, repo, reader)
		case "4":
			updateUnit(ctx, repo, reader)
		case "5":
			deleteUnit(ctx, repo, reader)
		case "6":
			return
		default:
			fmt.Println("Неправильный выбор. Попробуйте снова")
		}
	}
}

func createUnit(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
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
	unit, err := repo.CreateUnit(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("ЕИ создана с ID: %d\n", unit.ID)
}

func listUnits(ctx context.Context, repo repository.Repository) {
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

func getUnit(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID: ")
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неправильный ID.")
		return
	}
	unit, err := repo.GetUnit(ctx, id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("ID: %d, Имя: %s, Множитель: %g\n", unit.ID, unit.Name, unit.Multiplier)
}

func updateUnit(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID: ")
	id := readID(reader)
	if id == nil {
		return
	}
	unit, err := repo.GetUnit(ctx, *id)
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
	err = repo.UpdateUnit(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("ЕИ изменен.")
}

func deleteUnit(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
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
	err := repo.DeleteUnit(ctx, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("ЕИ удалена.")
}
