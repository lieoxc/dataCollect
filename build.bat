set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o data_collect