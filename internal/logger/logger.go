package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger
var HTTP *zap.SugaredLogger

var baseLogger *zap.Logger
var httpLogger *zap.Logger

func InitLogger(logDir string) error {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return fmt.Errorf("could not create log directory: %w", err)
	}

	// --- App logger ---
	appFile := filepath.Join(logDir, "ipam-api.log")
	appWriter := getWriter(appFile)

	appEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})

	appCore := zapcore.NewTee(
		//zapcore.NewCore(appEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		zapcore.NewCore(appEncoder, zapcore.AddSync(appWriter), zapcore.InfoLevel),
	)

	baseLogger = zap.New(appCore, zap.AddCaller())
	Log = baseLogger.Sugar()

	// --- HTTP logger ---
	httpFile := filepath.Join(logDir, "http.log")
	httpWriter := getWriter(httpFile)

	httpEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})

	httpCore := zapcore.NewTee(
		// zapcore.NewCore(httpEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		zapcore.NewCore(httpEncoder, zapcore.AddSync(httpWriter), zapcore.InfoLevel),
	)

	httpLogger = zap.New(httpCore, zap.AddCaller())
	HTTP = httpLogger.Named("http").Sugar()

	return nil
}

func Sync() {
	if baseLogger != nil {
		_ = baseLogger.Sync()
	}
	if httpLogger != nil {
		_ = httpLogger.Sync()
	}
}

func getWriter(path string) *os.File {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("unable to open log file: " + err.Error())
	}
	return f
}
