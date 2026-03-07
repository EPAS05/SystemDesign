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

func nodeMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Операции с узлами ---")
		fmt.Println("1. Создать узел")
		fmt.Println("2. Получить информацию об узле")
		fmt.Println("3. Список потомков")
		fmt.Println("4. Все потомки")
		fmt.Println("5. Все родители")
		fmt.Println("6. Показать родителя")
		fmt.Println("7. Сменить родителя узла")
		fmt.Println("8. Поменять имя узла")
		fmt.Println("9. Изменить порядок вывода")
		fmt.Println("10. Удалить узел")
		fmt.Println("11. Показать все терминальные метаклассы поддерева")
		fmt.Println("12. Главное меню")
		fmt.Print("Выбор: ")

		choice := readLine(reader)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		switch choice {
		case "1":
			createNode(ctx, repo, reader)
		case "2":
			getNode(ctx, repo, reader)
		case "3":
			listChildren(ctx, repo, reader)
		case "4":
			listDescendants(ctx, repo, reader)
		case "5":
			listAncestors(ctx, repo, reader)
		case "6":
			getParent(ctx, repo, reader)
		case "7":
			moveNode(ctx, repo, reader)
		case "8":
			renameNode(ctx, repo, reader)
		case "9":
			setNodeOrder(ctx, repo, reader)
		case "10":
			deleteNode(ctx, repo, reader)
		case "11":
			showAllTerminalDesc(ctx, repo, reader)
		case "12":
			return
		default:
			fmt.Println("Неправильный выбор. Попробуйте снова")
		}
	}
}

func createNode(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите имя узла: ")
	name := readLine(reader)

	fmt.Print("Тип узла (metaclass/leaf): ")
	typeStr := readLine(reader)
	nodeType := models.NodeType(typeStr)

	if nodeType == models.TypeEnum {
		fmt.Println("Перечисления создаются через меню перечислений.")
		return
	}

	var isTerminal *bool

	fmt.Print("ID родителя (Оставьте пустым для корня): ")
	parentStr := readLine(reader)
	var parentID *int
	if parentStr != "" {
		pid, err := strconv.Atoi(parentStr)
		if err != nil {
			fmt.Println("Неправильный ID")
			return
		}
		parentID = &pid
	}

	units, err := repo.GetAllUnits(ctx)
	if err != nil {
		fmt.Printf("Ошибка получения ЕИ: %v\n", err)
	} else if len(units) > 0 {
		fmt.Println("Доступные ЕИ:")
		for _, u := range units {
			fmt.Printf("  %d: %s (множитель: %g)\n", u.ID, u.Name, u.Multiplier)
		}
	}

	fmt.Print("ID ЕИ (оставьте пустым для наследования от родителя при его наличии): ")
	unitStr := readLine(reader)
	var unitID *int
	if unitStr != "" {
		uid, err := strconv.Atoi(unitStr)
		if err != nil {
			fmt.Println("Неправильный ID")
			return
		}
		unitID = &uid
	}

	fmt.Print("Порядок сортировки (оставьте пустым для авто): ")
	orderStr := readLine(reader)
	var sortOrder *int
	if orderStr != "" {
		order, err := strconv.Atoi(orderStr)
		if err != nil {
			fmt.Println("Неправильный порядок (введите число)")
			return
		}
		sortOrder = &order
	}

	var objectType *string
	var objectID *int

	if nodeType == models.TypeLeaf {
		fmt.Print("Тип единицы (mass/length/piece): ")
		ut := strings.ToLower(readLine(reader))
		var unitType *string
		if ut == "mass" || ut == "length" || ut == "piece" {
			unitType = &ut
		} else {
			fmt.Println("Неверный тип, оставляем пустым.")
		}

		fmt.Print("Вес погонного метра (т/м): ")
		wStr := readLine(reader)
		var weightPerMeter *float64
		if wStr != "" {
			w, err := strconv.ParseFloat(wStr, 64)
			if err == nil {
				weightPerMeter = &w
			} else {
				fmt.Println("Неверное число, оставляем пустым.")
			}
		}

		fmt.Print("Длина одной штуки (м): ")
		pStr := readLine(reader)
		var pieceLength *float64
		if pStr != "" {
			p, err := strconv.ParseFloat(pStr, 64)
			if err == nil {
				pieceLength = &p
			} else {
				fmt.Println("Неверное число, оставляем пустым.")
			}
		}

		fmt.Print("ID единицы измерения по умолчанию (оставьте пустым, если не задана): ")
		defStr := readLine(reader)
		var defaultUnitID *int
		if defStr != "" {
			defID, err := strconv.Atoi(defStr)
			if err == nil {
				defaultUnitID = &defID
			} else {
				fmt.Println("Неверный ID, оставляем пустым.")
			}
		}

		prodReq := models.CreateProductRequest{
			UnitType:       unitType,
			WeightPerMeter: weightPerMeter,
			PieceLength:    pieceLength,
			DefaultUnitID:  defaultUnitID,
		}
		product, err := repo.CreateProduct(ctx, prodReq)
		if err != nil {
			fmt.Printf("Ошибка создания продукта: %v\n", err)
			return
		}
		objectType := new(string)
		*objectType = "product"
		objectID = &product.ID
		fmt.Printf("Продукт создан с ID: %d\n", product.ID)
	}

	req := models.CreateNodeRequest{
		Name:       name,
		ParentID:   parentID,
		NodeType:   nodeType,
		IsTerminal: isTerminal,
		UnitID:     unitID,
		SortOrder:  sortOrder,
		ObjectType: objectType,
		ObjectID:   objectID,
	}

	node, err := repo.CreateNode(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка создания узла: %v\n", err)
		return
	}
	fmt.Printf("Узел создан с ID: %d\n", node.ID)
}

