package log

import (
	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

// NewLogger creates a new zap logger instance for use in the MCL command-line
// application.
func NewLogger(ws zapcore.WriteSyncer, debug bool) *zap.Logger {
	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}

	return zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.Lock(ws),
			zap.NewAtomicLevelAt(level),
		),
	)
}

// Must panics if there is a non-nil error, otherwise it returns a non-nil
// zap logger. This function is intended for use with the NewLogger function
// when the application must panic if a zap logger could not be created.
func Must(logger *zap.Logger, err error) *zap.Logger {
	if err != nil {
		panic(err)
	}
	return logger
}
