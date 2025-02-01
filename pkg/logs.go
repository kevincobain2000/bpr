package pkg

import (
	"log/slog"
	"os"

	"github.com/MatusOllah/slogcolor"
	"github.com/mattn/go-isatty"
)

const (
	LogLevelInfo  = 0
	LogLevelDebug = -4
	LogLevelWarn  = 4
	LogLevelError = 8
)

func SetupLoggingStdout(f Flags) error {
	opts := &slogcolor.Options{
		Level:   slog.Level(f.LogLevel),
		NoColor: !isatty.IsTerminal(os.Stderr.Fd()),
	}
	if f.LogLevel == LogLevelDebug {
		opts = &slogcolor.Options{
			Level:       slog.Level(f.LogLevel),
			TimeFormat:  "2006-01-02 15:04:05",
			NoColor:     !isatty.IsTerminal(os.Stderr.Fd()),
			SrcFileMode: slogcolor.ShortFile,
		}
	}

	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stdout, opts)))
	return nil
}
