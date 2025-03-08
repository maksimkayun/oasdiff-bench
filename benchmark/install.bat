@echo off
echo Building benchmark tool...
go build -o oasdiff-benchmark.exe .
echo Build complete.
echo.
echo To run basic benchmark: oasdiff-benchmark.exe
echo To run advanced benchmark: oasdiff-benchmark.exe advanced
