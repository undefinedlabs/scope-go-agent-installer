# scope-go-agent-installer

Tool to install the scope go agent in an existing project by adding/modifying the TestMain func with `scopeagent.Run` call in all packages, as seen in the documentation: https://docs.scope.dev/docs/go-installation

### Usage:
```
scope-go-agent-installer -folder={PROJECT FOLDER}
```

### Example:

#### Public Repo: https://github.com/gin-gonic/gin
```
git clone git@github.com:gin-gonic/gin.git
cd gin
scope-go-agent-installer -folder=.
go test ./...
```
