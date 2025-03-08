package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"
)

// Пути к файлам спецификаций OpenAPI
const (
	BaseSpecPath     = "E:\\Workspace\\openapidiff\\ghes-3.8.json"
	RevisionSpecPath = "E:\\Workspace\\openapidiff\\ghes-3.9.json"
)

// BenchmarkResult содержит результаты бенчмарка
type BenchmarkResult struct {
	LoadTime      time.Duration
	DiffTime      time.Duration
	MemoryUsage   uint64
	ChangesCount  int
	EndpointsBase int
	EndpointsRev  int
}

func main() {
	fmt.Println("==== OAS Diff Benchmark Tool ====")

	// Проверяем аргументы командной строки
	if len(os.Args) > 1 && os.Args[1] == "advanced" {
		runAdvancedBenchmark()
	} else {
		runSimpleBenchmark()
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
	}

	// Выводим результаты
	printAdvancedResults(benchResult)
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
	fmt.Printf("Base endpoints count:    %d\n", result.EndpointsBase)
	fmt.Printf("Revision endpoints count: %d\n", result.EndpointsRev)
	fmt.Printf("Total changes detected:   %d\n", result.ChangesCount)
	fmt.Printf("Load time:                %v\n", result.LoadTime)
	fmt.Printf("Diff computation time:    %v\n", result.DiffTime)
	fmt.Printf("Total processing time:    %v\n", result.LoadTime+result.DiffTime)
	fmt.Printf("Memory usage:             %.2f MB\n", float64(result.MemoryUsage)/(1024*1024))
	fmt.Printf("===================================\n")
}
