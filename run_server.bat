@echo off

REM Check if 'go' is installed
where go > nul 2>&1
if %errorlevel% neq 0 (
    echo Go is not installed. Installing...
    REM Install Go
    winget install Golang.Go -e -h
)

REM Run the Go program
echo Running Go program...
go mod tidy
go run ./cmd/jiotv_go/

echo Go program completed.
