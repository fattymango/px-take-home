package logreader

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/pkg/logger"
)

const (
	testTaskID = uint64(1234)
	numLines   = 500000 // 1M lines for aggressive benchmarks
	lineSize   = 200    // Make each line longer for more realistic data
)

var (
	testConfig *config.Config
	testLogger *logger.Logger
	tempDir    string
)

func setupBenchmark(b *testing.B) {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all CPU cores
	var err error

	tempDir, err = os.MkdirTemp("", "reader_benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}

	testConfig = &config.Config{
		TaskLogger: config.TaskLogger{
			DirPath: tempDir,
		},
	}

	testLogger, err = logger.NewLogger(testConfig)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	logFile := filepath.Join(tempDir, fmt.Sprintf("%d.log", testTaskID))
	f, err := os.Create(logFile)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}
	defer f.Close()

	padding := make([]byte, lineSize)
	for i := range padding {
		padding[i] = 'x'
	}
	for i := 0; i < numLines; i++ {
		_, err := fmt.Fprintf(f, "Test log line %d with additional padding: %s\n", i, padding)
		if err != nil {
			b.Fatalf("Failed to write test data: %v", err)
		}
	}
}

func cleanupBenchmark() {
	os.RemoveAll(tempDir)
}

func runReaderBenchmark(b *testing.B, reader LogReader) {
	b.Run("Latest100Lines", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _, _ = reader.Read(0, 0)
			}
		})
	})

	b.Run("First100Lines", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _, _ = reader.Read(1, 100)
			}
		})
	})

	b.Run("Last100Lines", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _, _ = reader.Read(numLines-100, numLines)
			}
		})
	})

	b.Run("Middle100Lines", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _, _ = reader.Read(numLines/2, numLines/2+100)
			}
		})
	})

	b.Run("RandomAccess", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				start := int(float64(numLines) * (float64(b.N%100) / 100.0))
				if start < 1 {
					start = 1
				}
				end := start + 100
				if end > numLines {
					end = numLines
				}
				_, _, _ = reader.Read(start, end)
			}
		})
	})
}

func BenchmarkTailHeadReader(b *testing.B) {
	setupBenchmark(b)
	defer cleanupBenchmark()
	reader := NewTailHeadReader(testConfig, testLogger, testTaskID)
	b.ResetTimer()
	runReaderBenchmark(b, reader)
}

func BenchmarkAwkReader(b *testing.B) {
	setupBenchmark(b)
	defer cleanupBenchmark()
	reader := NewAwkReader(testConfig, testLogger, testTaskID)
	b.ResetTimer()
	runReaderBenchmark(b, reader)
}

func BenchmarkSedReader(b *testing.B) {
	setupBenchmark(b)
	defer cleanupBenchmark()
	reader := NewSedReader(testConfig, testLogger, testTaskID)
	b.ResetTimer()
	runReaderBenchmark(b, reader)
}

func BenchmarkBufferReader(b *testing.B) {
	setupBenchmark(b)
	defer cleanupBenchmark()
	reader := NewBufferReader(testConfig, testLogger, testTaskID)
	b.ResetTimer()
	runReaderBenchmark(b, reader)
}
