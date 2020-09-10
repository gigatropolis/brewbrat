
set GOOS=js
set GOARCH=wasm

go build -o ..\..\assets\json.wasm

set GOOS=windows
set GOARCH=amd64
