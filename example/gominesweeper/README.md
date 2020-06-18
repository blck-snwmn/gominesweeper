## Prepare
using wasm_exec.js

curl -so ./example/gominesiper/wasm_exec.js https://raw.githubusercontent.com/golang/go/master/misc/wasm/wasm_exec.js

## Builds
in ./exaple/gominesiper/gominesiper.go
GOOS=js GOARCH=wasm go build -o main.wasm

## Run
in ./exaple/gominesiper/gominesiper.go
goexec 'http.ListenAndServe(":8888", http.FileServer(http.Dir(".")))'

## debug
if you use VSCode, you can create `example/gominesiper/.vscode/settings.json`.

```
{
    "go.toolsEnvVars": {
        "GOOS": "js",
        "GOARCH": "wasm",
    }
}
```