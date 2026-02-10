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

type StructInfo struct {
	Name     string
	Package  string
	FilePath string
	Fields   []FieldInfo
}

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
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	absRoot, err := resolveProjectRoot()
	if err != nil {
		return err
	}

	fmt.Printf("Сканирование пакетов в директории: %s\n", absRoot)

	structsToGenerate, err := scanStructs(absRoot)
	if err != nil {
		return err
	}

	generateForPackages(absRoot, structsToGenerate)

	fmt.Println("Генерация завершена")
	return nil
}

func resolveProjectRoot() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}

	generatorDir := filepath.Dir(execPath)
	if generatorDir == "." {
		generatorDir, _ = os.Getwd()
	}

	rootDir := findProjectRoot(generatorDir)
	if rootDir == "" {
		rootDir, _ = os.Getwd()
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("ошибка получения абсолютного пути: %w", err)
	}

	return absRoot, nil
}

func scanStructs(absRoot string) (map[string][]StructInfo, error) {
	structsToGenerate := make(map[string][]StructInfo)

	err := filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !isSourceFile(path) {
			return nil
		}

		structs, err := parseFileForStructs(path, absRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка парсинга файла %s: %v\n", path, err)
			return nil
		}

		if len(structs) == 0 {
			return nil
		}

		pkgPath := normalizePackagePath(absRoot, filepath.Dir(path))
		structsToGenerate[pkgPath] = append(structsToGenerate[pkgPath], structs...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка сканирования директорий: %w", err)
	}

	return structsToGenerate, nil
}

func shouldSkipDir(name string) bool {
	if strings.HasPrefix(name, ".") && name != "." {
		return true
	}

	switch name {
	case "vendor", "node_modules", "cmd":
		return true
	}

	return false
}

func isSourceFile(path string) bool {
	if !strings.HasSuffix(path, ".go") {
		return false
	}
	return !strings.HasSuffix(path, ".gen.go")
}

func normalizePackagePath(rootDir, fileDir string) string {
	relPath, _ := filepath.Rel(rootDir, fileDir)
	if relPath == "." {
		return ""
	}
	return relPath
}

func generateForPackages(absRoot string, structsToGenerate map[string][]StructInfo) {
	for pkgPath, structs := range structsToGenerate {
		if err := generateResetMethods(absRoot, pkgPath, structs); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка генерации для пакета %s: %v\n", pkgPath, err)
		}
	}
}

func parseFileForStructs(filePath, rootDir string) ([]StructInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var structs []StructInfo
	pkgName := node.Name.Name

	ast.Inspect(node, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

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

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

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

		if len(field.Names) == 0 {
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

func isPointerType(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

func isSliceType(expr ast.Expr) bool {
	arrType, ok := expr.(*ast.ArrayType)
	if !ok {
		return false
	}
	return arrType.Len == nil
}

func isMapType(expr ast.Expr) bool {
	_, ok := expr.(*ast.MapType)
	return ok
}

func isStructType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return false
	case *ast.SelectorExpr:
		return true
	case *ast.StarExpr:
		return isStructType(t.X)
	default:
		return false
	}
}

func generateResetMethods(rootDir, pkgPath string, structs []StructInfo) error {
	if len(structs) == 0 {
		return nil
	}

	var pkgDir string
	if pkgPath == "" {
		pkgDir = rootDir
	} else {
		pkgDir = filepath.Join(rootDir, pkgPath)
	}

	var sb strings.Builder
	sb.WriteString("// Code generated by cmd/reset/main.go. DO NOT EDIT.\n\n")
	sb.WriteString("package " + structs[0].Package + "\n\n")

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

	formatted, err := format.Source([]byte(sb.String()))
	if err != nil {
		formatted = []byte(sb.String())
		fmt.Fprintf(os.Stderr, "Предупреждение: не удалось отформатировать код для пакета %s: %v\n", pkgPath, err)
	}

	outputPath := filepath.Join(pkgDir, "reset.gen.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла %s: %w", outputPath, err)
	}

	fmt.Printf("Сгенерирован файл: %s (структур: %d)\n", outputPath, len(structs))
	return nil
}

func generateFieldReset(fieldName string, field FieldInfo) string {
	if field.IsPtr {
		baseType := strings.TrimPrefix(field.Type, "*")
		baseField := FieldInfo{
			Type:     baseType,
			IsPtr:    false,
			IsSlice:  strings.HasPrefix(baseType, "[]"),
			IsMap:    strings.HasPrefix(baseType, "map["),
			IsStruct: field.IsStruct || (!isPrimitiveType(baseType) && !strings.HasPrefix(baseType, "[]") && !strings.HasPrefix(baseType, "map[")),
		}

		if isPrimitiveType(baseType) {
			resetCode := generateValueReset("*s."+fieldName, baseType, baseField)
			return fmt.Sprintf("if s.%s != nil {\n\t\t%s\n\t}", fieldName, resetCode)
		}

		if baseField.IsStruct {
			if strings.Contains(baseType, ".") {
				return fmt.Sprintf("if s.%s != nil {\n\t\tif r, ok := interface{}(s.%s).(interface{ Reset() }); ok {\n\t\t\tr.Reset()\n\t\t} else {\n\t\t\t*s.%s = %s{}\n\t\t}\n\t}", fieldName, fieldName, fieldName, baseType)
			}
			return fmt.Sprintf("if s.%s != nil {\n\t\tif r, ok := interface{}(s.%s).(interface{ Reset() }); ok {\n\t\t\tr.Reset()\n\t\t} else {\n\t\t\t*s.%s = %s{}\n\t\t}\n\t}", fieldName, fieldName, fieldName, baseType)
		}

		if baseField.IsSlice {
			return fmt.Sprintf("if s.%s != nil {\n\t\t*s.%s = (*s.%s)[:0]\n\t}", fieldName, fieldName, fieldName)
		}

		if baseField.IsMap {
			return fmt.Sprintf("if s.%s != nil {\n\t\tclear(*s.%s)\n\t}", fieldName, fieldName)
		}

		resetCode := generateValueReset("*s."+fieldName, baseType, baseField)
		return fmt.Sprintf("if s.%s != nil {\n\t\t%s\n\t}", fieldName, resetCode)
	}

	if field.IsSlice {
		return fmt.Sprintf("s.%s = s.%s[:0]", fieldName, fieldName)
	}

	if field.IsMap {
		return fmt.Sprintf("clear(s.%s)", fieldName)
	}

	return generateValueReset("s."+fieldName, field.Type, field)
}

func generateValueReset(target, typeName string, field FieldInfo) string {
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

	if strings.HasPrefix(typeName, "[]") {
		return fmt.Sprintf("%s = %s[:0]", target, target)
	}

	if strings.HasPrefix(typeName, "map[") {
		return fmt.Sprintf("clear(%s)", target)
	}

	if field.IsStruct || (!field.IsPtr && !field.IsSlice && !field.IsMap && !isPrimitiveType(typeName)) {
		if strings.Contains(typeName, ".") {
			return fmt.Sprintf("if r, ok := interface{}(&%s).(interface{ Reset() }); ok { r.Reset() } else { %s = %s{} }", target, target, typeName)
		}
		return fmt.Sprintf("if r, ok := interface{}(&%s).(interface{ Reset() }); ok { r.Reset() } else { %s = %s{} }", target, target, typeName)
	}

	return fmt.Sprintf("%s = %s{}", target, typeName)
}

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

func findProjectRoot(startDir string) string {
	for dir := startDir; ; {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
