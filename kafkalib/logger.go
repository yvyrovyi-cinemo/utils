package kafkalib

import (
	"fmt"
	"log/slog"
)

type SaramaLogger struct {
	logger *slog.Logger
}

func NewSaramaLogger(logger *slog.Logger) *SaramaLogger {
	return &SaramaLogger{
		logger: logger,
	}
}

func (l *SaramaLogger) Printf(format string, v ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, v...))
}

func (l *SaramaLogger) Print(v ...interface{}) {
	l.logger.Debug(fmt.Sprint(v...))
}

func (l *SaramaLogger) Println(v ...interface{}) {
	l.logger.Debug(fmt.Sprint(v...))
}
