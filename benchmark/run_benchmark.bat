@echo off
echo === OASDIFF Benchmark Tool ===
echo.

echo Building benchmark tool...
go build -o oasdiff-benchmark.exe .
if %ERRORLEVEL% neq 0 (
  echo Build failed!
  exit /b %ERRORLEVEL%
)
echo Build successful.
echo.

echo === Running Basic Benchmark ===
oasdiff-benchmark.exe
echo.

echo === Running Advanced Benchmark ===
oasdiff-benchmark.exe advanced
echo.

echo Benchmark run completed!
