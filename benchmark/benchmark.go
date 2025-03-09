package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"
)

// Пути к файлам спецификаций OpenAPI
var (
	BaseSpecPath     = "E:\\Workspace\\openapidiff\\ghes-3.8.json"
	RevisionSpecPath = "E:\\Workspace\\openapidiff\\ghes-3.9.json"
)

// JMHCompatBenchmark структура для JMH-совместимых бенчмарков
type JMHCompatBenchmark struct {
	Name                  string
	WarmupIterations      int
	MeasurementIterations int
	Forks                 int
	IterationDuration     time.Duration
	OutputFile            string
	CompareJMHFile        string
	DescriptionFields     map[string]string
}

// BenchmarkResult содержит результаты бенчмарка
type BenchmarkResult struct {
	LoadTime      time.Duration
	DiffTime      time.Duration
	MemoryUsage   uint64
	ChangesCount  int
	EndpointsBase int
	EndpointsRev  int
	ThroughputOps float64 // Операций в секунду (для совместимости с JMH)
}

func main() {
	// Определяем флаги командной строки
	baseSpecPathFlag := flag.String("base", BaseSpecPath, "Путь к базовой спецификации OpenAPI")
	revSpecPathFlag := flag.String("revision", RevisionSpecPath, "Путь к пересмотренной спецификации OpenAPI")
	outputFileFlag := flag.String("output", "", "Файл для сохранения результатов")
	compareJMHFlag := flag.String("compare", "", "Файл JMH для сравнения")
	warmupFlag := flag.Int("warmup", 5, "Количество итераций прогрева")
	measurementFlag := flag.Int("measurement", 5, "Количество итераций измерения")
	forksFlag := flag.Int("forks", 5, "Количество форков (полных запусков)")
	durationFlag := flag.Int("duration", 10, "Продолжительность итерации в секундах")
	modeFlag := flag.String("mode", "jmh", "Режим бенчмарка: simple, advanced, jmh")
	jmhNameFlag := flag.String("name", "org.oasdiff.Benchmark", "Имя бенчмарка в формате JMH")

	flag.Parse()

	// Обновляем пути к файлам спецификаций из флагов
	if *baseSpecPathFlag != BaseSpecPath {
		BaseSpecPath = *baseSpecPathFlag
	}
	if *revSpecPathFlag != RevisionSpecPath {
		RevisionSpecPath = *revSpecPathFlag
	}

	// Перенаправляем вывод в файл, если указан
	if *outputFileFlag != "" {
		file, err := os.Create(*outputFileFlag)
		if err != nil {
			log.Fatalf("Не удалось создать файл вывода: %v", err)
		}
		defer file.Close()
		log.SetOutput(file)
	}

	// Выбираем режим бенчмарка
	switch *modeFlag {
	case "simple":
		runSimpleBenchmark()
	case "advanced":
		runAdvancedBenchmark()
	case "jmh":
		// Создаем JMH-совместимый бенчмарк
		benchmark := &JMHCompatBenchmark{
			Name:                  *jmhNameFlag,
			WarmupIterations:      *warmupFlag,
			MeasurementIterations: *measurementFlag,
			Forks:                 *forksFlag,
			IterationDuration:     time.Duration(*durationFlag) * time.Second,
			OutputFile:            *outputFileFlag,
			CompareJMHFile:        *compareJMHFlag,
			DescriptionFields:     make(map[string]string),
		}

		// Добавляем дополнительные описания
		benchmark.DescriptionFields["Base Spec"] = BaseSpecPath
		benchmark.DescriptionFields["Revision Spec"] = RevisionSpecPath

		// Запускаем JMH-совместимый бенчмарк
		runJMHCompatBenchmark(benchmark)
	default:
		log.Fatalf("Неизвестный режим бенчмарка: %s", *modeFlag)
	}
}

