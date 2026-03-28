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

func paramMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Управление параметрами ---")
		fmt.Println("1. Управление определениями параметров классов")
		fmt.Println("2. Управление значениями параметров изделий")
		fmt.Println("3. Поиск изделий по параметрам")
		fmt.Println("4. Назад в главное меню")
		fmt.Print("Выбор: ")
		choice := readLine(reader)

		switch choice {
		case "1":
			paramDefMenu(repo, reader)
		case "2":
			paramValueMenu(repo, reader)
		case "3":
			searchProductsByParams(repo, reader)
		case "4":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func paramDefMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Определения параметров классов ---")
		fmt.Println("1. Создать параметр для класса")
		fmt.Println("2. Просмотреть параметры класса (с наследованием)")
		fmt.Println("3. Редактировать параметр")
		fmt.Println("4. Удалить параметр")
		fmt.Println("5. Назад")
		fmt.Print("Выбор: ")
		choice := readLine(reader)

		switch choice {
		case "1":
			createParameterDefinition(repo, reader)
		case "2":
			listParameterDefinitionsForClass(repo, reader)
		case "3":
			updateParameterDefinition(repo, reader)
		case "4":
			deleteParameterDefinition(repo, reader)
		case "5":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func createParameterDefinition(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID класса (метакласса): ")
	classIDStr := readLine(reader)
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}

	fmt.Print("Введите название параметра: ")
	name := readLine(reader)
	if name == "" {
		fmt.Println("Название обязательно.")
		return
	}

	fmt.Print("Введите описание (необязательно): ")
	desc := readLine(reader)
	var description *string
	if desc != "" {
		description = &desc
	}

	fmt.Println("Выберите тип параметра:")
	fmt.Println("1. Числовой")
	fmt.Println("2. Перечисление")
	choice := readLine(reader)

	var paramType string
	var unitID *int
	var enumID *int

	switch choice {
	case "1":
		paramType = "number"
		ctxUnits, cancelUnits := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelUnits()
		units, err := repo.GetAllUnits(ctxUnits)
		if err == nil && len(units) > 0 {
			fmt.Println("Доступные единицы измерения:")
			for _, u := range units {
				fmt.Printf("  %d: %s (множитель: %g)\n", u.ID, u.Name, u.Multiplier)
			}
			fmt.Print("Введите ID единицы измерения (оставьте пустым, если не требуется): ")
			unitStr := readLine(reader)
			if unitStr != "" {
				uid, err := strconv.Atoi(unitStr)
				if err == nil {
					unitID = &uid
				}
			}
		}
	case "2":
		paramType = "enum"
		ctxEnums, cancelEnums := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelEnums()
		enums, err := repo.GetAllEnums(ctxEnums)
		if err != nil {
			fmt.Printf("Ошибка получения перечислений: %v\n", err)
			return
		}
		if len(enums) == 0 {
			fmt.Println("Сначала создайте хотя бы одно перечисление в меню перечислений.")
			return
		}
		fmt.Println("Доступные перечисления:")
		for _, e := range enums {
			fmt.Printf("  %d: %s (тип: %d)\n", e.ID, e.Name, e.TypeNodeID)
		}
		fmt.Print("Введите ID перечисления: ")
		enumStr := readLine(reader)
		enumIDInt, err := strconv.Atoi(enumStr)
		if err != nil {
			fmt.Println("Неверный ID.")
			return
		}
		enumID = &enumIDInt
	default:
		fmt.Println("Неверный выбор.")
		return
	}

	fmt.Print("Обязательный параметр? (yes/no): ")
	requiredStr := readLine(reader)
	isRequired := requiredStr == "yes"

	fmt.Print("Порядок сортировки (оставьте пустым для авто): ")
	orderStr := readLine(reader)
	var sortOrder *int
	if orderStr != "" {
		order, err := strconv.Atoi(orderStr)
		if err == nil {
			sortOrder = &order
		}
	}

	var constraints *models.ParameterConstraint
	if paramType == "number" {
		fmt.Print("Минимальное значение (оставьте пустым, если нет): ")
		minStr := readLine(reader)
		fmt.Print("Максимальное значение (оставьте пустым, если нет): ")
		maxStr := readLine(reader)
		if minStr != "" || maxStr != "" {
			var minVal, maxVal *float64
			if minStr != "" {
				min, err := strconv.ParseFloat(minStr, 64)
				if err == nil {
					minVal = &min
				}
			}
			if maxStr != "" {
				max, err := strconv.ParseFloat(maxStr, 64)
				if err == nil {
					maxVal = &max
				}
			}
			if minVal != nil && maxVal != nil && *minVal > *maxVal {
				fmt.Println("Минимальное значение не может быть больше максимального.")
				return
			}
			constraints = &models.ParameterConstraint{
				MinValue: minVal,
				MaxValue: maxVal,
			}
		}
	}

	req := models.CreateParameterDefinitionRequest{
		ClassNodeID:   classID,
		Name:          name,
		Description:   description,
		ParameterType: paramType,
		UnitID:        unitID,
		EnumID:        enumID,
		IsRequired:    isRequired,
		SortOrder:     sortOrder,
		Constraints:   constraints,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	param, err := repo.CreateParameterDefinition(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка создания параметра: %v\n", err)
		return
	}
	fmt.Printf("Параметр создан с ID: %d\n", param.ID)
}

func listParameterDefinitionsForClass(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID класса (метакласса): ")
	idStr := readLine(reader)
	classID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	params, err := repo.GetParameterDefinitionsForClass(ctx, classID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	if len(params) == 0 {
		fmt.Println("У этого класса нет параметров.")
		return
	}

	fmt.Println("Параметры (с учётом наследования):")
	for _, p := range params {
		unitStr := ""
		if p.UnitID != nil {
			unitStr = fmt.Sprintf(", ед.изм: %d", *p.UnitID)
		}
		enumStr := ""
		if p.EnumID != nil {
			enumStr = fmt.Sprintf(", перечисление: %d", *p.EnumID)
		}
		reqStr := ""
		if p.IsRequired {
			reqStr = ", обязательный"
		}
		fmt.Printf("ID: %d, Название: %s, Тип: %s%s%s%s\n",
			p.ID, p.Name, p.ParameterType, unitStr, enumStr, reqStr)

		if p.ParameterType == "number" {
			ctxConstraints, cancelConstraints := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancelConstraints()
			constraints, err := repo.GetParameterConstraints(ctxConstraints, p.ID)
			if err == nil && constraints != nil {
				constraintStr := ""
				if constraints.MinValue != nil && constraints.MaxValue != nil {
					constraintStr = fmt.Sprintf(" [%.2f .. %.2f]", *constraints.MinValue, *constraints.MaxValue)
				} else if constraints.MinValue != nil {
					constraintStr = fmt.Sprintf(" [>= %.2f]", *constraints.MinValue)
				} else if constraints.MaxValue != nil {
					constraintStr = fmt.Sprintf(" [<= %.2f]", *constraints.MaxValue)
				}
				if constraintStr != "" {
					fmt.Printf("    Ограничения: %s\n", constraintStr)
				}
			}
		}
	}
}

func updateParameterDefinition(repo repository.Repository, reader *bufio.Reader) {

	fmt.Print("Введите ID параметра для редактирования: ")
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}

	ctxGet, cancelGet := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelGet()
	param, err := repo.GetParameterDefinition(ctxGet, id)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("Текущее имя: %s\n", param.Name)
	fmt.Print("Новое имя (оставьте пустым для сохранения): ")
	name := readLine(reader)
	if name == "" {
		name = param.Name
	}

	fmt.Printf("Текущее описание: %v\n", param.Description)
	fmt.Print("Новое описание (оставьте пустым для сохранения): ")
	desc := readLine(reader)
	var description *string
	if desc != "" {
		description = &desc
	} else {
		description = param.Description
	}

	var unitID, enumID *int
	switch param.ParameterType {
	case "number":
		fmt.Printf("Текущая единица измерения ID: %v\n", param.UnitID)
		fmt.Print("Новый ID единицы измерения (оставьте пустым для сохранения): ")
		unitStr := readLine(reader)
		if unitStr != "" {
			uid, err := strconv.Atoi(unitStr)
			if err == nil {
				unitID = &uid
			}
		} else {
			unitID = param.UnitID
		}
	case "enum":
		fmt.Printf("Текущее перечисление ID: %v\n", param.EnumID)
		fmt.Print("Новый ID перечисления (оставьте пустым для сохранения): ")
		enumStr := readLine(reader)
		if enumStr != "" {
			eid, err := strconv.Atoi(enumStr)
			if err == nil {
				enumID = &eid
			}
		} else {
			enumID = param.EnumID
		}
	}

	fmt.Printf("Обязательный? (yes/no), текущее: %v\n", param.IsRequired)
	fmt.Print("Изменить? (yes/no/оставить пустым): ")
	reqStr := readLine(reader)
	isRequired := param.IsRequired
	switch reqStr {
	case "yes":
		isRequired = true
	case "no":
		isRequired = false
	default:
		isRequired = param.IsRequired
	}

	fmt.Print("Порядок сортировки (оставьте пустым для сохранения): ")
	orderStr := readLine(reader)
	var sortOrder *int
	if orderStr != "" {
		order, err := strconv.Atoi(orderStr)
		if err == nil {
			sortOrder = &order
		}
	} else {
		sortOrder = &param.SortOrder
	}

	req := models.UpdateParameterDefinitionRequest{
		ID:          param.ID,
		Name:        name,
		Description: description,
		UnitID:      unitID,
		EnumID:      enumID,
		IsRequired:  isRequired,
		SortOrder:   sortOrder,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = repo.UpdateParameterDefinition(ctx, req)
	if err != nil {
		fmt.Printf("Ошибка обновления: %v\n", err)
		return
	}
	fmt.Println("Параметр обновлён.")
}

func deleteParameterDefinition(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID параметра для удаления: ")
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	err = repo.DeleteParameterDefinition(ctx, id)
	if err != nil {
		fmt.Printf("Ошибка удаления: %v\n", err)
		return
	}
	fmt.Println("Параметр удалён.")
}

func paramValueMenu(repo repository.Repository, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- Значения параметров изделий ---")
		fmt.Println("1. Установить/обновить значение параметра для изделия")
		fmt.Println("2. Просмотреть значения параметров изделия")
		fmt.Println("3. Удалить значение параметра")
		fmt.Println("4. Назад")
		fmt.Print("Выбор: ")
		choice := readLine(reader)

		switch choice {
		case "1":
			setParameterValue(repo, reader)
		case "2":
			getParameterValuesForProduct(repo, reader)
		case "3":
			deleteParameterValue(repo, reader)
		case "4":
			return
		default:
			fmt.Println("Неверный выбор.")
		}
	}
}

func setParameterValue(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID изделия: ")
	productIDStr := readLine(reader)
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}

	ctxGet, cancelGet := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelGet()

	product, err := repo.GetProduct(ctxGet, productID)
	if err != nil {
		fmt.Printf("Ошибка получения продукта: %v\n", err)
		return
	}
	classID := product.ClassNodeID

	params, err := repo.GetParameterDefinitionsForClass(ctxGet, classID)
	if err != nil {
		fmt.Printf("Ошибка получения параметров: %v\n", err)
		return
	}
	if len(params) == 0 {
		fmt.Println("У этого класса нет параметров.")
		return
	}

	fmt.Println("Доступные параметры:")
	for _, p := range params {
		fmt.Printf("  ID: %d, Название: %s, Тип: %s\n", p.ID, p.Name, p.ParameterType)
	}
	fmt.Print("Введите ID параметра: ")
	paramIDStr := readLine(reader)
	paramID, err := strconv.Atoi(paramIDStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}

	var valueNumeric *float64
	var valueEnumID *int

	var targetParam *models.ParameterDefinition
	for _, p := range params {
		if p.ID == paramID {
			targetParam = p
			break
		}
	}
	if targetParam == nil {
		fmt.Println("Параметр не найден.")
		return
	}

	switch targetParam.ParameterType {
	case "number":
		fmt.Print("Введите числовое значение: ")
		numStr := readLine(reader)
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			fmt.Println("Неверное число.")
			return
		}
		valueNumeric = &num
	case "enum":
		if targetParam.EnumID == nil {
			fmt.Println("Ошибка: параметр перечисление не привязан к перечислению.")
			return
		}
		ctxEnum, cancelEnum := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelEnum()
		values, err := repo.GetEnumValues(ctxEnum, *targetParam.EnumID)
		if err != nil {
			fmt.Printf("Ошибка получения значений: %v\n", err)
			return
		}
		if len(values) == 0 {
			fmt.Println("В перечислении нет значений.")
			return
		}
		fmt.Println("Доступные значения:")
		for _, v := range values {
			fmt.Printf("  ID: %d, Значение: %s\n", v.ID, v.Value)
		}
		fmt.Print("Введите ID значения: ")
		valIDStr := readLine(reader)
		valID, err := strconv.Atoi(valIDStr)
		if err != nil {
			fmt.Println("Неверный ID.")
			return
		}
		valueEnumID = &valID
	}

	req := models.CreateParameterValueRequest{
		ProductID:    productID,
		ParamDefID:   paramID,
		ValueNumeric: valueNumeric,
		ValueEnumID:  valueEnumID,
	}

	ctxSet, cancelSet := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelSet()
	pv, err := repo.SetParameterValue(ctxSet, req)
	if err != nil {
		fmt.Printf("Ошибка установки значения: %v\n", err)
		return
	}
	fmt.Printf("Значение установлено (ID: %d)\n", pv.ID)
}

func getParameterValuesForProduct(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID изделия (листа): ")
	productIDStr := readLine(reader)
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}
	ctxGet, cancelGet := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelGet()
	values, err := repo.GetParameterValuesForProduct(ctxGet, productID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	if len(values) == 0 {
		fmt.Println("Для этого изделия не заданы параметры.")
		return
	}

	fmt.Println("Значения параметров:")
	for _, v := range values {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
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

func deleteParameterValue(repo repository.Repository, reader *bufio.Reader) {
	fmt.Print("Введите ID значения параметра (из таблицы parameter_values): ")
	idStr := readLine(reader)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Неверный ID.")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	err = repo.DeleteParameterValue(ctx, id)
	if err != nil {
		fmt.Printf("Ошибка удаления: %v\n", err)
		return
	}
	fmt.Println("Значение удалено.")
}

func searchProductsByParams(repo repository.Repository, reader *bufio.Reader) {
	readLine(reader)
	results, err := repo.FindProductsByParameters(context.Background(), 0, nil)
	if err != nil {
		fmt.Printf("Ошибка поиска: %v\n", err)
		return
	}
	fmt.Printf("Функция поиска изделий по параметрам пока не реализована. Найдено изделий: %d\n", len(results))
}
