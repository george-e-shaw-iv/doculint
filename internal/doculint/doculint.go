// Package doculint contains the necessary logic for the doculint linter.
package doculint

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer exports the doculint analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: "doculint",
	Doc:  "checks for proper function, type, package, constant, and string and numeric literal documentation",
	Run:  doculint,
}

// doculint is the function that gets passed to the Analyzer which runs the actual
// analysis for the doculint linter on a set of files.
func doculint(pass *analysis.Pass) (interface{}, error) {
	// packageWithSameNameFile keep track of which packages have a file with the same
	// name as the package and which do not (the convention is that this file will
	// contain the package documentation).
	packageWithSameNameFile := make(map[string]bool)

	if msg := validatePackageName(pass.Pkg.Name()); msg != "" {
		pass.Reportf(0, msg)
	}

	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			// Ignore the main package, it doesn't need a package comment.

			// Add this package to packageWithSameNameFile if it does not already
			// exist.
			if _, exists := packageWithSameNameFile[pass.Pkg.Name()]; !exists {
				packageWithSameNameFile[pass.Pkg.Name()] = false
			}

			if file.Name.Name == pass.Pkg.Name() {
				packageWithSameNameFile[pass.Pkg.Name()] = true

				if file.Doc == nil {
					pass.Reportf(0, "package \"%s\" has no comment associated with it in \"%s.go\"", pass.Pkg.Name(), pass.Pkg.Name())
				} else {
					expectedPrefix := fmt.Sprintf("Package %s", pass.Pkg.Name())
					if !strings.HasPrefix(strings.TrimSpace(file.Doc.Text()), expectedPrefix) {
						pass.Reportf(0, "comment for package \"%s\" should begin with \"%s\"", pass.Pkg.Name(), expectedPrefix)
					}
				}
			}
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch expr := n.(type) {
			case *ast.FuncDecl:
				if pass.Pkg.Name() == "main" && expr.Name.Name == "main" {
					// Ignore func main in main package.
					return true
				}

				if expr.Name.Name == "init" {
					// Ignore init functions.
					return true
				}

				if expr.Doc == nil {
					pass.Reportf(expr.Pos(), "function \"%s\" has no comment associated with it", expr.Name.Name)
					return true
				}

				if !strings.HasPrefix(strings.TrimSpace(expr.Doc.Text()), expr.Name.Name) {
					pass.Reportf(expr.Pos(), "comment for function \"%s\" should begin with \"%s\"", expr.Name.Name, expr.Name.Name)
					return true
				}
			case *ast.IfStmt:
				be, ok := expr.Cond.(*ast.BinaryExpr)
				if !ok {
					return true
				}

				if literal, ok := be.X.(*ast.BasicLit); ok {
					pass.Reportf(literal.Pos(), "literal found in conditional")
				}

				if literal, ok := be.Y.(*ast.BasicLit); ok {
					pass.Reportf(literal.Pos(), "literal found in conditional")
				}
			case *ast.GenDecl:
				if expr.Tok == token.CONST {
					if expr.Lparen.IsValid() {
						// Constant block
						if expr.Doc == nil {
							pass.Reportf(expr.Pos(), "constant block has no comment associated with it")
						}
					}

					for i := range expr.Specs {
						vs, ok := expr.Specs[i].(*ast.ValueSpec)
						if ok {
							if len(vs.Names) > 1 {
								var names []string
								for j := range vs.Names {
									names = append(names, vs.Names[j].Name)
								}

								pass.Reportf(vs.Pos(), "constants \"%s\" should be separated and each have a comment associated with them", strings.Join(names, ", "))
								continue
							}

							name := vs.Names[0].Name

							doc := vs.Doc
							if !expr.Lparen.IsValid() {
								// If this constant isn't apart of a constant block it's comment is stored in the *ast.GenDecl type.
								doc = expr.Doc
							}

							if doc == nil {
								pass.Reportf(vs.Pos(), "constant \"%s\" has no comment associated with it", name)
								continue
							}

							if !strings.HasPrefix(strings.TrimSpace(doc.Text()), name) {
								pass.Reportf(vs.Pos(), "comment for constant \"%s\" should begin with \"%s\"", name, name)
							}
						}
					}
				} else if expr.Tok == token.TYPE {
					if expr.Lparen.IsValid() {
						// Type block
						if expr.Doc == nil {
							pass.Reportf(expr.Pos(), "type block has no comment associated with it")
						}
					}

					for i := range expr.Specs {
						ts, ok := expr.Specs[i].(*ast.TypeSpec)
						if ok {
							doc := ts.Doc
							if !expr.Lparen.IsValid() {
								// If this type isn't apart of a type block it's comment is stored in the *ast.GenDecl type.
								doc = expr.Doc
							}

							if doc == nil {
								pass.Reportf(ts.Pos(), "type \"%s\" has no comment associated with it", ts.Name.Name)
								continue
							}

							if !strings.HasPrefix(strings.TrimSpace(doc.Text()), ts.Name.Name) {
								pass.Reportf(ts.Pos(), "comment for type \"%s\" should begin with \"%s\"", ts.Name.Name, ts.Name.Name)
							}
						}
					}
				}
			default:
				return true
			}

			return true
		})
	}

	for pkg := range packageWithSameNameFile {
		if !packageWithSameNameFile[pkg] {
			pass.Reportf(0, "package \"%s\" has no file with the same name containing package comment", pkg)
		}
	}

	return nil, nil
}

// validatePackageName ensures that a given package name follows the conventions that can
// be read about here: https://blog.golang.org/package-names
func validatePackageName(pkg string) string {
	if strings.ContainsAny(pkg, "_-") {
		return fmt.Sprintf("package \"%s\" should not contain - or _ in name", pkg)
	}

	if pkg != strings.ToLower(pkg) {
		return fmt.Sprintf("package \"%s\" should be all lowercase", pkg)
	}

	return ""
}
