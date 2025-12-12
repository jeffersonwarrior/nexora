package indexer

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Symbol represents a code element extracted from AST
type Symbol struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"` // func, struct, interface, var, const
	Package   string   `json:"package"`
	File      string   `json:"file"`
	Line      int      `json:"line"`
	Column    int      `json:"column"`
	Signature string   `json:"signature"`
	Doc       string   `json:"doc"`
	Imports   []string `json:"imports"`
	Callers   []string `json:"callers"`
	Calls     []string `json:"calls"`
	Public    bool     `json:"public"`
	Params    []Param  `json:"params,omitempty"`
	Returns   []string `json:"returns,omitempty"`
	Fields    []Field  `json:"fields,omitempty"`
	Methods   []string `json:"methods,omitempty"`
}

type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ASTParser extracts symbols from Go source code
type ASTParser struct {
	fset    *token.FileSet
	symbols []Symbol
}

func NewASTParser() *ASTParser {
	return &ASTParser{
		fset:    token.NewFileSet(),
		symbols: make([]Symbol, 0),
	}
}

// ParseDirectory parses all Go files in a directory using golang.org/x/tools/go/packages
// for better build tag support and module awareness
func (p *ASTParser) ParseDirectory(ctx context.Context, dir string) ([]Symbol, error) {
	p.symbols = make([]Symbol, 0)

	// Check if there's a go.mod file to determine if we should use packages.Load
	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		// No go.mod file, use legacy parser.ParseDir approach
		return p.parseDirectoryLegacy(ctx, dir)
	}

	// Use packages.Load for better build tag support and module awareness
	config := &packages.Config{
		Mode:  packages.NeedSyntax | packages.NeedFiles | packages.NeedCompiledGoFiles,
		Dir:   dir,
		Tests: false,
	}

	// Load packages with pattern "." to get current directory package
	pkgs, err := packages.Load(config, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages from directory %s: %w", dir, err)
	}

	// Process all loaded packages
	for _, pkg := range pkgs {
		if pkg.Errors != nil {
			// If packages.Load has errors, fall back to the legacy parser.ParseDir approach
			return p.parseDirectoryLegacy(ctx, dir)
		}

		// Extract package name from the package
		pkgName := pkg.Name

		// Process each syntax tree (Go file) in the package
		for _, file := range pkg.Syntax {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				// Find the corresponding filename
				filename := ""
				if len(pkg.CompiledGoFiles) > 0 {
					// Use compiled Go files if available (better with build tags)
					for i, syntax := range pkg.Syntax {
						if syntax == file {
							if i < len(pkg.CompiledGoFiles) {
								filename = pkg.CompiledGoFiles[i]
							}
							break
						}
					}
				}

				// Fallback to Go files if compiled file not found
				if filename == "" {
					for i, syntax := range pkg.Syntax {
						if syntax == file {
							if i < len(pkg.GoFiles) {
								filename = pkg.GoFiles[i]
								// Make absolute path relative to the directory
								if !filepath.IsAbs(filename) {
									filename = filepath.Join(dir, filepath.Base(filename))
								}
							}
							break
						}
					}
				}

				// Final fallback to generate filename
				if filename == "" {
					filename = filepath.Join(dir, "generated.go")
				}

				p.parseFile(ctx, pkgName, filename, file)
			}
		}
	}

	// If no packages were found, try legacy approach
	if len(p.symbols) == 0 {
		return p.parseDirectoryLegacy(ctx, dir)
	}

	return p.symbols, nil
}

// parseDirectoryLegacy falls back to parser.ParseDir for compatibility
func (p *ASTParser) parseDirectoryLegacy(ctx context.Context, dir string) ([]Symbol, error) {
	p.symbols = make([]Symbol, 0)

	pkgs, err := parser.ParseDir(p.fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory %s: %w", dir, err)
	}

	for pkgName, pkg := range pkgs {
		for filename, file := range pkg.Files {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				p.parseFile(ctx, pkgName, filename, file)
			}
		}
	}

	return p.symbols, nil
}

