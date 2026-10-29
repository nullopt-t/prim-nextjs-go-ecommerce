package arch

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getBackendInternalDir resolves the backend/internal absolute directory.
func getBackendInternalDir(t *testing.T) string {
	// Tests run inside backend/tests/arch
	dir, err := os.Getwd()
	require.NoError(t, err)

	// Traverse up to find backend/internal
	for {
		candidate := filepath.Join(dir, "internal")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		candidate = filepath.Join(dir, "backend", "internal")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("Could not locate backend/internal directory")
	return ""
}

// =============================================================================
// Invariant Check 1: Zero Raw DB Errors in HTTP Handlers (Rule 5)
// =============================================================================

// TestArch_ZeroRawDBErrorsInHandlers verifies that no HTTP handler files
// inspect raw pgx, pgconn, or pgerrcode driver errors. Handlers must only
// handle typed *apierr.APIError or invoke c.Error(err).
func TestArch_ZeroRawDBErrorsInHandlers(t *testing.T) {
	internalDir := getBackendInternalDir(t)
	forbiddenRegex := regexp.MustCompile(`\b(pgx\.|pgconn\.|pgerrcode\b)`)

	var violations []string

	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		filename := d.Name()
		if strings.Contains(filename, "handler") && strings.HasSuffix(filename, ".go") && !strings.HasSuffix(filename, "_test.go") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			lines := strings.Split(string(content), "\n")
			for i, line := range lines {
				// Strip comments
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "//") {
					continue
				}
				if forbiddenRegex.MatchString(line) {
					violations = append(violations, path+":"+string(rune(i+1))+": "+trimmed)
				}
			}
		}
		return nil
	})

	require.NoError(t, err)
	assert.Empty(t, violations, "HTTP handlers must never inspect raw database errors or driver codes:\n"+strings.Join(violations, "\n"))
}

// =============================================================================
// Invariant Check 2: Repository Method Signature Convention (Rule 1 & 2)
// =============================================================================

// TestArch_RepositorySignatures verifies that all exported methods on Repository
// structs follow the standard signature: (ctx context.Context, qe database.QueryExecutor, ...).
func TestArch_RepositorySignatures(t *testing.T) {
	internalDir := getBackendInternalDir(t)
	fset := token.NewFileSet()

	var nonCompliantMethods []string

	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		// Only inspect repository implementation files
		if !strings.Contains(d.Name(), "repo") && !strings.Contains(d.Name(), "repository") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		for _, decl := range node.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			// Only check exported methods
			if !fn.Name.IsExported() {
				continue
			}

			// Receiver check: must be a repository type
			recvType := ""
			switch t := fn.Recv.List[0].Type.(type) {
			case *ast.StarExpr:
				if ident, ok := t.X.(*ast.Ident); ok {
					recvType = ident.Name
				}
			case *ast.Ident:
				recvType = t.Name
			}

			if !strings.Contains(strings.ToLower(recvType), "repo") {
				continue
			}

			// Method parameter check: param 0 = context.Context, param 1 = database.QueryExecutor
			params := fn.Type.Params.List
			if len(params) < 2 {
				nonCompliantMethods = append(nonCompliantMethods, path+": "+fn.Name.Name+" has < 2 params")
				continue
			}

			// Param 1 check: context.Context
			param0Type := typeToString(params[0].Type)
			if !strings.Contains(param0Type, "context.Context") && !strings.Contains(param0Type, "Context") {
				nonCompliantMethods = append(nonCompliantMethods, path+": "+fn.Name.Name+" 1st param is not context.Context ("+param0Type+")")
			}

			// Param 2 check: database.QueryExecutor or *database.Tx
			param1Type := typeToString(params[1].Type)
			if !strings.Contains(param1Type, "QueryExecutor") &&
				!strings.Contains(param1Type, "database.QueryExecutor") &&
				!strings.Contains(param1Type, "database.Tx") &&
				!strings.Contains(param1Type, "Tx") {
				nonCompliantMethods = append(nonCompliantMethods, path+": "+fn.Name.Name+" 2nd param is not database.QueryExecutor or *database.Tx ("+param1Type+")")
			}
		}
		return nil
	})

	require.NoError(t, err)
	assert.Empty(t, nonCompliantMethods, "Repository methods must strictly begin with (ctx context.Context, qe database.QueryExecutor, ...):\n"+strings.Join(nonCompliantMethods, "\n"))
}

