# Folder Web

This is an example to show how Folder may be used on the Web.

## Setup

1. Compile the WebAssembly binary by running:
```sh
GOOS=js GOARCH=wasm go build -o folder.wasm
```
2. Now the `wasm_exec.js` file need to be copied from Go source code by running:
```sh
cp $(go env GOROOT)/misc/wasm/wasm_exec.js .
```
3. Copy an existing index directory to this directory and name it `index`.
4. Run a web server that host this directory (e.g. Python's web server):
```sh
python3 -m http.server
```
5. Now you should be able to go to http://localhost:8000 with a WebAssembly-capable browser and see the example in action!