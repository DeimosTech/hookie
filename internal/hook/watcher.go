package hook

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type TypeRegistry struct {
	types map[string]reflect.Type // Key: "packagePath.typeName"
}

func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{types: make(map[string]reflect.Type)}
}

func (tr *TypeRegistry) Register(pkgPath, typeName string, typ reflect.Type) {
	key := fmt.Sprintf("%s.%s", pkgPath, typeName)
	tr.types[key] = typ
}

func (tr *TypeRegistry) Lookup(pkgPath, typeName string) (reflect.Type, bool) {
	key := fmt.Sprintf("%s.%s", pkgPath, typeName)
	typ, exists := tr.types[key]
	return typ, exists
}

// WatchAndInjectHooks finds structs with hookie.Inject and calls their hooks
func WatchAndInjectHooks(rootDir string, ctx context.Context) error {
	// Load all packages from the specified directory and its subdirectories
	goDirs, err := collectGoDirs("./")
	if err != nil {
		log.Fatalf("Error collecting Go directories: %v", err)
	}
	moduleName, err := getGoModuleName("./")
	if err != nil {
		log.Fatalf("Error getting Go module name: %v", err)
	}
	// Convert the list of directories into a pattern for packages.Load
	//dirPatterns := make([]string, len(goDirs))
	for _, dir := range goDirs {
		path := filepath.Join(moduleName, dir)
		err = WatchAndRegister(path, ctx)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

// WatchAndRegister finds structs with hookie.Inject and calls their hooks
func WatchAndRegister(dir string, ctx context.Context) error {
	_log := slog.Default()
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedImports, // Avoid NeedDeps if unnecessary
	}

	// Handle the panic gracefully
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	_packages, err := packages.Load(cfg, "gitlab.com/shikho/backend/mask/"+dir)
	if err != nil {
		log.Fatal(err)
	}

	registry := NewTypeRegistry()

	for _, pkg := range _packages {
		for _, file := range pkg.Syntax {
			// Inspect the AST
			for _, decl := range file.Decls {
				//fmt.Println(":::::---> ", decl)
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						typeSpec := spec.(*ast.TypeSpec)
						fmt.Println("----> ", typeSpec.Name.Name)
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							fmt.Println("-xx--> ", structType)
							// Check for hookie.Inject tag
							if isInjectable(structType) {
								structName := typeSpec.Name.Name
								_log.Info("Found injectable struct: %s in package: %s\n", structName, pkg.PkgPath)

								// Get the full type from the types package
								obj := pkg.Types.Scope().Lookup(structName)
								if obj == nil {
									return fmt.Errorf("could not find type for struct: %s", structName)
								}

								// Ensure the object is of type *types.Named
								if t, ok := obj.Type().(*types.Named); ok && t != nil {
									// Register the type in the registry
									registry.Register(pkg.PkgPath, structName, reflect.TypeOf(struct{}{})) // Create a dummy instance to get reflect.Type

									// Get the underlying type
									underlyingType := t.Underlying()
									_, ok := underlyingType.(*types.Struct)
									if !ok {
										return fmt.Errorf("%s is not a struct type", structName)
									}

									// Now look up the type from the registry
									reflectType, exists := registry.Lookup(pkg.PkgPath, structName)
									if !exists {
										return fmt.Errorf("reflect type for %s not found in registry", structName)
									}

									// Use reflection to create a new instance of the struct
									instance := reflect.New(reflectType).Interface()
									RegisterModel(instance)
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// isInjectable checks if the struct has the hookie.Inject embedded type
func isInjectable(structType *ast.StructType) bool {

	for _, field := range structType.Fields.List {
		// Check if the field is a type
		if selExpr, ok := field.Type.(*ast.SelectorExpr); ok {
			pkgIdent, ok := selExpr.X.(*ast.Ident)
			if ok && pkgIdent.Name == "instance" && selExpr.Sel.Name == "Inject" {
				// Found a field of type instance.Inject
				return true
			}
		}

		// Handle non-qualified names (for example, just Inject)
		if typeIdent, ok := field.Type.(*ast.Ident); ok {
			if typeIdent.Name == "Inject" {
				return true
			}
		}

		// Handle pointer to a type (e.g., *Inject or *instance.Inject)
		if typeSpec, ok := field.Type.(*ast.StarExpr); ok {
			// If it's a pointer to a simple type (e.g., *Inject)
			if ident, ok := typeSpec.X.(*ast.Ident); ok && ident.Name == "Inject" {
				return true
			}
			// If it's a pointer to a qualified type (e.g., *instance.Inject)
			if selExpr, ok := typeSpec.X.(*ast.SelectorExpr); ok {
				pkgIdent, ok := selExpr.X.(*ast.Ident)
				if ok && pkgIdent.Name == "instance" && selExpr.Sel.Name == "Inject" {
					return true
				}
			}
		}
	}
	return false
}

func collectGoDirs(baseDir string) ([]string, error) {
	var goDirs []string

	// Walk through the base directory recursively
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current path is a directory
		if info.IsDir() {
			// Check if the directory contains Go files
			if hasGoFiles(path) && !isExclude(path) {
				goDirs = append(goDirs, path)
			}
		}
		return nil
	})
	return goDirs, err
}

func hasGoFiles(dir string) bool {
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", dir, err)
		return false
	}

	// Check each file in the directory
	for _, file := range files {
		// If any file has a ".go" extension, the directory has Go files
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
			return true
		}
	}
	return false
}
func isExclude(dir string) bool {
	if strings.HasPrefix(dir, "cmd") ||
		strings.HasPrefix(dir, "golang.org") {
		return true
	}
	return false
}

func getGoModuleName(dir string) (string, error) {
	goModPath := filepath.Join(dir, "go.mod")

	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("could not read go.mod file: %v", err)
	}
	// Parse the go.mod file
	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", fmt.Errorf("could not parse go.mod file: %v", err)
	}
	// Return the module name
	return modFile.Module.Mod.Path, nil
}
