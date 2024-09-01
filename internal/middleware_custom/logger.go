package middleware_custom

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logMode string

type LogEntry struct {
	Logger logrus.FieldLogger
}

type LogFormatter struct{}

func SetupLogger(mode string) {
	logMode = mode

	logFile := &lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    50,
		MaxBackups: 10,
		Compress:   true,
	}

	logrus.SetOutput(logFile)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05",
		DisableQuote:    true,
	})
}

func (l *LogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	if logMode == "ALL" || (logMode == "ERR" && status >= 400) {
		elapsedTimeMs := float64(elapsed.Nanoseconds()) / 1e6
		logFields := logrus.Fields{
			"status":  status,
			"bytes":   bytes,
			"elapsed": fmt.Sprintf("%.2f ms", elapsedTimeMs),
		}

		if status >= 400 {
			l.Logger.WithFields(logFields).Log(logrus.ErrorLevel)
		} else {
			l.Logger.WithFields(logFields).Log(logrus.InfoLevel)
		}
	}
}

func (l *LogEntry) Panic(v interface{}, stack []byte) {
	l.Logger.WithFields(logrus.Fields{
		"panic": v,
		"stack": formatStackTrace(stack),
	}).Log(logrus.ErrorLevel)
}

func formatStackTrace(stack []byte) string {
	return string(stack)
}

func (l *LogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := logrus.WithFields(logrus.Fields{
		"method": r.Method,
		"uri":    r.RequestURI,
		"ip":     r.RemoteAddr,
	})
	return &LogEntry{Logger: entry}
}