// =============================================================================
// Invariant Check 2b: Consumer Readers Are Read-Only
// =============================================================================

// TestArch_ConsumerRepositoriesAreReadOnly enforces that consumer-defined *Reader
// and *Repository interfaces outside their owning package contain ONLY read operations (lookups, queries).
// Cross-domain writes must go through the owning domain's Service or *Writer interface.
func TestArch_ConsumerRepositoriesAreReadOnly(t *testing.T) {
	internalDir := getBackendInternalDir(t)
	fset := token.NewFileSet()

	forbiddenWritePrefixes := []string{
		"Create", "Update", "Delete", "Replace", "Insert", "Record", "Add", "Remove", "Clear", "Adjust", "SoftDelete",
	}

	var writeViolations []string

	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		// Only check service files or consumers where reader/repository interfaces are defined
		if !strings.Contains(d.Name(), "service") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				name := typeSpec.Name.Name
				if !strings.HasSuffix(name, "Reader") && !strings.HasSuffix(name, "Repository") {
					continue
				}

				// ObjectRepository is an infrastructure exception for storage blob pairing per Rule 3
				if name == "ObjectRepository" {
					continue
				}

				iface, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				for _, method := range iface.Methods.List {
					for _, mname := range method.Names {
						methodName := mname.Name
						for _, prefix := range forbiddenWritePrefixes {
							if strings.HasPrefix(methodName, prefix) {
								writeViolations = append(writeViolations, path+": "+name+"."+methodName)
							}
						}
					}
				}
			}
		}
		return nil
	})

	require.NoError(t, err)
	assert.Empty(t, writeViolations, "Consumer reader interfaces must be read-only (cross-domain writes must use domain services or writers):\n"+strings.Join(writeViolations, "\n"))
}

// =============================================================================
// Invariant Check 3: Foreign Domain Service Injections (Rule 1)
// =============================================================================

// TestArch_ZeroForeignServiceInjections verifies that domain services do not inject
// foreign domain services (*ObjectService, *VariantService, *InventoryService, *ProductService).
func TestArch_ZeroForeignServiceInjections(t *testing.T) {
	internalDir := getBackendInternalDir(t)
	fset := token.NewFileSet()

	// Forbidden cross-domain service types
	forbiddenServices := map[string]string{
		"ObjectService":    "object",
		"VariantService":   "variant",
		"InventoryService": "inventory",
		"ProductService":   "product",
	}

	strictAll := os.Getenv("STRICT_ARCH_ALL") == "1"

	var violations []string
	var pendingMilestoneViolations []string

	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		// Only check service files
		if !strings.Contains(d.Name(), "service") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		pkgName := node.Name.Name

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
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

				// Check each field in the Service struct
				for _, field := range structType.Fields.List {
					fieldType := typeToString(field.Type)
					for forbidden, ownerPkg := range forbiddenServices {
						if strings.Contains(fieldType, forbidden) && pkgName != ownerPkg {
							fieldNames := ""
							for _, n := range field.Names {
								fieldNames += n.Name + " "
							}
							violation := path + ": struct " + typeSpec.Name.Name + " field '" + fieldNames + "' has forbidden foreign service type " + fieldType

							// Check if this file is pending milestone implementation
							if strings.Contains(path, "checkout") || strings.Contains(path, "order") || strings.Contains(path, "product_service.go") {
								pendingMilestoneViolations = append(pendingMilestoneViolations, violation)
							} else {
								violations = append(violations, violation)
							}
						}
					}
				}
			}
		}
		return nil
	})

	require.NoError(t, err)

	// Completed/baseline domains must have 0 violations
	assert.Empty(t, violations, "Completed domain services must have zero foreign service injections:\n"+strings.Join(violations, "\n"))

	if strictAll {
		assert.Empty(t, pendingMilestoneViolations, "Strict architecture check: zero foreign service injections across entire repository:\n"+strings.Join(pendingMilestoneViolations, "\n"))
	} else if len(pendingMilestoneViolations) > 0 {
		t.Logf("Tracked foreign service injections pending milestone completion (%d items):\n%s",
			len(pendingMilestoneViolations), strings.Join(pendingMilestoneViolations, "\n"))
	}
}