func getNode(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	node, err := repo.GetNode(ctx, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	printNode(node)
}

func listChildren(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	children, err := repo.GetChildren(ctx, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(children) == 0 {
		fmt.Println("Детей нет.")
		return
	}
	fmt.Println("Дети:")
	for _, node := range children {
		printNode(node)
	}
}

func listDescendants(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	descendants, err := repo.GetAllDescendants(ctx, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(descendants) == 0 {
		fmt.Println("Нет потомков.")
		return
	}
	fmt.Println("Потомки:")
	for _, node := range descendants {
		printNode(node)
	}
}

func listAncestors(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	ancestors, err := repo.GetAllAncestors(ctx, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(ancestors) == 0 {
		fmt.Println("Нет предшественников.")
		return
	}
	fmt.Println("Предшественники:")
	for _, node := range ancestors {
		printNode(node)
	}
}

func getParent(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	parent, err := repo.GetParent(ctx, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if parent == nil {
		fmt.Println("Нет родителя.")
		return
	}
	printNode(parent)
}

func moveNode(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	fmt.Print("ID нового родителя (Пустой для корня): ")
	parentStr := readLine(reader)
	var newParentID *int
	if parentStr != "" {
		pid, err := strconv.Atoi(parentStr)
		if err != nil {
			fmt.Println("Неправильный ID.")
			return
		}
		newParentID = &pid
	}
	req := models.SetParentRequest{
		NodeId:      *id,
		NewParentID: newParentID,
	}
	err := repo.SetParent(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Узел подвинут.")
}

func renameNode(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	fmt.Print("Новое имя: ")
	name := readLine(reader)
	req := models.SetNameRequest{
		NodeId: *id,
		Name:   name,
	}
	err := repo.SetName(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Узел переименован.")
}

func setNodeOrder(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	fmt.Print("Новый порядок: ")
	orderStr := readLine(reader)
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		fmt.Println("Неправильно (введите число).")
		return
	}
	err = repo.SetNodeOrder(ctx, *id, order)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Порядок изменен.")
}

func deleteNode(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
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
	err := repo.DeleteNode(ctx, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Узел удален.")
}

func showAllTerminalDesc(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	terminals, err := repo.GetAllTerminalDescendants(ctx, *id)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	if len(terminals) == 0 {
		fmt.Println("Нет терминальных метаклассов в поддереве.")
		return
	}
	fmt.Println("Терминальные метаклассы:")
	for _, node := range terminals {
		printNode(node)
	}
}

func printNode(node *models.Node) {
	term := ""
	if node.NodeType == models.TypeMetaclass && node.IsTerminal != nil {
		term = fmt.Sprintf(", Терминальный: %v", *node.IsTerminal)
	}
	unit := ""
	if node.UnitID != nil {
		unit = fmt.Sprintf(", ID ЕИ: %d", *node.UnitID)
	}
	obj := ""
	if node.ObjectType != nil && node.ObjectID != nil {
		obj = fmt.Sprintf(", объект: %s:%d", *node.ObjectType, *node.ObjectID)
	}
	fmt.Printf("ID: %d, название: %s, тип: %s%s%s, порядок: %d%s\n",
		node.ID, node.Name, node.NodeType, term, unit, node.SortOrder, obj)
}
