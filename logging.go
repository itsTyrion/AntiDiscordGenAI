package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

func ShortNameEncoder(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
	segments := strings.Split(loggerName, ".")
	enc.AppendString(segments[len(segments)-1])
}

func SetupLogging() *zap.SugaredLogger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeTime:    zapcore.TimeEncoderOfLayout("15:04:05"), // Time format
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,        // Colored levels
		EncodeName:    ShortNameEncoder,                        // Short logger name
		EncodeCaller:  zapcore.ShortCallerEncoder,              // Short caller
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig), // Console encoder
		zapcore.AddSync(zapcore.Lock(os.Stdout)), // Output to stdout
		zapcore.DebugLevel,                       // Log level
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	defer logger.Sync()

	return logger.Sugar()
}
