package console

import (
	"bufio"
	"bytes"
	"fmt"
	"sync"
	"testing"
)

func TestLoggerConcurrentWritesAreNotInterleaved(t *testing.T) {
	var output bytes.Buffer
	logger := NewLogger(&output)
	const workers = 64

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			logger.Info("log-%d", i)
			_, _ = fmt.Fprintf(logger.StderrWriter(), "stderr-%d\n", i)
		}(i)
	}
	wg.Wait()

	lines := make(map[string]bool, workers*2)
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		lines[scanner.Text()] = true
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(lines) != workers*2 {
		t.Fatalf("got %d distinct output lines, want %d", len(lines), workers*2)
	}
	for i := 0; i < workers; i++ {
		for _, line := range []string{fmt.Sprintf("log-%d", i), fmt.Sprintf("stderr-%d", i)} {
			if !lines[line] {
				t.Errorf("output line %q is missing or corrupted", line)
			}
		}
	}
}
