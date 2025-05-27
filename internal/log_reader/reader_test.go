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
	numLines   = 500000 // Number of lines in test file
	lineSize   = 200    // Size of each line for more realistic data
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

	testLogger = logger.NewTestLogger()

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

func runReaderBenchmark(b *testing.B, reader Reader) {
	// Case 1: Default case - last 100 lines (from=0, to=0)
	b.Run("DefaultLastLines", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _, _ = reader.Read(0, 0)
			}
		})
	})

	// Case 2: Specific range with both from and to > 0
	// Test with different ranges to get a comprehensive view
	ranges := []struct {
		name     string
		from, to int
	}{
		{"SmallRange_10Percent", numLines / 10, numLines/10 + 100},  // 10% into file
		{"MidRange_25Percent", numLines / 4, numLines/4 + 500},      // 25% into file
		{"LargeRange_50Percent", numLines / 2, numLines/2 + 1000},   // 50% into file
		{"HugeRange_30to90Percent", numLines / 3, numLines / 3 * 2}, // 30% to 90% of file
		{"FullRange", 1, numLines},                                  // All lines
	}

	for _, r := range ranges {
		b.Run(fmt.Sprintf("SpecificRange_%s", r.name), func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, _, _ = reader.Read(r.from, r.to)
				}
			})
		})
	}
}

func BenchmarkTailHeadReader(b *testing.B) {
	setupBenchmark(b)
	defer cleanupBenchmark()
	reader := NewTailHeadReader(testConfig, testLogger, testTaskID)
	runReaderBenchmark(b, reader)
}

func BenchmarkSedReader(b *testing.B) {
	setupBenchmark(b)
	defer cleanupBenchmark()
	reader := NewSedReader(testConfig, testLogger, testTaskID)
	runReaderBenchmark(b, reader)
}

func BenchmarkAwkReader(b *testing.B) {
	setupBenchmark(b)
	defer cleanupBenchmark()
	reader := NewAwkReader(testConfig, testLogger, testTaskID)
	runReaderBenchmark(b, reader)
}

func BenchmarkBufferReader(b *testing.B) {
	setupBenchmark(b)
	defer cleanupBenchmark()
	reader := NewBufferReader(testConfig, testLogger, testTaskID)
	runReaderBenchmark(b, reader)
}