// runSimpleBenchmark выполняет базовый бенчмарк с несколькими итерациями
func runSimpleBenchmark() {
	fmt.Println("Starting basic OAS diff benchmark...")
	fmt.Printf("Comparing:\n  Base: %s\n  Revision: %s\n\n", BaseSpecPath, RevisionSpecPath)

	// Загружаем спецификации
	fmt.Println("Loading specifications...")
	loadStart := time.Now()

	// Создаем загрузчик для OpenAPI спецификаций
	loader := openapi3.NewLoader()

	// Загрузка базовой спецификации
	baseSpec, err := load.NewSpecInfo(loader, load.NewSource(BaseSpecPath))
	if err != nil {
		log.Fatalf("Error loading base spec: %v", err)
	}

	// Загрузка новой спецификации
	revisionSpec, err := load.NewSpecInfo(loader, load.NewSource(RevisionSpecPath))
	if err != nil {
		log.Fatalf("Error loading revision spec: %v", err)
	}

	loadDuration := time.Since(loadStart)
	fmt.Printf("Specifications loaded in %v\n\n", loadDuration)

	// Создаем конфигурацию для diff
	config := diff.NewConfig()

	// Прогрев
	fmt.Println("Warming up...")
	_, _ = diff.Get(config, baseSpec.Spec, revisionSpec.Spec)

	// Запускаем 5 итераций для получения среднего значения
	const iterations = 5
	var totalDuration time.Duration

	fmt.Printf("Running %d iterations:\n", iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()

		// Сравниваем спецификации
		diffResult, err := diff.Get(config, baseSpec.Spec, revisionSpec.Spec)
		if err != nil {
			log.Printf("Error in iteration %d: %v", i+1, err)
			continue
		}

		iterDuration := time.Since(start)
		totalDuration += iterDuration

		// Подсчет изменений для текущей итерации
		changes := 0
		if diffResult != nil && diffResult.PathsDiff != nil {
			changes = countChanges(diffResult.PathsDiff)
		}

		// Вывод статистики итерации
		fmt.Printf("  Iteration %d: %v - Found %d changes\n", i+1, iterDuration, changes)
	}

	// Вывод итоговой статистики
	avgDuration := totalDuration / iterations
	fmt.Printf("\nAverage execution time: %v\n", avgDuration)
}

// runAdvancedBenchmark выполняет расширенный бенчмарк с подробной информацией
func runAdvancedBenchmark() {
	fmt.Println("Starting advanced OAS diff benchmark...")
	fmt.Printf("Comparing:\n  Base: %s\n  Revision: %s\n\n", BaseSpecPath, RevisionSpecPath)

	var memStats runtime.MemStats

	// Запускаем сборку мусора перед началом измерений
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMemory := memStats.Alloc

	// Загружаем спецификации и замеряем время
	loadStart := time.Now()

	// Создаем загрузчик для OpenAPI спецификаций
	loader := openapi3.NewLoader()

	// Загрузка базовой спецификации
	baseSpec, err := load.NewSpecInfo(loader, load.NewSource(BaseSpecPath))
	if err != nil {
		log.Fatalf("Error loading base spec: %v", err)
	}

	// Загрузка новой спецификации
	revisionSpec, err := load.NewSpecInfo(loader, load.NewSource(RevisionSpecPath))
	if err != nil {
		log.Fatalf("Error loading revision spec: %v", err)
	}

	loadDuration := time.Since(loadStart)

	// Подсчитываем количество endpoints в каждой спецификации
	baseEndpoints := countEndpoints(baseSpec.Spec)
	revEndpoints := countEndpoints(revisionSpec.Spec)

	// Запускаем diff и замеряем время
	diffStart := time.Now()

	// Сравниваем спецификации
	result, err := diff.Get(diff.NewConfig(), baseSpec.Spec, revisionSpec.Spec)
	if err != nil {
		log.Fatalf("Error comparing specs: %v", err)
	}
	diffDuration := time.Since(diffStart)

	// Измеряем использование памяти после выполнения
	runtime.ReadMemStats(&memStats)
	memoryUsed := memStats.Alloc - initialMemory

	// Подсчитываем количество изменений
	changesCount := 0
	if result != nil && result.PathsDiff != nil {
		changesCount = countChanges(result.PathsDiff)
	}

	// Формируем результат
	benchResult := BenchmarkResult{
		LoadTime:      loadDuration,
		DiffTime:      diffDuration,
		MemoryUsage:   memoryUsed,
		ChangesCount:  changesCount,
		EndpointsBase: baseEndpoints,
		EndpointsRev:  revEndpoints,
		ThroughputOps: 1.0 / diffDuration.Seconds(), // Примерная пропускная способность
	}

	// Выводим результаты
	printAdvancedResults(benchResult)
}

