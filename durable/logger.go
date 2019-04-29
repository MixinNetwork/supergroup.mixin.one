package durable

import (
	"log"
)

type LoggerClient struct{}

type Logger struct{}

func NewLoggerClient() *LoggerClient {
	return &LoggerClient{}
}

func BuildLogger() *Logger {
	return &Logger{}
}

func (logger *Logger) Debug(v ...interface{}) {
	log.Println(v...)
}

func (logger *Logger) Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (logger *Logger) Info(v ...interface{}) {
	log.Println(v...)
}

func (logger *Logger) Infof(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (logger *Logger) Error(v ...interface{}) {
	log.Println(v...)
}

func (logger *Logger) Errorf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (logger *Logger) Panicln(v ...interface{}) {
	log.Panicln(v...)
}
