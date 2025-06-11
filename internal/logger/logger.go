package logger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger
var HTTP *zap.SugaredLogger

var baseLogger *zap.Logger
var httpLogger *zap.Logger

type SplunkConfig struct {
	Url   string
	Token string
}

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

	appCores := []zapcore.Core{
		zapcore.NewCore(appEncoder, zapcore.AddSync(os.Stderr), zapcore.InfoLevel),
		zapcore.NewCore(appEncoder, zapcore.AddSync(appWriter), zapcore.InfoLevel),
	}

	// splunkConfig := SplunkConfig{Url: viper.GetString("splunk.url"), Token: viper.GetString("splunk.token")}

	// if splunkConfig.Url != "" && splunkConfig.Token != "" {
	// 	splunkWriter := &SplunkHECWriter{
	// 		URL:        splunkConfig.Url,
	// 		Token:      splunkConfig.Token,
	// 		Index:      "vitistack",
	// 		Source:     "ipam-api",
	// 		SourceType: "ipam-api:app",
	// 		Client:     &http.Client{Timeout: 5 * time.Second},
	// 	}
	// 	splunkCore := zapcore.NewCore(appEncoder, zapcore.AddSync(splunkWriter), zapcore.InfoLevel)
	// 	appCores = append(appCores, splunkCore)
	// }

	appCore := zapcore.NewTee(appCores...)

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

	httpCores := []zapcore.Core{
		zapcore.NewCore(httpEncoder, zapcore.AddSync(httpWriter), zapcore.InfoLevel),
	}

	// if splunkConfig.Url != "" && splunkConfig.Token != "" {
	// 	splunkWriter := &SplunkHECWriter{
	// 		URL:        splunkConfig.Url,
	// 		Token:      splunkConfig.Token,
	// 		Index:      "vitistack",
	// 		Source:     "ipam-api",
	// 		SourceType: "ipam-api:http",
	// 		Client:     &http.Client{Timeout: 5 * time.Second},
	// 	}
	// 	splunkCore := zapcore.NewCore(httpEncoder, zapcore.AddSync(splunkWriter), zapcore.InfoLevel)
	// 	httpCores = append(httpCores, splunkCore)
	// }

	httpCore := zapcore.NewTee(httpCores...)

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

// SplunkHECWriter implements zapcore.WriteSyncer
type SplunkHECWriter struct {
	URL        string
	Token      string
	Index      string
	Source     string
	SourceType string
	Client     *http.Client
}

func (w *SplunkHECWriter) Write(p []byte) (n int, err error) {
	payload := map[string]interface{}{
		"event":      json.RawMessage(p),
		"time":       time.Now().Unix(),
		"host":       "ipam-api",
		"source":     w.Source,
		"sourcetype": w.SourceType,
		"index":      w.Index,
	}

	client := resty.New().
		SetTimeout(5 * time.Second)

	resp, err := client.R().
		SetHeader("Authorization", "Splunk "+w.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(w.URL + "/services/collector/event")

	if err != nil {
		// Fail silently — we do not want to block the app if Splunk is down
		return len(p), nil
	}

	if resp.StatusCode() != 200 {
		// Optional: log if Splunk returns error (but do not propagate error)
		return len(p), nil
	}

	return len(p), nil
}

func (w *SplunkHECWriter) Sync() error {
	// Nothing to sync for HTTP writer
	return nil
}