// =============================================================================
// Invariant Check 4: Elimination of Dead ...WithTx Twin Methods (Rule 4)
// =============================================================================

// TestArch_NoDeadTwinWithTxMethods verifies that dead ...WithTx twin methods are eliminated.
func TestArch_NoDeadTwinWithTxMethods(t *testing.T) {
	internalDir := getBackendInternalDir(t)
	fset := token.NewFileSet()

	// Dead twins targeted for removal in PROJECT.md R4 & F1:
	// ProductService.GetByIDWithTx, VariantService.GetVariantByIDWithTx, ObjectService.CreateObjectWithTx
	deadTwins := map[string]string{
		"GetByIDWithTx":        "ProductService",
		"GetVariantByIDWithTx": "VariantService",
		"CreateObjectWithTx":   "ObjectService",
	}

	strictAll := os.Getenv("STRICT_ARCH_ALL") == "1"
	var foundDeadTwins []string

	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		for _, decl := range node.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}

			methodName := fn.Name.Name
			recvType := typeToString(fn.Recv.List[0].Type)

			for twinMethod, targetService := range deadTwins {
				if methodName == twinMethod && strings.Contains(recvType, targetService) {
					foundDeadTwins = append(foundDeadTwins, path+": "+recvType+"."+methodName)
				}
			}
		}
		return nil
	})

	require.NoError(t, err)

	if strictAll {
		assert.Empty(t, foundDeadTwins, "Strict architecture check: dead ...WithTx twins must be removed:\n"+strings.Join(foundDeadTwins, "\n"))
	} else if len(foundDeadTwins) > 0 {
		t.Logf("Remaining dead ...WithTx twin methods tracked for removal in M1 (%d items):\n%s",
			len(foundDeadTwins), strings.Join(foundDeadTwins, "\n"))
	}
}

// =============================================================================
// Invariant Check 5: Direct Storage Resolution (Rule 3)
// =============================================================================

// TestArch_DirectStorageResolution verifies that domains generating asset URLs
// use storage.StorageProvider.GetPublicURL directly.
func TestArch_DirectStorageResolution(t *testing.T) {
	internalDir := getBackendInternalDir(t)

	// Domains that must use storageProvider: catalog/product, object, checkout
	targetFiles := []string{
		filepath.Join(internalDir, "catalog", "product", "product_service.go"),
		filepath.Join(internalDir, "object", "object_service.go"),
	}

	for _, file := range targetFiles {
		if _, err := os.Stat(file); err == nil {
			content, err := os.ReadFile(file)
			require.NoError(t, err)
			fileStr := string(content)

			assert.True(t,
				strings.Contains(fileStr, "storage.StorageProvider") || strings.Contains(fileStr, "StorageProvider"),
				file+" must inject storage.StorageProvider",
			)
			assert.True(t,
				strings.Contains(fileStr, "GetPublicURL"),
				file+" must use GetPublicURL for public asset resolution",
			)
		}
	}
}

// helper: typeToString formats an ast.Expr type representation.
func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	case *ast.Ellipsis:
		return "..." + typeToString(t.Elt)
	default:
		return ""
	}
}
