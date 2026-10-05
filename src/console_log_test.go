package opennox

import (
	"bytes"
	"testing"
)

type consoleLogTestWriter struct {
	bytes.Buffer
	closed bool
}

func (w *consoleLogTestWriter) Close() error { w.closed = true; return nil }

func TestConsoleLogSwitchFlushesOldDestination(t *testing.T) {
	if logFile != nil {
		t.Fatal("test requires an unconfigured log file")
	}
	t.Cleanup(closeLog)
	a, b := &consoleLogTestWriter{}, &consoleLogTestWriter{}
	setLogFile(a)
	_, _ = logBuf.WriteString("first destination\n")
	setLogFile(b)
	if a.String() != "first destination\n" || !a.closed || b.Len() != 0 {
		t.Fatalf("switch: old=%q closed=%t new=%q", a.String(), a.closed, b.String())
	}
	_, _ = logBuf.WriteString("second destination\n")
	closeLog()
	if b.String() != "second destination\n" || !b.closed || logFile != nil {
		t.Fatalf("stop: new=%q closed=%t file=%v", b.String(), b.closed, logFile)
	}
}
