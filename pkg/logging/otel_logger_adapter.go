package logging

import "log/slog"

type LoggerAdapter struct {
	logger *slog.Logger
}

func NewLoggerAdapter(logger *slog.Logger) *LoggerAdapter {
	return &LoggerAdapter{logger: logger}
}

func (l *LoggerAdapter) Output(_ int, s string) error {
	l.logger.Info(s)

	return nil
}