// runJMHCompatBenchmark выполняет бенчмарк в стиле JMH
func runJMHCompatBenchmark(bench *JMHCompatBenchmark) {
	printJMHHeader(bench)

	// Загружаем спецификации один раз для всех запусков
	loader := openapi3.NewLoader()

	baseSpec, err := load.NewSpecInfo(loader, load.NewSource(BaseSpecPath))
	if err != nil {
		log.Fatalf("Error loading base spec: %v", err)
	}

	revisionSpec, err := load.NewSpecInfo(loader, load.NewSource(RevisionSpecPath))
	if err != nil {
		log.Fatalf("Error loading revision spec: %v", err)
	}

	// Массив для хранения всех результатов
	allResults := make([]float64, 0, bench.Forks*bench.MeasurementIterations)

	// Запускаем указанное количество форков
	for fork := 1; fork <= bench.Forks; fork++ {
		fmt.Printf("# Run progress: %.2f%% complete, ETA %s\n",
			float64(fork-1)*100.0/float64(bench.Forks),
			formatETA(fork-1, bench))

		fmt.Printf("# Fork: %d of %d\n", fork, bench.Forks)

		// Прогрев (warmup)
		for i := 1; i <= bench.WarmupIterations; i++ {
			// Запускаем сборку мусора перед каждой итерацией
			runtime.GC()

			// Замеряем производительность
			opsPerSecond := runDiffIteration(baseSpec, revisionSpec, bench.IterationDuration)

			fmt.Printf("# Warmup Iteration %3d: %.3f ops/s\n", i, opsPerSecond)
		}

		// Измерения (measurement)
		for i := 1; i <= bench.MeasurementIterations; i++ {
			// Запускаем сборку мусора перед каждой итерацией
			runtime.GC()

			// Замеряем производительность
			opsPerSecond := runDiffIteration(baseSpec, revisionSpec, bench.IterationDuration)

			fmt.Printf("Iteration %3d: %.3f ops/s\n", i, opsPerSecond)
			allResults = append(allResults, opsPerSecond)
		}

		fmt.Println()
	}

	// Выводим результаты в формате JMH
	printJMHResults(bench, allResults)
}

// runDiffIteration выполняет одну итерацию diff и возвращает операций в секунду
func runDiffIteration(baseSpec *load.SpecInfo, revisionSpec *load.SpecInfo, duration time.Duration) float64 {
	config := diff.NewConfig()
	startTime := time.Now()
	endTime := startTime.Add(duration)

	operationCount := 0

	// Выполняем операции diff до истечения времени
	for time.Now().Before(endTime) {
		_, err := diff.Get(config, baseSpec.Spec, revisionSpec.Spec)
		if err != nil {
			log.Printf("Ошибка при выполнении diff: %v", err)
			continue
		}
		operationCount++
	}

	actualDuration := time.Since(startTime)
	return float64(operationCount) / actualDuration.Seconds()
}

// Подсчет изменений в результате diff
func countChanges(pathsDiff *diff.PathsDiff) int {
	count := 0

	// Подсчет добавленных путей
	count += len(pathsDiff.Added)

	// Подсчет удаленных путей
	count += len(pathsDiff.Deleted)

	// Подсчет измененных путей
	for _, pathDiff := range pathsDiff.Modified {
		// Для каждого измененного пути считаем изменения в операциях
		if pathDiff.OperationsDiff != nil {
			count += len(pathDiff.OperationsDiff.Added)
			count += len(pathDiff.OperationsDiff.Deleted)
			count += len(pathDiff.OperationsDiff.Modified)
		}
	}

	return count
}

// Подсчет количества эндпоинтов в спецификации
func countEndpoints(spec *openapi3.T) int {
	count := 0
	// Перебираем все пути в спецификации
	for _, pathItem := range spec.Paths.Map() {
		// Подсчитываем все HTTP методы для каждого пути
		if pathItem.Get != nil {
			count++
		}
		if pathItem.Post != nil {
			count++
		}
		if pathItem.Put != nil {
			count++
		}
		if pathItem.Delete != nil {
			count++
		}
		if pathItem.Options != nil {
			count++
		}
		if pathItem.Head != nil {
			count++
		}
		if pathItem.Patch != nil {
			count++
		}
	}
	return count
}

// Вывод результатов расширенного бенчмарка
func printAdvancedResults(result BenchmarkResult) {
	fmt.Println("\n=== Advanced Benchmark Results ===")
	fmt.Printf("Base endpoints count:     %d\n", result.EndpointsBase)
	fmt.Printf("Revision endpoints count: %d\n", result.EndpointsRev)
	fmt.Printf("Total changes detected:   %d\n", result.ChangesCount)
	fmt.Printf("Load time:                %v\n", result.LoadTime)
	fmt.Printf("Diff computation time:    %v\n", result.DiffTime)
	fmt.Printf("Total processing time:    %v\n", result.LoadTime+result.DiffTime)
	fmt.Printf("Memory usage:             %.2f MB\n", float64(result.MemoryUsage)/(1024*1024))
	fmt.Printf("Throughput:               %.3f ops/s\n", result.ThroughputOps)
	fmt.Printf("===================================\n")
}

