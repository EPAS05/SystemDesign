package main

import (
	"bufio"
	"classifier/internal/db"
	"classifier/internal/models"
	"classifier/internal/repository"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	connStr := getDBConnStr()
	database, err := db.NewConnection(connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	repo := repository.NewPostgresRepository(database)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== Главное меню ===")
		fmt.Println("1. Операции с узлами")
		fmt.Println("2. Операции с ЕИ")
		fmt.Println("3. Выход")
		fmt.Print("Выбор: ")

		choice := readLine(reader)
		switch choice {
		case "1":
			nodeMenu(repo, reader)
		case "2":
			unitMenu(repo, reader)
		case "3":
			fmt.Println("Завершение работы!")
			return
		default:
			fmt.Println("Неправильный выбор. Попробуйте снова")
		}
	}
}

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
		fmt.Println("11. Главное меню")
		fmt.Print("Выбор: ")

		choice := readLine(reader)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
			return
		default:
			fmt.Println("Неправильный выбор. Попробуйте снова")
		}
	}
}

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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

func createNode(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите имя узла: ")
	name := readLine(reader)

	fmt.Print("Тип узла (metaclass/leaf): ")
	typeStr := readLine(reader)
	nodeType := models.NodeType(typeStr)

	var isTerminal *bool
	if nodeType == models.TypeMetaclass {
		fmt.Print("Терминальный? (true/false): ")
		termStr := readLine(reader)
		term, err := strconv.ParseBool(termStr)
		if err != nil {
			fmt.Println("Неправильно. Используйте true или false.")
			return
		}
		isTerminal = &term
	}

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

	fmt.Print("Порядок сортировки (оставьте пустым для авто) ")
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
		Name:       name,
		ParentID:   parentID,
		NodeType:   nodeType,
		IsTerminal: isTerminal,
		UnitID:     unitID,
		SortOrder:  sortOrder,
	}

	node, err := repo.CreateNode(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка создания узла: %v\n", err)
		return
	}
	fmt.Printf("Узел создан с ID: %d\n", node.ID)
}

func getNode(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	id := readNodeID(reader)
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
	id := readNodeID(reader)
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
	id := readNodeID(reader)
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
	id := readNodeID(reader)
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
	id := readNodeID(reader)
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
	id := readNodeID(reader)
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
	id := readNodeID(reader)
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
	id := readNodeID(reader)
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
	id := readNodeID(reader)
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
		ID:         id,
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
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неправильный ID.")
		return
	}
	fmt.Print("Уверены? (yes/no): ")
	confirm := readLine(reader)
	if confirm != "yes" {
		fmt.Println("Удаление отменено.")
		return
	}
	err = repo.DeleteUnit(ctx, id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("ЕИ удалена.")
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
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
	fmt.Printf("ID: %d, название: %s, тип: %s%s%s, порядок: %d\n",
		node.ID, node.Name, node.NodeType, term, unit, node.SortOrder)
}

func readNodeID(reader *bufio.Reader) *int {
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

func getDBConnStr() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "classifier")
	password := getEnv("DB_PASSWORD", "secret")
	dbname := getEnv("DB_NAME", "classifier")
	sslmode := getEnv("DB_SSLMODE", "disable")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
