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

func productMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Операции с изделиями (продуктами) ---")
		fmt.Println("1. Создать изделие")
		fmt.Println("2. Список изделий по классу")
		fmt.Println("3. Просмотреть изделие")
		fmt.Println("4. Редактировать изделие")
		fmt.Println("5. Удалить изделие")
		fmt.Println("6. Назад в главное меню")
		fmt.Print("Выбор: ")
		choice := readLine(reader)

		switch choice {
		case "1":
			createProduct(repo, reader)
		case "2":
			listProductsByClass(repo, reader)
		case "3":
			getProduct(repo, reader)
		case "4":
			updateProduct(repo, reader)
		case "5":
			deleteProduct(repo, reader)
		case "6":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func createProduct(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите название изделия: ")
	name := readLine(reader)
	if name == "" {
		fmt.Println("Название обязательно.")
		return
	}

	ctxGetDescendants, cancelGetDescendants := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelGetDescendants()
	classes, err := repo.GetAllDescendants(ctxGetDescendants, 2)
	if err != nil {
		fmt.Printf("Ошибка получения классов: %v\n", err)
		return
	}
	var available []*models.Node
	for _, c := range classes {
		if c.IsTerminal == nil || (c.IsTerminal != nil && *c.IsTerminal) {
			available = append(available, c)
		}
	}
	if len(available) == 0 {
		fmt.Println("Нет доступных классов для добавления изделий. Сначала создайте класс через меню узлов.")
		return
	}
	fmt.Println("Доступные классы (метаклассы):")
	for _, c := range available {
		fmt.Printf("  ID: %d, Название: %s\n", c.ID, c.Name)
	}
	fmt.Print("Введите ID класса: ")
	classIDStr := readLine(reader)
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}

	ctxCheckClass, cancelCheckClass := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelCheckClass()
	targetClass, err := repo.GetNode(ctxCheckClass, classID)
	if err != nil {
		fmt.Printf("Ошибка получения класса: %v\n", err)
		return
	}
	if targetClass.IsTerminal != nil && *targetClass.IsTerminal == false {
		fmt.Println("Выбранный класс не может содержать изделия (is_terminal = false).")
		return
	}

	fmt.Print("Тип единицы (mass/length/piece, оставьте пустым если не требуется): ")
	ut := strings.ToLower(readLine(reader))
	var unitType *string
	if ut != "" && (ut == "mass" || ut == "length" || ut == "piece") {
		unitType = &ut
	}

	fmt.Print("Вес погонного метра (т/м, оставьте пустым если не требуется): ")
	wStr := readLine(reader)
	var weightPerMeter *float64
	if wStr != "" {
		w, err := strconv.ParseFloat(wStr, 64)
		if err == nil {
			weightPerMeter = &w
		}
	}

	fmt.Print("Длина одной штуки (м, оставьте пустым если не требуется): ")
	pStr := readLine(reader)
	var pieceLength *float64
	if pStr != "" {
		p, err := strconv.ParseFloat(pStr, 64)
		if err == nil {
			pieceLength = &p
		}
	}

	fmt.Print("ID единицы измерения по умолчанию (оставьте пустым если не требуется): ")
	defStr := readLine(reader)
	var defaultUnitID *int
	if defStr != "" {
		defID, err := strconv.Atoi(defStr)
		if err == nil {
			defaultUnitID = &defID
		}
	}

	req := models.CreateProductRequest{
		Name:           name,
		ClassNodeID:    classID,
		UnitType:       unitType,
		WeightPerMeter: weightPerMeter,
		PieceLength:    pieceLength,
		DefaultUnitID:  defaultUnitID,
	}

	ctxCreate, cancelCreate := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelCreate()
	product, err := repo.CreateProduct(ctxCreate, req)
	if err != nil {
		fmt.Printf("Ошибка создания изделия: %v\n", err)
		return
	}
	fmt.Printf("Изделие создано с ID: %d\n", product.ID)

	if targetClass.IsTerminal == nil {
		err = repo.UpdateNodeIsTerminal(ctxCreate, classID, boolPtr(true))
		if err != nil {
			fmt.Printf("Предупреждение: не удалось обновить статус класса: %v\n", err)
		} else {
			fmt.Println("Класс стал терминальным (теперь может принимать только изделия).")
		}
	}
}