// printJMHHeader выводит заголовок в формате JMH
func printJMHHeader(bench *JMHCompatBenchmark) {
	fmt.Println("# JMH-compatible Go Benchmark Framework")
	fmt.Printf("# Go version: %s\n", runtime.Version())
	fmt.Printf("# GOOS: %s, GOARCH: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("# CPU: %d, %s\n", runtime.NumCPU(), runtime.GOARCH)

	// Выводим дополнительные поля описания
	for key, value := range bench.DescriptionFields {
		fmt.Printf("# %s: %s\n", key, value)
	}

	fmt.Printf("# Warmup: %d iterations, %v each\n", bench.WarmupIterations, bench.IterationDuration)
	fmt.Printf("# Measurement: %d iterations, %v each\n", bench.MeasurementIterations, bench.IterationDuration)
	fmt.Printf("# Timeout: 10 min per iteration\n")
	fmt.Printf("# Threads: 1 thread\n")
	fmt.Printf("# Benchmark mode: Throughput, ops/time\n")
	fmt.Printf("# Benchmark: %s\n", bench.Name)
	fmt.Println()
}

// formatETA форматирует оставшееся время выполнения
func formatETA(completedForks int, bench *JMHCompatBenchmark) string {
	totalTimePerFork := time.Duration(bench.WarmupIterations+bench.MeasurementIterations) * bench.IterationDuration
	remainingTime := time.Duration(bench.Forks-completedForks) * totalTimePerFork

	hours := int(remainingTime.Hours())
	minutes := int(remainingTime.Minutes()) % 60
	seconds := int(remainingTime.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// printJMHResults выводит результаты в формате JMH
func printJMHResults(bench *JMHCompatBenchmark, results []float64) {
	// Рассчитываем статистику
	avg := calculateMean(results)
	min, max := calculateMinMax(results)
	stdev := calculateStdDev(results, avg)

	// Рассчитываем погрешность (99.9% доверительный интервал)
	tValue := 3.5 // Примерно для 99.9% CI
	errorValue := (tValue * stdev) / math.Sqrt(float64(len(results)))

	fmt.Printf("\nResult \"%s\":\n", bench.Name)
	fmt.Printf("  %.3f ±(99.9%%) %.3f ops/s [Average]\n", avg, errorValue)
	fmt.Printf("  (min, avg, max) = (%.3f, %.3f, %.3f), stdev = %.3f\n", min, avg, max, stdev)
	fmt.Printf("  CI (99.9%%): [%.3f, %.3f] (assumes normal distribution)\n", avg-errorValue, avg+errorValue)
	fmt.Println()

	totalTime := formatTotalTime(bench)
	fmt.Printf("\n# Run complete. Total time: %s\n", totalTime)

	fmt.Println("\nREMEMBER: The numbers below are just data. To gain reusable insights, you need to follow up on")
	fmt.Println("why the numbers are the way they are. Use profilers, design factorial experiments, make sure")
	fmt.Println("the benchmarking environment is safe on JVM/OS/HW level, ask for reviews from the domain experts.")
	fmt.Println("Do not assume the numbers tell you what you want them to tell.")
	fmt.Println()

	fmt.Println("Benchmark         Mode  Cnt  Score   Error  Units")
	// Получаем короткое имя бенчмарка
	benchmarkShortName := bench.Name
	if lastDot := strings.LastIndex(benchmarkShortName, "."); lastDot != -1 {
		packageName := benchmarkShortName[:lastDot]
		shortName := benchmarkShortName[lastDot+1:]
		fmt.Printf("%-10s.%-10s thrpt %4d %6.3f ± %5.3f  ops/s\n",
			packageName, shortName, len(results), avg, errorValue)
	} else {
		fmt.Printf("%-21s thrpt %4d %6.3f ± %5.3f  ops/s\n",
			benchmarkShortName, len(results), avg, errorValue)
	}
}

// formatTotalTime форматирует общее время выполнения бенчмарка
func formatTotalTime(bench *JMHCompatBenchmark) string {
	totalTimeNano := int64(bench.Forks) * int64(bench.WarmupIterations+bench.MeasurementIterations) * bench.IterationDuration.Nanoseconds()
	totalTime := time.Duration(totalTimeNano)

	hours := int(totalTime.Hours())
	minutes := int(totalTime.Minutes()) % 60
	seconds := int(totalTime.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// Статистические функции
func calculateMean(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func calculateMinMax(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}

	min := data[0]
	max := data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func calculateVariance(data []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += math.Pow(v-mean, 2)
	}
	return sum / float64(len(data))
}

func calculateStdDev(data []float64, mean float64) float64 {
	return math.Sqrt(calculateVariance(data, mean))
}
