package logger

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
	}
	return sb.String()
}

func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	f()
	w.Close()
	os.Stderr = old
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
	}
	return sb.String()
}

func TestInfoLogging(t *testing.T) {
	l := NewDefaultLogger()
	l.SetLevel(debugger.LevelInfo)
	out := captureStdout(func() {
		l.Info("hello", map[string]interface{}{"k": "v"})
	})
    if !strings.Contains(out, "info") || !strings.Contains(out, "hello") {
        t.Fatalf("expected INFO log with message, got: %s", out)
    }
}

func TestWarnFilter(t *testing.T) {
	l := NewDefaultLogger()
	l.SetLevel(debugger.LevelWarn)
	outInfo := captureStdout(func() { l.Info("i") })
	if outInfo != "" {
		t.Fatalf("expected no info output at warn level, got: %s", outInfo)
	}
	outWarn := captureStderr(func() { l.Warn("w") })
    if !strings.Contains(outWarn, "warn") {
        t.Fatalf("expected WARN output, got: %s", outWarn)
    }
}

func TestErrorToStderr(t *testing.T) {
	l := NewDefaultLogger()
	l.SetLevel(debugger.LevelError)
	outErr := captureStderr(func() { l.Error("e") })
    if !strings.Contains(outErr, "error") {
        t.Fatalf("expected ERROR output, got: %s", outErr)
    }
	outInfo := captureStdout(func() { l.Info("i") })
	if outInfo != "" {
		t.Fatalf("expected no info output at error level, got: %s", outInfo)
	}
}

// ---- Benchmarks ----

// simulate old string-based shouldLog
func oldShouldLog(current string, level string) bool {
	levelOrder := map[string]int{
		"debug": 1,
		"info":  2,
		"warn":  3,
		"error": 4,
	}
	cur := levelOrder[current]
	lv := levelOrder[level]
	return lv >= cur
}

func BenchmarkShouldLogNew(b *testing.B) {
	l := NewDefaultLogger()
	l.SetLevel(debugger.LevelWarn)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.shouldLog(debugger.LevelError)
		_ = l.shouldLog(debugger.LevelWarn)
		_ = l.shouldLog(debugger.LevelInfo)
	}
}

func BenchmarkShouldLogOld(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = oldShouldLog("warn", "error")
		_ = oldShouldLog("warn", "warn")
		_ = oldShouldLog("warn", "info")
	}
}