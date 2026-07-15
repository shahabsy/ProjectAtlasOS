@echo off
echo Setting up Atlas Phase 0 environment...

go mod tidy >nul

if %errorLevel% neq 0 (
    echo go mod tidy failed. Ensure go is in PATH
    exit /b 1
)

echo Setup complete!
echo Run: go build -o atlas.exe .