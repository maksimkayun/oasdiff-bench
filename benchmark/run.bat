@echo off
echo === OASDIFF Benchmark Tool ===
echo.

echo Building and running benchmark...
go run benchmark.go
echo.

echo Running advanced benchmark...
go run benchmark.go advanced
echo.

echo All benchmarks completed.