func listProductsByClass(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID класса (метакласса): ")
	idStr := readLine(reader)
	classID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	products, err := repo.GetProductsByClass(ctx, classID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	if len(products) == 0 {
		fmt.Println("В этом классе нет изделий.")
		return
	}
	fmt.Println("Изделия:")
	for _, p := range products {
		fmt.Printf("ID: %d, Название: %s\n", p.ID, p.Name)
	}
}

func getProduct(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID изделия: ")
	idStr := readLine(reader)
	productID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	product, err := repo.GetProduct(ctx, productID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	printProduct(product)

	fmt.Println("\nПараметры изделия:")
	showProductParameters(repo, productID)
}

func updateProduct(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID изделия для редактирования: ")
	idStr := readLine(reader)
	productID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}

	ctxGet, cancelGet := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelGet()
	product, err := repo.GetProduct(ctxGet, productID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	printProduct(product)

	fmt.Print("Новое название (оставьте пустым для сохранения): ")
	name := readLine(reader)
	if name == "" {
		name = product.Name
	}

	fmt.Printf("Текущий класс ID: %d\n", product.ClassNodeID)
	fmt.Print("Новый ID класса (оставьте пустым для сохранения): ")
	classStr := readLine(reader)
	classID := product.ClassNodeID
	if classStr != "" {
		classID, err = strconv.Atoi(classStr)
		if err != nil {
			fmt.Println("Неверный ID класса.")
			return
		}
	}

	fmt.Print("Новый тип единицы (mass/length/piece, оставьте пустым для сохранения): ")
	ut := strings.ToLower(readLine(reader))
	unitType := product.UnitType
	if ut != "" && (ut == "mass" || ut == "length" || ut == "piece") {
		unitType = &ut
	} else if ut == "" {
		unitType = product.UnitType
	} else {
		fmt.Println("Неверный тип, оставляем как есть.")
	}

	fmt.Print("Новый вес погонного метра (т/м, оставьте пустым для сохранения): ")
	wStr := readLine(reader)
	weightPerMeter := product.WeightPerMeter
	if wStr != "" {
		w, err := strconv.ParseFloat(wStr, 64)
		if err == nil {
			weightPerMeter = &w
		} else {
			fmt.Println("Неверное число, оставляем как есть.")
		}
	}

	fmt.Print("Новая длина одной штуки (м, оставьте пустым для сохранения): ")
	pStr := readLine(reader)
	pieceLength := product.PieceLength
	if pStr != "" {
		p, err := strconv.ParseFloat(pStr, 64)
		if err == nil {
			pieceLength = &p
		} else {
			fmt.Println("Неверное число, оставляем как есть.")
		}
	}

	fmt.Print("Новый ID единицы измерения по умолчанию (оставьте пустым для сохранения): ")
	defStr := readLine(reader)
	defaultUnitID := product.DefaultUnitID
	if defStr != "" {
		defID, err := strconv.Atoi(defStr)
		if err == nil {
			defaultUnitID = &defID
		} else {
			fmt.Println("Неверный ID, оставляем как есть.")
		}
	}

	req := models.UpdateProductRequest{
		ID:             productID,
		Name:           name,
		ClassNodeID:    classID,
		UnitType:       unitType,
		WeightPerMeter: weightPerMeter,
		PieceLength:    pieceLength,
		DefaultUnitID:  defaultUnitID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	product, err = repo.UpdateProduct(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка обновления: %v\n", err)
		return
	}
	fmt.Printf("Изделие %d обновлено.\n", product.ID)
}

func deleteProduct(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID изделия для удаления: ")
	idStr := readLine(reader)
	productID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	product, err := repo.GetProduct(ctx, productID)
	if err != nil {
		fmt.Printf("Ошибка получения изделия: %v\n", err)
		return
	}
	classID := product.ClassNodeID
	err = repo.DeleteProduct(ctx, productID)
	if err != nil {
		fmt.Printf("Ошибка удаления: %v\n", err)
		return
	}
	fmt.Println("Изделие удалено.")
	products, err := repo.GetProductsByClass(ctx, classID)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось проверить наличие других изделий: %v\n", err)
		return
	}
	if len(products) == 0 {
		err = repo.UpdateNodeIsTerminal(ctx, classID, nil)
		if err != nil {
			fmt.Printf("Предупреждение: не удалось сбросить статус класса: %v\n", err)
		} else {
			fmt.Println("Класс больше не терминальный (теперь может принимать дочерние метаклассы).")
		}
	}
}

func printProduct(product *models.Product) {
	fmt.Printf("ID: %d, Название: %s, Класс ID: %d\n", product.ID, product.Name, product.ClassNodeID)
	if product.UnitType != nil {
		fmt.Printf("  Тип единицы: %s\n", *product.UnitType)
	}
	if product.WeightPerMeter != nil {
		fmt.Printf("  Вес погонного метра: %.2f т/м\n", *product.WeightPerMeter)
	}
	if product.PieceLength != nil {
		fmt.Printf("  Длина штуки: %.2f м\n", *product.PieceLength)
	}
	if product.DefaultUnitID != nil {
		fmt.Printf("  Единица по умолчанию ID: %d\n", *product.DefaultUnitID)
	}
}

func showProductParameters(repo repository.Repository, productID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	values, err := repo.GetParameterValuesForProduct(ctx, productID)
	if err != nil {
		fmt.Printf("Ошибка получения параметров: %v\n", err)
		return
	}
	if len(values) == 0 {
		fmt.Println("  нет значений")
		return
	}
	for _, v := range values {
		param, err := repo.GetParameterDefinition(ctx, v.ParamDefID)
		if err != nil {
			continue
		}
		if param.ParameterType == "number" && v.ValueNumeric != nil {
			unitStr := ""
			if param.UnitID != nil {
				unit, err := repo.GetUnit(ctx, *param.UnitID)
				if err == nil {
					unitStr = fmt.Sprintf(" %s", unit.Name)
				}
			}
			fmt.Printf("  %s: %.2f%s\n", param.Name, *v.ValueNumeric, unitStr)
		} else if param.ParameterType == "enum" && v.ValueEnumID != nil {
			enumVal, err := repo.GetEnumValue(ctx, *v.ValueEnumID)
			if err == nil {
				fmt.Printf("  %s: %s\n", param.Name, enumVal.Value)
			}
		}
	}
}
