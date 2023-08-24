@echo off

REM Check if 'go' is installed
where go > nul 2>&1
if %errorlevel% neq 0 (
    echo Go is not installed. Installing...
    REM Install Go (assuming you have chocolatey installed)
    winget install Golang.Go -e -h -q
)

REM Run the Go program
echo Running Go program...
go mod tidy
go run .

echo Go program completed.
