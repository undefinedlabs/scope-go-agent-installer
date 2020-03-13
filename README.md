# scope-go-agent-installer

Tool to install the scope go agent in an existing project by adding a file with the import autoinstrumentation in each package containing tests, as seen in the documentation: https://docs.scope.dev/docs/go-installation

### Installation:
```bash
go install github.com/undefinedlabs/scope-go-agent-installer
```

This build and copy the binary to: `~/go/bin/scope-go-agent-installer`

### Usage:
```bash
scope-go-agent-installer -folder={PROJECT FOLDER}
```

### Example:

#### Public Repo: https://github.com/gin-gonic/gin
```bash
git clone git@github.com:gin-gonic/gin.git
cd gin
scope-go-agent-installer -folder=.
go test ./...
```
