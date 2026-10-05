package opennox

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/opennox/libs/env"
	"github.com/opennox/libs/log"
)

var (
	logFile    io.Closer
	logBuf     bufio.Writer
	logFileHnd = log.NewTextHandler(nil)
)

func init() {
	logFileHnd.SetLevel(slog.LevelDebug)
	log.AddHandler(logFileHnd)
}

func closeLog() {
	setLogFile(nil)
}

func setLogFile(w io.WriteCloser) {
	if f := logFile; f != nil {
		// Flush the old buffer before Reset changes its destination.
		_ = logBuf.Flush()
		_ = f.Close()
	}
	if w == nil {
		logFile = nil
		logFileHnd.SetWriter(nil)
		return
	}
	logFile = w
	logBuf.Reset(w)
	logFileHnd.SetWriter(&logBuf)
}

func defaultLogDir() string {
	logDir := filepath.Dir(os.Args[0])
	if sdir := env.AppUserDir(); sdir != "" {
		logDir = sdir
	}
	return filepath.Join(logDir, "logs")
}

func writeLogsToDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	name := "opennox.log"
	if isDedicatedServer {
		name = "opennox-server.log"
	}
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	setLogFile(f)
	return nil
}
