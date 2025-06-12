package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
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

	// --- Splunk config ---
	splunkEnabled := viper.GetBool("splunk.enable")
	splunkURL := viper.GetString("splunk.url")
	splunkToken := viper.GetString("splunk.token")
	splunkIndex := viper.GetString("splunk.index")

	// App specific
	splunkSource := viper.GetString("splunk.source")
	if splunkSource == "" {
		splunkSource = "vitistack:ipam-api"
	}

	// HTTP specific
	splunkSourcetypeApp := viper.GetString("splunk.sourcetype_app")
	if splunkSourcetypeApp == "" {
		splunkSourcetypeApp = "ipam-api:app"
	}

	// If Splunk configured, add Splunk core for App
	if splunkEnabled && splunkURL != "" && splunkToken != "" && splunkIndex != "" {
		splunkWriter := NewSplunkHECWriter(
			splunkURL,
			splunkToken,
			splunkIndex,
			splunkSource,
			splunkSourcetypeApp,
			5*time.Second,
		)

		splunkCore := zapcore.NewCore(appEncoder, zapcore.AddSync(splunkWriter), zapcore.InfoLevel)
		appCores = append(appCores, splunkCore)
	}

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

	// HTTP specific
	splunkSourcetypeHTTP := viper.GetString("splunk.sourcetype_http")
	if splunkSourcetypeHTTP == "" {
		splunkSourcetypeHTTP = "ipam-api:http"
	}

	// If Splunk configured, add Splunk core for HTTP
	if splunkEnabled && splunkURL != "" && splunkToken != "" && splunkIndex != "" {
		splunkWriter := NewSplunkHECWriter(
			splunkURL,
			splunkToken,
			splunkIndex,
			splunkSource,
			splunkSourcetypeHTTP,
			5*time.Second,
		)

		splunkCore := zapcore.NewCore(httpEncoder, zapcore.AddSync(splunkWriter), zapcore.InfoLevel)
		httpCores = append(httpCores, splunkCore)
	}

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

type SplunkHECWriter struct {
	URL        string
	Token      string
	Index      string
	Source     string
	SourceType string
	Client     *resty.Client
}

func NewSplunkHECWriter(url, token, index, source, sourcetype string, timeout time.Duration) *SplunkHECWriter {
	return &SplunkHECWriter{
		URL:        url,
		Token:      token,
		Index:      index,
		Source:     source,
		SourceType: sourcetype,
		Client: resty.New().
			SetTimeout(timeout),
	}
}

func (w *SplunkHECWriter) Write(p []byte) (n int, err error) {
	var logTimestamp int64
	var logEntry map[string]interface{}

	// Unmarshal original Zap event
	if err := json.Unmarshal(p, &logEntry); err != nil {
		logTimestamp = time.Now().Unix()
	} else {
		// Extract timestamp
		if tsStr, ok := logEntry["timestamp"].(string); ok {
			t, err := time.Parse(time.RFC3339, tsStr)
			if err == nil {
				logTimestamp = t.Unix()
			} else {
				logTimestamp = time.Now().Unix()
			}
		} else {
			logTimestamp = time.Now().Unix()
		}

		// Remove timestamp from event payload
		delete(logEntry, "timestamp")
	}

	// Marshal cleaned event payload back to []byte
	cleanedEvent, err := json.Marshal(logEntry)
	if err != nil {
		// fallback — use original p as event
		cleanedEvent = p
	}

	// Build Splunk payload
	payload := map[string]interface{}{
		"event":      json.RawMessage(cleanedEvent),
		"time":       logTimestamp,
		"host":       "ipam-api", // could also be dynamic
		"source":     w.Source,
		"sourcetype": w.SourceType,
		"index":      w.Index,
	}

	// Send to Splunk HEC using Resty
	resp, err := w.Client.R().
		SetHeader("Authorization", "Splunk "+w.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(w.URL + "/services/collector/event")

	if err != nil {
		// Fail silently — do not block app
		return len(p), nil
	}

	if resp.StatusCode() != 200 {
		// Optionally log here
		return len(p), nil
	}

	return len(p), nil
}
