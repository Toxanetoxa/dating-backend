package logging

import (
	"errors"
	"log/slog"
	"os"
)

const (
	LogLvlError = "ERROR"
	LogLvlWarn  = "WARN"
	LogLvlInfo  = "INFO"
	LogLvlDebug = "DEBUG"

	DefaultLogLvl = slog.LevelError
)

var (
	ErrInvalidLogLvl = errors.New("invalid log level")
)

func InitLogger(lvl string) *slog.Logger {
	slogLvl, err := ParseLogLevel(lvl)
	if err != nil {
		slog.Error("can't parse log level, set default", "level", DefaultLogLvl)
	}

	opts := &slog.HandlerOptions{
		Level: slogLvl,
	}

	h := slog.NewJSONHandler(os.Stdout, opts)

	logger := slog.New(h)

	logger.Info("log initiated", "level", slogLvl)

	return logger
}

func ParseLogLevel(lvl string) (slog.Level, error) {
	switch lvl {
	case LogLvlDebug:
		return slog.LevelDebug, nil
	case LogLvlInfo:
		return slog.LevelInfo, nil
	case LogLvlWarn:
		return slog.LevelWarn, nil
	case LogLvlError:
		return slog.LevelError, nil
	default:
		return DefaultLogLvl, ErrInvalidLogLvl
	}
}
