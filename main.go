package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

type testMainInfo struct {
	firstFileOfPackage *string
	testMainFile       *string
	testMainParam      bool
}

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
		packagesTestMainInfo := map[string]*testMainInfo{}
		fmt.Printf("Processing: %s\n", *folder)
		err := processFolderTestFiles(*folder, processFileTestFuncs, &packagesTestMainInfo)
		if err != nil {
			fmt.Println(err)
		}
		for _, tmInfo := range packagesTestMainInfo {
			if tmInfo.testMainFile == nil {
				err = processFileTestMain(*tmInfo.firstFileOfPackage)
			} else {
				err = processFileTestMain(*tmInfo.testMainFile)
			}
			if err != nil {
				fmt.Println(err)
			}
		}
		fmt.Println("Done.")
	}
}

func processFolderTestFiles(folder string, fileProcessor func(string, interface{}) error, state interface{}) error {
	f, err := os.Open(folder)
	if err != nil {
		return err
	}
	entries, err := f.Readdir(-1)
	if err != nil {
		return err
	}
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
		} else if len(filePath) > 8 {
			if filePath[len(filePath)-8:] != "_test.go" {
				continue
			}
			fErr := fileProcessor(filePath, state)
			if fErr != nil {
				return fErr
			}
		}
	}
	return nil
}

func processFileTestFuncs(filePath string, state interface{}) error {
	fSet := token.NewFileSet()
	fileParser, err := parser.ParseFile(fSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	tmInfoMap := *(state.(*map[string]*testMainInfo))
	tmInfo := tmInfoMap[fileParser.Name.Name]
	if tmInfo == nil {
		tmInfo = &testMainInfo{
			firstFileOfPackage: &filePath,
			testMainFile:       nil,
		}
		tmInfoMap[fileParser.Name.Name] = tmInfo
	}

	hasImports := false
	dirty := false
	currentImportName := ImportName
	for _, decl := range fileParser.Decls {
		if genDcl, ok := isImportDeclaration(decl); ok {
			hasImports = true
			importFound := false
			for _, impSpec := range genDcl.Specs {
				importSentence := impSpec.(*ast.ImportSpec)
				if importSentence.Path.Value == ImportPath {
					importFound = true
					currentImportName = importSentence.Name.Name
					break
				}
			}
			if !importFound {
				genDcl.Specs = append(genDcl.Specs, getAgentImportSpec())
			}
		} else if fDecl, varName, ok := isTestFunc(decl); ok {
			if !isStartTestAlreadyImplemented(fDecl, currentImportName) {
				fDecl.Body.List = append([]ast.Stmt{
					getScopeAgentStartTestSentence(currentImportName, varName),
					getScopeAgentEndTestDeferSentence(),
				}, fDecl.Body.List...)
				dirty = true
			}
		} else if _, hasTMParam, hasTMName := isTestMainFunc(decl); hasTMName {
			tmInfo.testMainFile = &filePath
			tmInfo.testMainParam = hasTMParam
		}
	}

	if !hasImports {
		importDeclaration := getImportDeclaration()
		importDeclaration.Specs = append(importDeclaration.Specs, getAgentImportSpec())
		fileParser.Decls = append([]ast.Decl{importDeclaration}, fileParser.Decls...)
	}

	if dirty {
		//fmt.Printf("Updating %s file.\n", filePath)
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		err = printer.Fprint(f, fSet, fileParser)
		if err != nil {
			return err
		}
	}
	return nil
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
			importFound := false
			for _, impSpec := range genDcl.Specs {
				importSentence := impSpec.(*ast.ImportSpec)
				if importSentence.Path.Value == ImportPath {
					currentImportName = importSentence.Name.Name
					continue
				}
				if importSentence.Path.Value == "\"os\"" {
					importFound = true
					continue
				}
			}
			if !importFound {
				genDcl.Specs = append(genDcl.Specs, getOsImportSpec())
			}
		} else if fDecl, hTmParams, hasTmName := isTestMainFunc(decl); hasTmName {
			testMainFunc = fDecl
			hasTMParams = hTmParams
		}
	}

	if !hasImports {
		importDeclaration := getImportDeclaration()
		importDeclaration.Specs = append(importDeclaration.Specs, getAgentImportSpec(), getOsImportSpec())
		fileParser.Decls = append([]ast.Decl{importDeclaration}, fileParser.Decls...)
	}

	if testMainFunc == nil {
		fileParser.Decls = append(fileParser.Decls, getTestMainFunc(currentImportName))
	} else {
		_ = hasTMParams
		if !testMainHasGlobalAgent(testMainFunc, currentImportName) {
			fmt.Printf("\tPackage '%s' has a TestMain func already in '%s', please modify the file manually.\n",
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
