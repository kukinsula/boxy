package entity

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Log struct {
	// TODO: Date time.Time
	UUID    string                 `json:"uuid"`
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Meta    map[string]interface{} `json:"meta"`
}

func (log Log) Clean() {
	for k, v := range log.Meta {
		if v == nil {
			delete(log.Meta, k)
		}
	}
}

type Logger func(log Log) error

func NoOpLogger(log Log) error {
	return nil
}

func WriterLogger(writer io.Writer) Logger {
	return func(log Log) error {
		_, err := fmt.Fprintf(writer, "[%s] %s %v\n",
			log.UUID, log.Message, log.Meta)

		return err
	}
}

var StdoutLogger = WriterLogger(os.Stdout)
var StderrLogger = WriterLogger(os.Stderr)

func LogrusLogger(log Log) error {
	logrus.
		WithFields(log.Meta).
		Infof("[%s] %s", log.UUID, log.Message)

	return nil
}

func CleanMetaLogger(logger Logger) Logger {
	return func(log Log) error {
		log.Clean()

		return logger(log)
	}
}

func ChainLoggers(loggers ...Logger) Logger {
	return func(log Log) (err error) {
		for _, logger := range loggers {
			err = logger(log)
			if err != nil {
				return err
			}
		}

		return
	}
}
