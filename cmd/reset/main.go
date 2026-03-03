package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// StructInfo содержит информацию о структуре, для которой нужно сгенерировать Reset()
type StructInfo struct {
	Name     string
	Package  string
	FilePath string
	Fields   []FieldInfo
}

// FieldInfo содержит информацию о поле структуры
type FieldInfo struct {
	Name     string
	Type     string
	IsPtr    bool
	IsSlice  bool
	IsMap    bool
	IsStruct bool
	Tag      string
}

func main() {
	// Автоматически находим корневую директорию проекта (ищем go.mod)
	// Начинаем с директории, где находится генератор
	execPath, err := os.Executable()
	if err != nil {
		// Если не удалось получить путь к исполняемому файлу, используем текущую директорию
		execPath = "."
	}

	// Получаем директорию генератора
	generatorDir := filepath.Dir(execPath)
	if generatorDir == "." {
		// Если запускаем через go run, получаем рабочую директорию
		generatorDir, _ = os.Getwd()
	}

	// Ищем go.mod, поднимаясь вверх по директориям
	rootDir := findProjectRoot(generatorDir)
	if rootDir == "" {
		// Если go.mod не найден, используем текущую рабочую директорию
		rootDir, _ = os.Getwd()
	}

	// Получаем абсолютный путь
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения абсолютного пути: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Сканирование пакетов в директории: %s\n", absRoot)

	// Сканируем все пакеты и находим структуры с комментарием // generate:reset
	structsToGenerate := make(map[string][]StructInfo) // ключ - путь пакета

	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории, начинающиеся с точки, vendor, и cmd (чтобы не сканировать сам генератор)
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			if info.Name() == "vendor" || info.Name() == "node_modules" || info.Name() == "cmd" {
				return filepath.SkipDir
			}
			return nil
		}

		// Обрабатываем только .go файлы (кроме уже сгенерированных)
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".gen.go") {
			return nil
		}

		structs, err := parseFileForStructs(path, absRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка парсинга файла %s: %v\n", path, err)
			return nil // Продолжаем обработку других файлов
		}

		if len(structs) > 0 {
			// Определяем путь пакета относительно корня
			relPath, _ := filepath.Rel(absRoot, filepath.Dir(path))
			pkgPath := relPath
			if pkgPath == "." {
				pkgPath = ""
			}
			structsToGenerate[pkgPath] = append(structsToGenerate[pkgPath], structs...)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка сканирования директорий: %v\n", err)
		os.Exit(1)
	}

	// Генерируем методы Reset() для каждого пакета
	for pkgPath, structs := range structsToGenerate {
		if err := generateResetMethods(absRoot, pkgPath, structs); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка генерации для пакета %s: %v\n", pkgPath, err)
		}
	}

	fmt.Println("Генерация завершена")
}

// parseFileForStructs парсит файл и находит структуры с комментарием // generate:reset
func parseFileForStructs(filePath, rootDir string) ([]StructInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var structs []StructInfo
	pkgName := node.Name.Name

	// Проходим по всем декларациям в файле
	ast.Inspect(node, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		// Проверяем комментарии перед декларацией
		if genDecl.Doc == nil {
			return true
		}

		hasGenerateReset := false
		for _, comment := range genDecl.Doc.List {
			if strings.Contains(comment.Text, "generate:reset") {
				hasGenerateReset = true
				break
			}
		}

		if !hasGenerateReset {
			return true
		}

		// Ищем type declarations
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Найдена структура с комментарием // generate:reset
			structName := typeSpec.Name.Name
			fields := parseStructFields(structType)

			relPath, _ := filepath.Rel(rootDir, filepath.Dir(filePath))
			pkgPath := relPath
			if pkgPath == "." {
				pkgPath = ""
			}

			structs = append(structs, StructInfo{
				Name:     structName,
				Package:  pkgName,
				FilePath: filepath.Dir(filePath),
				Fields:   fields,
			})
		}

		return true
	})

	return structs, nil
}

// parseStructFields извлекает информацию о полях структуры
func parseStructFields(structType *ast.StructType) []FieldInfo {
	var fields []FieldInfo

	for _, field := range structType.Fields.List {
		fieldType := getTypeString(field.Type)
		isPtr := isPointerType(field.Type)
		baseType := fieldType
		if isPtr {
			baseType = strings.TrimPrefix(baseType, "*")
		}

		fieldInfo := FieldInfo{
			Type:     fieldType,
			IsPtr:    isPtr,
			IsSlice:  isSliceType(field.Type),
			IsMap:    isMapType(field.Type),
			IsStruct: isStructType(field.Type) || (!isPrimitiveType(baseType) && !isPtr && !isSliceType(field.Type) && !isMapType(field.Type)),
		}

		if field.Tag != nil {
			fieldInfo.Tag = field.Tag.Value
		}

		// Обрабатываем имена полей
		if len(field.Names) == 0 {
			// Встроенное поле - пропускаем, так как для них Reset() будет вызван через встроенный тип
			continue
		} else {
			for _, name := range field.Names {
				if name.IsExported() || name.Name != "_" {
					fieldInfo.Name = name.Name
					fields = append(fields, fieldInfo)
				}
			}
		}
	}

	return fields
}