// ParseFile extracts symbols from a single Go file
func (p *ASTParser) ParseFile(ctx context.Context, filename string) ([]Symbol, error) {
	p.symbols = make([]Symbol, 0)

	// Parse the single file
	file, err := parser.ParseFile(p.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filename, err)
	}

	// Extract package name
	pkgName := file.Name.Name

	// Parse the file
	p.parseFile(ctx, pkgName, filename, file)

	return p.symbols, nil
}

// parseFile extracts symbols from a single Go file (internal method)
func (p *ASTParser) parseFile(ctx context.Context, pkgName, filename string, file *ast.File) {
	doc := extractFileDoc(file)
	imports := extractImports(file)

	// Visit all nodes in the AST
	ast.Inspect(file, func(n ast.Node) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		switch node := n.(type) {
		case *ast.FuncDecl:
			p.parseFunction(pkgName, filename, node, doc, imports)
		case *ast.GenDecl:
			p.parseGenDecl(pkgName, filename, node, doc, imports)
		case *ast.TypeSpec:
			p.parseTypeSpec(pkgName, filename, node, file, doc, imports)
		}

		return true
	})
}

// parseFunction extracts function information
func (p *ASTParser) parseFunction(pkgName, filename string, decl *ast.FuncDecl, fileDoc string, imports []string) {
	pos := p.fset.Position(decl.Pos())
	signature := formatFunctionSignature(decl)
	doc := extractNodeDoc(decl.Doc)

	sym := Symbol{
		Name:      decl.Name.Name,
		Type:      "function",
		Package:   pkgName,
		File:      filename,
		Line:      pos.Line,
		Column:    pos.Column,
		Signature: signature,
		Doc:       doc,
		Imports:   imports,
		Public:    decl.Name.IsExported(),
		Params:    extractParams(decl.Type.Params),
		Returns:   extractReturns(decl.Type.Results),
	}

	// Handle methods
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		recvType := extractRecvType(decl.Recv.List[0].Type)
		sym.Name = recvType + "." + sym.Name
		sym.Type = "method"
	}

	// Extract function calls within this function
	ast.Inspect(decl, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok {
				sym.Calls = append(sym.Calls, ident.Name)
			}
		}
		return true
	})

	p.symbols = append(p.symbols, sym)
}

// parseGenDecl handles general declarations (var, const)
func (p *ASTParser) parseGenDecl(pkgName, filename string, decl *ast.GenDecl, fileDoc string, imports []string) {
	doc := extractNodeDoc(decl.Doc)
	declType := "var"
	if decl.Tok == token.CONST {
		declType = "const"
	}

	for _, spec := range decl.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			for _, name := range valueSpec.Names {
				pos := p.fset.Position(name.Pos())
				varType := ""
				if valueSpec.Type != nil {
					varType = formatType(valueSpec.Type)
				}

				sym := Symbol{
					Name:    name.Name,
					Type:    declType,
					Package: pkgName,
					File:    filename,
					Line:    pos.Line,
					Column:  pos.Column,
					Doc:     doc,
					Imports: imports,
					Public:  name.IsExported(),
				}

				if varType != "" {
					sym.Signature = fmt.Sprintf("%s %s", name.Name, varType)
				}

				p.symbols = append(p.symbols, sym)
			}
		}
	}
}

// parseTypeSpec handles type declarations
func (p *ASTParser) parseTypeSpec(pkgName, filename string, spec *ast.TypeSpec, file *ast.File, fileDoc string, imports []string) {
	pos := p.fset.Position(spec.Pos())
	doc := extractNodeDoc(spec.Doc)
	typeDef := formatType(spec.Type)

	sym := Symbol{
		Name:      spec.Name.Name,
		Type:      "type",
		Package:   pkgName,
		File:      filename,
		Line:      pos.Line,
		Column:    pos.Column,
		Signature: fmt.Sprintf("type %s %s", spec.Name.Name, typeDef),
		Doc:       doc,
		Imports:   imports,
		Public:    spec.Name.IsExported(),
	}

	// Extract specific type information
	switch t := spec.Type.(type) {
	case *ast.StructType:
		sym.Type = "struct"
		sym.Fields = extractStructFields(t)
	case *ast.InterfaceType:
		sym.Type = "interface"
		sym.Methods = extractInterfaceMethods(t)
	}

	p.symbols = append(p.symbols, sym)
}

