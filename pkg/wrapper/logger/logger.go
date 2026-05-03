package logger

import "log"

type Field struct {
	Key   string
	Value interface{}
}

type Logger struct{}

func (l *Logger) With(fields ...Field) *Logger      { return l }
func (l *Logger) Info(msg string, fields ...Field)  {}
func (l *Logger) Warn(msg string, fields ...Field)  {}
func (l *Logger) Error(msg string, fields ...Field) {}
func (l *Logger) Warnf(format string, args ...interface{}) {
	log.Printf(format, args...)
}
func (l *Logger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}
func (l *Logger) Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func F(key string, value interface{}) Field { return Field{Key: key, Value: value} }
func New(service string) *Logger            { return (&Logger{}).With(F("service", service)) }
