package logging

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// NewLogrus makes a new Interface backed by a logrus logger
func NewLogrus(level Level) Interface {
	log := logrus.New()
	log.Out = os.Stderr
	log.Level = level.Logrus
	log.Formatter = &textFormatter{}
	return logrusLogger{log}
}

// Logrus wraps an existing Logrus logger.
func Logrus(l *logrus.Logger) Interface {
	return logrusLogger{l}
}

type logrusLogger struct {
	*logrus.Logger
}

func (l logrusLogger) WithField(key string, value interface{}) Interface {
	return logusEntry{
		Entry: l.Logger.WithField(key, value),
	}
}

func (l logrusLogger) WithFields(fields Fields) Interface {
	return logusEntry{
		Entry: l.Logger.WithFields(map[string]interface{}(fields)),
	}
}

type logusEntry struct {
	*logrus.Entry
}

func (l logusEntry) WithField(key string, value interface{}) Interface {
	return logusEntry{
		Entry: l.Entry.WithField(key, value),
	}
}

func (l logusEntry) WithFields(fields Fields) Interface {
	return logusEntry{
		Entry: l.Entry.WithFields(map[string]interface{}(fields)),
	}
}

type textFormatter struct{}

// Based off logrus.TextFormatter, which behaves completely
// differently when you don't want colored output
func (f *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	levelText := strings.ToUpper(entry.Level.String())[0:4]
	timeStamp := entry.Time.Format("2006/01/02 15:04:05.000000")
	fmt.Fprintf(b, "%s: %s %s", levelText, timeStamp, entry.Message)
	if len(entry.Data) > 0 {
		b.WriteString(" " + fieldsToString(entry.Data))
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}