// Helper functions for extracting information

func extractFileDoc(file *ast.File) string {
	if len(file.Comments) > 0 {
		return file.Comments[0].Text()
	}
	return ""
}

func extractNodeDoc(doc *ast.CommentGroup) string {
	if doc != nil {
		return doc.Text()
	}
	return ""
}

func extractImports(file *ast.File) []string {
	imports := make([]string, 0)
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, importPath)
	}
	return imports
}

func extractParams(params *ast.FieldList) []Param {
	if params == nil {
		return nil
	}
	result := make([]Param, 0)
	for _, field := range params.List {
		paramType := formatType(field.Type)
		for _, name := range field.Names {
			result = append(result, Param{
				Name: name.Name,
				Type: paramType,
			})
		}
		// Handle unnamed parameters
		if len(field.Names) == 0 {
			result = append(result, Param{
				Name: "",
				Type: paramType,
			})
		}
	}
	return result
}

func extractReturns(results *ast.FieldList) []string {
	if results == nil {
		return nil
	}
	returnValues := make([]string, 0)
	for _, field := range results.List {
		typeStr := formatType(field.Type)
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				returnValues = append(returnValues, fmt.Sprintf("%s %s", name.Name, typeStr))
			}
		} else {
			returnValues = append(returnValues, typeStr)
		}
	}
	return returnValues
}

func extractStructFields(structType *ast.StructType) []Field {
	fields := make([]Field, 0)
	for _, field := range structType.Fields.List {
		fieldType := formatType(field.Type)
		for _, name := range field.Names {
			fields = append(fields, Field{
				Name: name.Name,
				Type: fieldType,
			})
		}
	}
	return fields
}

func extractInterfaceMethods(iface *ast.InterfaceType) []string {
	methods := make([]string, 0)
	for _, method := range iface.Methods.List {
		switch m := method.Type.(type) {
		case *ast.FuncType:
			funcName := ""
			if len(method.Names) > 0 {
				funcName = method.Names[0].Name
			}
			methods = append(methods, funcName)
		case *ast.Ident:
			methods = append(methods, m.Name)
		}
	}
	return methods
}

func extractRecvType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return extractRecvType(t.X)
	case *ast.Ident:
		return t.Name
	default:
		return formatType(expr)
	}
}

func formatFunctionSignature(decl *ast.FuncDecl) string {
	var params []string
	if decl.Type.Params != nil {
		for _, field := range decl.Type.Params.List {
			paramType := formatType(field.Type)
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					params = append(params, fmt.Sprintf("%s %s", name.Name, paramType))
				}
			} else {
				params = append(params, paramType)
			}
		}
	}

	var returns []string
	if decl.Type.Results != nil {
		returns = extractReturns(decl.Type.Results)
	}

	signature := fmt.Sprintf("func %s(%s)", decl.Name.Name, strings.Join(params, ", "))
	if len(returns) > 0 {
		signature += fmt.Sprintf(" (%s)", strings.Join(returns, ", "))
	}

	return signature
}

func formatType(expr ast.Expr) string {
	// Simplified type formatting - can be expanded for more complex types
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		if t.Len != nil {
			return fmt.Sprintf("[%s]%s", formatExpr(t.Len), formatType(t.Elt))
		}
		return "[]" + formatType(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatType(t.Key), formatType(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return "chan<- " + formatType(t.Value)
		case ast.RECV:
			return "<-chan " + formatType(t.Value)
		default:
			return "chan " + formatType(t.Value)
		}
	case *ast.FuncType:
		return "func()" // Simplified for now
	default:
		return fmt.Sprintf("%T", expr) // Fallback
	}
}

func formatExpr(expr ast.Expr) string {
	// Simplified expression formatting
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.BasicLit:
		return t.Value
	default:
		return fmt.Sprintf("%T", expr)
	}
}