// getTypeString возвращает строковое представление типа
func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + getTypeString(t.Elt)
		}
		return "[" + getTypeString(t.Len) + "]" + getTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeString(t.Key) + "]" + getTypeString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ChanType:
		return "chan " + getTypeString(t.Value)
	default:
		return "unknown"
	}
}

// isPointerType проверяет, является ли тип указателем
func isPointerType(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

// isSliceType проверяет, является ли тип слайсом
func isSliceType(expr ast.Expr) bool {
	arrType, ok := expr.(*ast.ArrayType)
	if !ok {
		return false
	}
	return arrType.Len == nil
}

// isMapType проверяет, является ли тип мапой
func isMapType(expr ast.Expr) bool {
	_, ok := expr.(*ast.MapType)
	return ok
}

// isStructType проверяет, является ли тип структурой
func isStructType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		// Не можем точно определить без дополнительной информации
		return false
	case *ast.SelectorExpr:
		// Пакет.Тип - возможно структура
		return true
	case *ast.StarExpr:
		return isStructType(t.X)
	default:
		return false
	}
}

// generateResetMethods генерирует методы Reset() для всех структур пакета
func generateResetMethods(rootDir, pkgPath string, structs []StructInfo) error {
	if len(structs) == 0 {
		return nil
	}

	// Определяем директорию пакета
	var pkgDir string
	if pkgPath == "" {
		pkgDir = rootDir
	} else {
		pkgDir = filepath.Join(rootDir, pkgPath)
	}

	// Создаем содержимое файла reset.gen.go
	var sb strings.Builder
	sb.WriteString("// Code generated by cmd/reset/main.go. DO NOT EDIT.\n\n")
	sb.WriteString("package " + structs[0].Package + "\n\n")

	// Генерируем метод Reset() для каждой структуры
	for _, s := range structs {
		sb.WriteString(fmt.Sprintf("// Reset сбрасывает состояние структуры %s к начальным значениям\n", s.Name))
		sb.WriteString(fmt.Sprintf("func (s *%s) Reset() {\n", s.Name))

		for _, field := range s.Fields {
			fieldName := field.Name
			resetCode := generateFieldReset(fieldName, field)
			if resetCode != "" {
				sb.WriteString("\t" + resetCode + "\n")
			}
		}

		sb.WriteString("}\n\n")
	}

	// Форматируем код
	formatted, err := format.Source([]byte(sb.String()))
	if err != nil {
		// Если форматирование не удалось, используем исходный код
		formatted = []byte(sb.String())
		fmt.Fprintf(os.Stderr, "Предупреждение: не удалось отформатировать код для пакета %s: %v\n", pkgPath, err)
	}

	// Записываем в файл reset.gen.go
	outputPath := filepath.Join(pkgDir, "reset.gen.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла %s: %w", outputPath, err)
	}

	fmt.Printf("Сгенерирован файл: %s (структур: %d)\n", outputPath, len(structs))
	return nil
}

