package hook

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io/fs"
	"log"
	"log/slog"
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
	goDirs, err := getGoPackages("internal/")
	if err != nil {
		panic(err)
	}

	_log := slog.Default()
	cfg := &packages.Config{
		Mode: packages.NeedDeps | packages.NeedSyntax,
	}

	// Load only Go directories
	_packages, err := packages.Load(cfg, goDirs...)
	if err != nil {
		panic(err)
	}
	// Filter out unwanted packages
	filteredPkgs := excludeInternalAndVendor(_packages)
	fmt.Println(_packages)

	registry := NewTypeRegistry() // Create a new type registry

	// Iterate over all loaded packages
	for _, pkg := range filteredPkgs {
		fmt.Println(pkg)
		for _, file := range pkg.Syntax {
			// Inspect the AST
			for _, decl := range file.Decls {
				fmt.Println(":::::---> ", decl)
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
	log.Println("---> ")
	for _, field := range structType.Fields.List {
		// Check if the field is a type
		if field.Type != nil {
			// Check for embedded structs by examining the type
			if typeIdent, ok := field.Type.(*ast.Ident); ok {
				// Check if the type matches "Inject"
				if typeIdent.Name == "Inject" {
					return true
				}
			} else if typeSpec, ok := field.Type.(*ast.StarExpr); ok {
				// Check if the embedded type is a pointer to a type
				if ident, ok := typeSpec.X.(*ast.Ident); ok && ident.Name == "Inject" {
					return true
				}
			}
		}
	}
	return false
}

func getGoPackages(root string) ([]string, error) {
	var goDirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the directory contains Go files
		if d.IsDir() {
			hasGoFile := false
			err := filepath.WalkDir(path, func(subPath string, subD fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if filepath.Ext(subPath) == ".go" {
					hasGoFile = true
					return filepath.SkipDir // We found a Go file, no need to continue
				}
				return nil
			})

			if err == nil && hasGoFile {
				goDirs = append(goDirs, path)
			}
		}

		return nil
	})

	return goDirs, err
}

func excludeInternalAndVendor(pkgs []*packages.Package) []*packages.Package {
	var filtered []*packages.Package
	for _, pkg := range pkgs {
		if !strings.Contains(pkg.PkgPath, "cmd/internal/") &&
			!strings.Contains(pkg.PkgPath, "cmd/vendor/") &&
			!strings.Contains(pkg.PkgPath, "golang.org/x/arch/") {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}
