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
		fmt.Println("3. Операции с перечислениями")
		fmt.Println("4. Выход")
		fmt.Print("Выбор: ")

		choice := readLine(reader)
		switch choice {
		case "1":
			nodeMenu(repo, reader)
		case "2":
			unitMenu(repo, reader)
		case "3":
			enumMenu(repo, reader)
		case "4":
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
		fmt.Println("11. Показать все терминальные метаклассы поддерева")
		fmt.Println("12. Главное меню")
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
			showAllTerminalDesc(ctx, repo, reader)
		case "12":
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

func createNode(ctx context.Context, repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите имя узла: ")
	name := readLine(reader)

	fmt.Print("Тип узла (metaclass/leaf): ")
	typeStr := readLine(reader)
	nodeType := models.NodeType(typeStr)

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

	var unitType *string
	var weightPerMeter *float64
	var pieceLength *float64
	var defaultUnitID *int

	if nodeType == models.TypeLeaf {

		fmt.Print("Тип единицы (mass/length/piece): ")
		ut := strings.ToLower(readLine(reader))
		if ut == "mass" || ut == "length" || ut == "piece" {
			unitType = &ut
		} else {
			fmt.Println("Неверный тип, оставляем пустым.")
		}
		fmt.Print("Вес погонного метра (т/м): ")
		wStr := readLine(reader)
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
		if defStr != "" {
			defID, err := strconv.Atoi(defStr)
			if err == nil {
				defaultUnitID = &defID
			} else {
				fmt.Println("Неверный ID, оставляем пустым.")
			}
		}
	}

	req := models.CreateNodeRequest{
		Name:           name,
		ParentID:       parentID,
		NodeType:       nodeType,
		IsTerminal:     isTerminal,
		UnitID:         unitID,
		SortOrder:      sortOrder,
		UnitType:       unitType,
		WeightPerMeter: weightPerMeter,
		PieceLength:    pieceLength,
		DefaultUnitID:  defaultUnitID,
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
	description := readLine(reader)
	if description == "" {
		description = *enum.Description
	}
	descr := &description
	req := models.UpdateEnumRequest{
		ID:          enum.ID,
		Name:        name,
		Description: descr,
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
	rolled := ""
	if node.UnitType != nil {
		rolled = fmt.Sprintf(", тип: %s", *node.UnitType)
		if node.WeightPerMeter != nil {
			rolled += fmt.Sprintf(", вес/м: %g", *node.WeightPerMeter)
		}
		if node.PieceLength != nil {
			rolled += fmt.Sprintf(", длина/шт: %g", *node.PieceLength)
		}
		if node.DefaultUnitID != nil {
			rolled += fmt.Sprintf(", ед.по умолч.: %d", *node.DefaultUnitID)
		}
	}
	fmt.Printf("ID: %d, название: %s, тип: %s%s%s, порядок: %d%s\n",
		node.ID, node.Name, node.NodeType, term, unit, node.SortOrder, rolled)
}

func readID(reader *bufio.Reader) *int {
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