// generateFieldReset генерирует код для сброса поля
func generateFieldReset(fieldName string, field FieldInfo) string {
	if field.IsPtr {
		// Для указателей проверяем на nil и сбрасываем значение, на которое они указывают
		baseType := strings.TrimPrefix(field.Type, "*")
		// Создаем новую FieldInfo для базового типа
		baseField := FieldInfo{
			Type:     baseType,
			IsPtr:    false,
			IsSlice:  strings.HasPrefix(baseType, "[]"),
			IsMap:    strings.HasPrefix(baseType, "map["),
			IsStruct: field.IsStruct || (!isPrimitiveType(baseType) && !strings.HasPrefix(baseType, "[]") && !strings.HasPrefix(baseType, "map[")),
		}

		// Для указателей на примитивы - сбрасываем значение к zero value
		if isPrimitiveType(baseType) {
			resetCode := generateValueReset("*s."+fieldName, baseType, baseField)
			return fmt.Sprintf("if s.%s != nil {\n\t\t%s\n\t}", fieldName, resetCode)
		}

		// Для указателей на структуры - вызываем Reset() если есть, иначе сбрасываем к zero value
		if baseField.IsStruct {
			// Пытаемся вызвать Reset() на разыменованном указателе
			if strings.Contains(baseType, ".") {
				// Тип из другого пакета
				return fmt.Sprintf("if s.%s != nil {\n\t\tif r, ok := interface{}(s.%s).(interface{ Reset() }); ok {\n\t\t\tr.Reset()\n\t\t} else {\n\t\t\t*s.%s = %s{}\n\t\t}\n\t}", fieldName, fieldName, fieldName, baseType)
			}
			// Локальный тип
			return fmt.Sprintf("if s.%s != nil {\n\t\tif r, ok := interface{}(s.%s).(interface{ Reset() }); ok {\n\t\t\tr.Reset()\n\t\t} else {\n\t\t\t*s.%s = %s{}\n\t\t}\n\t}", fieldName, fieldName, fieldName, baseType)
		}

		// Для указателей на слайсы - обрезаем по длине
		if baseField.IsSlice {
			return fmt.Sprintf("if s.%s != nil {\n\t\t*s.%s = (*s.%s)[:0]\n\t}", fieldName, fieldName, fieldName)
		}

		// Для указателей на мапы - очищаем
		if baseField.IsMap {
			return fmt.Sprintf("if s.%s != nil {\n\t\tclear(*s.%s)\n\t}", fieldName, fieldName)
		}

		// Для остальных - сбрасываем к zero value
		resetCode := generateValueReset("*s."+fieldName, baseType, baseField)
		return fmt.Sprintf("if s.%s != nil {\n\t\t%s\n\t}", fieldName, resetCode)
	}

	if field.IsSlice {
		// Слайсы обрезаем по длине, но не зануляем
		return fmt.Sprintf("s.%s = s.%s[:0]", fieldName, fieldName)
	}

	if field.IsMap {
		// Мапы очищаем
		return fmt.Sprintf("clear(s.%s)", fieldName)
	}

	// Для обычных типов и структур
	return generateValueReset("s."+fieldName, field.Type, field)
}

// generateValueReset генерирует код для сброса значения
func generateValueReset(target, typeName string, field FieldInfo) string {
	// Примитивные типы
	switch typeName {
	case "int", "int8", "int16", "int32", "int64":
		return fmt.Sprintf("%s = 0", target)
	case "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return fmt.Sprintf("%s = 0", target)
	case "float32", "float64":
		return fmt.Sprintf("%s = 0", target)
	case "complex64", "complex128":
		return fmt.Sprintf("%s = 0", target)
	case "string":
		return fmt.Sprintf("%s = \"\"", target)
	case "bool":
		return fmt.Sprintf("%s = false", target)
	case "byte", "rune":
		return fmt.Sprintf("%s = 0", target)
	}

	// Для слайсов (если это не было обработано ранее)
	if strings.HasPrefix(typeName, "[]") {
		return fmt.Sprintf("%s = %s[:0]", target, target)
	}

	// Для мап (если это не было обработано ранее)
	if strings.HasPrefix(typeName, "map[") {
		return fmt.Sprintf("clear(%s)", target)
	}

	// Для структур проверяем, есть ли метод Reset()
	// Если это структура (не примитив), пытаемся вызвать Reset()
	if field.IsStruct || (!field.IsPtr && !field.IsSlice && !field.IsMap && !isPrimitiveType(typeName)) {
		// Проверяем, является ли это встроенным типом или типом из другого пакета
		if strings.Contains(typeName, ".") {
			// Тип из другого пакета - пытаемся вызвать Reset(), если он есть
			return fmt.Sprintf("if r, ok := interface{}(&%s).(interface{ Reset() }); ok { r.Reset() } else { %s = %s{} }", target, target, typeName)
		}
		// Локальный тип - предполагаем, что может быть Reset()
		// Генерируем вызов Reset(), если метод существует
		return fmt.Sprintf("if r, ok := interface{}(&%s).(interface{ Reset() }); ok { r.Reset() } else { %s = %s{} }", target, target, typeName)
	}

	// Для неизвестных типов используем zero value
	return fmt.Sprintf("%s = %s{}", target, typeName)
}

// isPrimitiveType проверяет, является ли тип примитивным
func isPrimitiveType(typeName string) bool {
	primitives := []string{
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64",
		"complex64", "complex128",
		"string", "bool", "byte", "rune",
	}
	for _, p := range primitives {
		if typeName == p {
			return true
		}
	}
	return false
}

// findProjectRoot ищет корневую директорию проекта, поднимаясь вверх и ища go.mod
func findProjectRoot(startDir string) string {
	for dir := startDir; ; {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		// filepath.Dir возвращает родительскую директорию для данного пути
		// Например: filepath.Dir("/a/b/c") вернет "/a/b"
		parent := filepath.Dir(dir)
		if parent == dir {
			// Достигли корня файловой системы (например, "/" на Unix или "C:\" на Windows)
			break
		}
		dir = parent
	}
	return ""
}
