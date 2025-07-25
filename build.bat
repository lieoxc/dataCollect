set GOOS=linux
set GOARCH=arm64
go mod tidy
go build -ldflags="-s -w" -o data_collect