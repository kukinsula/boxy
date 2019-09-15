package log

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Level string

const DEBUG = Level("DEBUG")
const INFO = Level("INFO")
const WARN = Level("WARN")
const ERROR = Level("ERROR")

type Log struct {
	Date    time.Time              `json:"date"`
	UUID    string                 `json:"uuid"`
	Level   Level                  `json:"level"`
	Message string                 `json:"message"`
	Meta    map[string]interface{} `json:"meta"`
}

type Logger func(uuid string,
	level Level,
	message string,
	meta map[string]interface{}) error

func NoOpLogger(uuid string,
	level Level,
	message string,
	meta map[string]interface{}) error {

	return nil
}

func WriterLogger(writer io.Writer) Logger {
	return func(uuid string,
		level Level,
		message string,
		meta map[string]interface{}) error {

		_, err := fmt.Fprintf(writer, "%s [%s] %s %s %v\n",
			time.Now().Format(time.RFC3339), uuid, level, message, meta)

		return err
	}
}

var StdoutLogger = WriterLogger(os.Stdout)
var StderrLogger = WriterLogger(os.Stderr)

func LogrusLogger(uuid string,
	level Level,
	message string,
	meta map[string]interface{}) error {

	logrus.
		WithFields(meta).
		Infof("%s [%s] %s", time.Now(), uuid, message)

	return nil
}

func CleanMetaLogger(logger Logger) Logger {
	return func(uuid string,
		level Level,
		message string,
		meta map[string]interface{}) error {

		for k, v := range meta {
			if v == nil {
				delete(meta, k)
			}
		}

		if len(meta) == 0 {
			meta = nil
		}

		return logger(uuid, level, message, meta)
	}
}

func ChainLoggers(loggers ...Logger) Logger {
	return func(uuid string,
		level Level,
		message string,
		meta map[string]interface{}) (err error) {

		for _, logger := range loggers {
			err = logger(uuid, level, message, meta)
			if err != nil {
				return err
			}
		}

		return
	}
}
