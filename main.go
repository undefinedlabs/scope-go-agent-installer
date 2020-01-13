package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path"
	"text/template"
)

type testPackageInfo struct {
	Name      string
	Folder    string
	MainFile  *string
	MainParam bool
}

var testMainFileTemplate, _ = template.New("testMain").Parse(
	`
package {{.Name}}

import (
	"fmt"
	"os"
	"go.undefinedlabs.com/scopeagent"
	"go.undefinedlabs.com/scopeagent/agent"
	"go.undefinedlabs.com/scopeagent/instrumentation/nethttp"
)

func TestMain(m *testing.M) {
	fmt.Println("Starting Tests")
	nethttp.PatchHttpDefaultClient()
	os.Exit(scopeagent.Run(m, agent.WithSetGlobalTracer()))
}
`)

func main() {
	folder := flag.String("folder", "", "Source code folder to analyze.")
	required := []string{"folder"}
	flag.Parse()

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			_, _ = fmt.Fprintf(os.Stderr, "missing required -%s argument.\n", req)
			os.Exit(2)
		}
	}

	if folder != nil {
		testPackageInfoMap := map[string]*testPackageInfo{}
		fmt.Printf("Processing: %s\n", *folder)
		err := processFolderTestFiles(*folder, processTestFile, &testPackageInfoMap)
		if err != nil {
			fmt.Println(err)
		}
		for _, tpi := range testPackageInfoMap {
			if tpi.MainFile == nil {
				mFile := path.Join(tpi.Folder, "scope_pkg_test.go")
				tpi.MainFile = &mFile

				fmt.Printf("Creating TestMain func for package %v in %v.\n", tpi.Name, *tpi.MainFile)

				file, errFile := os.Create(mFile)
				if errFile != nil {
					log.Fatalf("execution failed: %s", errFile)
				}
				writer := bufio.NewWriter(file)
				err := testMainFileTemplate.Execute(writer, tpi)
				if err != nil {
					log.Fatalf("execution failed: %s", err)
				}
				err = writer.Flush()
				err = file.Close()

			} else {
				fmt.Printf("Updating TestMain func for package %v in %v.\n", tpi.Name, *tpi.MainFile)
				err = processFileTestMain(*tpi.MainFile)
			}
			if err != nil {
				fmt.Println(err)
			}
		}
		fmt.Println("Done.")
	}
}

func processFolderTestFiles(folder string, fileProcessor func(string, interface{}) (bool, error), state interface{}) error {
	f, err := os.Open(folder)
	if err != nil {
		return err
	}
	entries, err := f.Readdir(-1)
	if err != nil {
		return err
	}
	mainFound := false
	for _, entry := range entries {
		entryName := entry.Name()
		if entryName[0] == '.' {
			continue
		}
		filePath := fmt.Sprintf("%s/%s", folder, entryName)
		if entry.IsDir() {
			fErr := processFolderTestFiles(filePath, fileProcessor, state)
			if fErr != nil {
				return fErr
			}
		} else if len(filePath) > 8 && !mainFound {
			if filePath[len(filePath)-8:] != "_test.go" {
				continue
			}
			found, fErr := fileProcessor(filePath, state)
			if fErr != nil {
				return fErr
			}
			mainFound = found
		}
	}
	return nil
}

func processTestFile(filePath string, state interface{}) (bool, error) {
	fSet := token.NewFileSet()
	fileParser, err := parser.ParseFile(fSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return false, err
	}

	tmInfoMap := *(state.(*map[string]*testPackageInfo))
	tpi := tmInfoMap[fileParser.Name.Name]
	if tpi == nil {
		tpi = &testPackageInfo{
			Name:      fileParser.Name.Name,
			Folder:    path.Dir(filePath),
			MainFile:  nil,
			MainParam: false,
		}
		tmInfoMap[fileParser.Name.Name] = tpi
	}

	found := false
	for _, decl := range fileParser.Decls {
		if _, hasTMParam, hasTMName := isTestMainFunc(decl); hasTMName {
			tpi.MainFile = &filePath
			tpi.MainParam = hasTMParam
			found = true
		}
	}

	return found, nil
}

func processFileTestMain(filePath string) error {
	fSet := token.NewFileSet()
	fileParser, err := parser.ParseFile(fSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	hasImports := false
	hasTMParams := false
	var testMainFunc *ast.FuncDecl
	currentImportName := ImportName
	for _, decl := range fileParser.Decls {
		if genDcl, ok := isImportDeclaration(decl); ok {
			hasImports = true
			osImportFound := false
			testingImportFound := false
			agentImportFound := false
			agentOptionsImportFound := false
			for _, impSpec := range genDcl.Specs {
				importSentence := impSpec.(*ast.ImportSpec)
				if importSentence.Path.Value == ImportPath {
					agentImportFound = true
					continue
				}
				if importSentence.Path.Value == AgentImportPath {
					agentOptionsImportFound = true
					continue
				}
				if importSentence.Path.Value == "\"os\"" {
					osImportFound = true
					continue
				}
				if importSentence.Path.Value == "\"testing\"" {
					testingImportFound = true
					continue
				}
			}
			if !osImportFound {
				genDcl.Specs = append(genDcl.Specs, getOsImportSpec())
			}
			if !testingImportFound {
				genDcl.Specs = append(genDcl.Specs, getTestingImportSpec())
			}
			if !agentImportFound {
				genDcl.Specs = append(genDcl.Specs, getAgentImportSpec())
			}
			if !agentOptionsImportFound {
				genDcl.Specs = append(genDcl.Specs, getAgentOptionsImportSpec())
			}
		} else if fDecl, hTmParams, hasTmName := isTestMainFunc(decl); hasTmName {
			testMainFunc = fDecl
			hasTMParams = hTmParams
		}
	}

	if !hasImports {
		importDeclaration := getImportDeclaration()
		importDeclaration.Specs = append(importDeclaration.Specs, getAgentImportSpec(), getAgentOptionsImportSpec(),
			getOsImportSpec(), getTestingImportSpec())
		fileParser.Decls = append([]ast.Decl{importDeclaration}, fileParser.Decls...)
	}

	if testMainFunc != nil {
		if !testMainHasGlobalAgent(testMainFunc, currentImportName) {
			if hasTMParams {
				if modifyExistingTestMain(testMainFunc, currentImportName) {
					fmt.Printf("  The TestMain func of package '%s' in '%s' has been patched.\n",
						fileParser.Name.Name, filePath)
				} else {
					fmt.Printf("  Package '%s' already has a TestMain func in '%s', please modify the file manually.\n",
						fileParser.Name.Name, filePath)
				}
			} else {
				fmt.Printf("  Package '%s' already has a TestMain func in '%s' but doesn't have the right 'testing.M' parameter.\n",
					fileParser.Name.Name, filePath)
				return nil
			}
		} else {
			fmt.Printf("  Package '%s' already patched in '%s'.\n",
				fileParser.Name.Name, filePath)
		}
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = printer.Fprint(f, fSet, fileParser)
	if err != nil {
		return err
	}

	return nil
}
