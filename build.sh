# Build Linux & Windows binaries
export GOOS=linux
go build -o goedit editor.go
export GOOS=windows
go build -o goedit.exe editor.go
