package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"text/template"

	_ "go.undefinedlabs.com/scopeagent"
)

type testPackageInfo struct {
	Name             string
	Folder           string
	MainFile         *string
	InstrumentedFile string
	Instrumented     bool
	SkippedFile      string
	Skipped          bool
}

var testMainFileTemplate, _ = template.New("testMain").Parse(
	`package {{.Name}}

import (
	_ "go.undefinedlabs.com/scopeagent/autoinstrument"
)
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
		processFolder(*folder)
	}
}

func processFolder(folder string) {
	testPackageInfoMap := map[string]*testPackageInfo{}
	fmt.Printf("Processing: %s\n", folder)
	err := processFolderTestFiles(folder, processTestFile, &testPackageInfoMap)
	if err != nil {
		fmt.Println(err)
	}
	for _, tpi := range testPackageInfoMap {

		if tpi.Instrumented {
			fmt.Printf("[SKIPPED] package '%v' is already instrumented in %v\n", tpi.Name, tpi.InstrumentedFile)
		} else if tpi.Skipped {
			fmt.Printf("[SKIPPED] package '%v' is importing %v in %v\n", tpi.Name, ImportPath, tpi.SkippedFile)
		} else {

			fileName := fmt.Sprintf("scope_pkg_%v_test.go", tpi.Name)
			mFile := path.Join(tpi.Folder, fileName)
			fmt.Printf("Auto instrumenting package %v in %v.\n", tpi.Name, mFile)

			file, errFile := os.Create(mFile)
			if errFile != nil {
				log.Fatalf("execution failed: %s", errFile)
			}
			writer := bufio.NewWriter(file)
			err := testMainFileTemplate.Execute(writer, tpi)
			if err != nil {
				log.Fatalf("execution failed: %s", err)
			}
			writer.Flush()
			file.Close()
		}
	}
	fmt.Println("Done.")
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

func processTestFile(filePath string, state interface{}) error {
	fSet := token.NewFileSet()
	fileParser, err := parser.ParseFile(fSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	tmInfoMap := *(state.(*map[string]*testPackageInfo))
	packageName := fileParser.Name.Name
	folder := path.Dir(filePath)
	key := fmt.Sprintf("%s.%s", folder, packageName)
	tpi := tmInfoMap[key]
	if tpi == nil {
		tpi = &testPackageInfo{
			Name:         packageName,
			Folder:       folder,
			MainFile:     nil,
			Instrumented: false,
			Skipped:      false,
		}
		tmInfoMap[key] = tpi
	}

	for _, decl := range fileParser.Decls {
		if mainFunc, _, hasTMName := isTestMainFunc(decl); hasTMName {
			tpi.MainFile = &filePath
			if testMainHasGlobalAgent(mainFunc) {
				tpi.Instrumented = true
				tpi.InstrumentedFile = filePath
			}
		}
		if genDcl, ok := isImportDeclaration(decl); ok {
			for _, impSpec := range genDcl.Specs {
				importSentence := impSpec.(*ast.ImportSpec)
				if importSentence.Path.Value == AutoInstrumentationImportPath {
					tpi.Instrumented = true
					tpi.InstrumentedFile = filePath
				}
				if importSentence.Path.Value == ImportPath {
					tpi.SkippedFile = filePath
					tpi.Skipped = true
				}
			}
		}
	}
	return nil
}
