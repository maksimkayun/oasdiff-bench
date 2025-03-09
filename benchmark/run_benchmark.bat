@echo off
setlocal EnableDelayedExpansion

echo === OASDIFF JMH-compatible Benchmark Tool ===
echo.

:: Проверяем, указан ли файл для сравнения
set COMPARE_JMH=
set OUTPUT_FILE=
set WARMUP=5
set MEASUREMENT=5
set FORKS=5
set DURATION=10
set BASE_SPEC=
set REV_SPEC=
set MODE=jmh
set NAME=org.oasdiff.Benchmark

:: Обработка параметров
:params_loop
if "%~1"=="" goto params_done
if /i "%~1"=="--compare" (
  set COMPARE_JMH=%~2
  shift
) else if /i "%~1"=="--output" (
  set OUTPUT_FILE=%~2
  shift
) else if /i "%~1"=="--warmup" (
  set WARMUP=%~2
  shift
) else if /i "%~1"=="--measurement" (
  set MEASUREMENT=%~2
  shift
) else if /i "%~1"=="--forks" (
  set FORKS=%~2
  shift
) else if /i "%~1"=="--duration" (
  set DURATION=%~2
  shift
) else if /i "%~1"=="--base" (
  set BASE_SPEC=%~2
  shift
) else if /i "%~1"=="--rev" (
  set REV_SPEC=%~2
  shift
) else if /i "%~1"=="--mode" (
  set MODE=%~2
  shift
) else if /i "%~1"=="--name" (
  set NAME=%~2
  shift
)
shift
goto params_loop
:params_done

echo Building benchmark tool...
go build -o oasdiff-benchmark.exe .
if %ERRORLEVEL% neq 0 (
  echo Build failed!
  exit /b %ERRORLEVEL%
)
echo Build successful.
echo.

:: Формируем аргументы
set ARGS=-mode=%MODE% -name=%NAME% -warmup=%WARMUP% -measurement=%MEASUREMENT% -forks=%FORKS% -duration=%DURATION%

if not "%BASE_SPEC%"=="" (
  set ARGS=!ARGS! -base="%BASE_SPEC%"
)

if not "%REV_SPEC%"=="" (
  set ARGS=!ARGS! -revision="%REV_SPEC%"
)

if not "%OUTPUT_FILE%"=="" (
  set ARGS=!ARGS! -output="%OUTPUT_FILE%"
)

if not "%COMPARE_JMH%"=="" (
  set ARGS=!ARGS! -compare="%COMPARE_JMH%"
)

echo === Running %MODE% Benchmark ===
echo Parameters:
echo   Mode: %MODE%
echo   Benchmark Name: %NAME%
echo   Warmup Iterations: %WARMUP%
echo   Measurement Iterations: %MEASUREMENT%
echo   Forks: %FORKS%
echo   Duration per Iteration: %DURATION%s
if not "%BASE_SPEC%"=="" echo   Base Spec: %BASE_SPEC%
if not "%REV_SPEC%"=="" echo   Revision Spec: %REV_SPEC%
if not "%OUTPUT_FILE%"=="" echo   Output File: %OUTPUT_FILE%
if not "%COMPARE_JMH%"=="" echo   JMH Compare File: %COMPARE_JMH%
echo.

echo Running benchmark with arguments: %ARGS%
oasdiff-benchmark.exe %ARGS%
echo.

echo Benchmark run completed!

:: Если есть файл вывода, показываем сообщение
if not "%OUTPUT_FILE%"=="" (
  echo Results saved to: %OUTPUT_FILE%
)

endlocal