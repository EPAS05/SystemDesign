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

const PRODUCT_NODE_ROOT_ID = 2

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

		switch choice {
		case "1":
			createNode(repo, reader)
		case "2":
			getNode(repo, reader)
		case "3":
			listChildren(repo, reader)
		case "4":
			listDescendants(repo, reader)
		case "5":
			listAncestors(repo, reader)
		case "6":
			getParent(repo, reader)
		case "7":
			moveNode(repo, reader)
		case "8":
			renameNode(repo, reader)
		case "9":
			setNodeOrder(repo, reader)
		case "10":
			deleteNode(repo, reader)
		case "11":
			showAllTerminalDesc(repo, reader)
		case "12":
			return
		default:
			fmt.Println("Неправильный выбор. Попробуйте снова")
		}
	}
}

func createNode(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите имя узла: ")
	name := readLine(reader)

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
	} else {
		pid := PRODUCT_NODE_ROOT_ID
		parentID = &pid
	}

	ctxUnits, cancelUnits := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelUnits()
	units, err := repo.GetAllUnits(ctxUnits)
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

	req := models.CreateNodeRequest{
		Name:      name,
		ParentID:  parentID,
		UnitID:    unitID,
		SortOrder: sortOrder,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	node, err := repo.CreateNode(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка создания узла: %v\n", err)
		return
	}
	fmt.Printf("Узел создан с ID: %d\n", node.ID)
}

func getNode(repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	node, err := repo.GetNode(ctx, *id)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	printNode(node)
}

func listChildren(repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	children, err := repo.GetChildren(ctx, *id)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
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

func listDescendants(repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	descendants, err := repo.GetAllDescendants(ctx, *id)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
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

func listAncestors(repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ancestors, err := repo.GetAllAncestors(ctx, *id)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
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

func getParent(repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	parent, err := repo.GetParent(ctx, *id)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	if parent == nil {
		fmt.Println("Нет родителя.")
		return
	}
	printNode(parent)
}

func moveNode(repo repository.Repository, reader *bufio.Reader) {
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
	} else {
		pid := PRODUCT_NODE_ROOT_ID
		newParentID = &pid
	}
	req := models.SetParentRequest{
		NodeId:      *id,
		NewParentID: newParentID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := repo.SetParent(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	fmt.Println("Узел перемещён.")
}

func renameNode(repo repository.Repository, reader *bufio.Reader) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := repo.SetName(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	fmt.Println("Узел переименован.")
}

func setNodeOrder(repo repository.Repository, reader *bufio.Reader) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = repo.SetNodeOrder(ctx, *id, order)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	fmt.Println("Порядок изменён.")
}

func deleteNode(repo repository.Repository, reader *bufio.Reader) {
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
	err := repo.DeleteNode(ctx, *id)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	fmt.Println("Узел удалён.")
}

func showAllTerminalDesc(repo repository.Repository, reader *bufio.Reader) {
	id := readID(reader)
	if id == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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
	if node.IsTerminal != nil {
		term = fmt.Sprintf(", терминальный: %v", *node.IsTerminal)
	}
	unit := ""
	if node.UnitID != nil {
		unit = fmt.Sprintf(", ID ЕИ: %d", *node.UnitID)
	}
	fmt.Printf("ID: %d, название: %s, порядок: %d%s%s\n",
		node.ID, node.Name, node.SortOrder, term, unit)
}
