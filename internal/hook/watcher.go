package hook

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"reflect"
)

// WatchAndInjectHooks finds structs with hookie.Inject and calls their hooks
func WatchAndInjectHooks(rootDir string, ctx context.Context) error {
	log.Println("watching hooks")
	// Load packages from the specified directory
	cfg := &packages.Config{
		Dir:  rootDir,
		Mode: packages.LoadSyntax,
	}

	_packages, err := packages.Load(cfg, ".")
	if err != nil {
		return err
	}
	log.Println("packages:", _packages)

	// Iterate over all loaded packages
	for _, pkg := range _packages {
		for _, file := range pkg.Syntax {
			// Inspect the AST
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						typeSpec := spec.(*ast.TypeSpec)
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							// Check for hookie.Inject tag
							if isInjectable(structType) {
								structName := typeSpec.Name.Name
								fmt.Printf("Found injectable struct: %s\n", structName)
								obj := pkg.Types.Scope().Lookup(structName)
								if obj == nil {
									return fmt.Errorf("could not find type for struct: %s", structName)
								}
								if t, ok := obj.Type().(*types.Struct); ok {
									reflectType := reflect.TypeOf(t)
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

// isInjectable checks if the struct has the hookie.Inject tag
func isInjectable(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if name.String() == "Inject" { // Change as needed for your tag checking logic
				return true
			}
		}
	}
	return false
}
