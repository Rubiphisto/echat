package logger

import (
	"log"
)

func init() {
	logger = &defaultLogger{}
}

type defaultLogger struct {
}

func (*defaultLogger) Debug(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format+"\n", v...)
}

func (*defaultLogger) Info(format string, v ...interface{}) {
	log.Printf("[INFO] "+format+"\n", v...)
}

func (*defaultLogger) Warn(format string, v ...interface{}) {
	log.Printf("[WARN] "+format+"\n", v...)
}

func (*defaultLogger) Error(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format+"\n", v...)
}

func (*defaultLogger) Panic(format string, v ...interface{}) {
	log.Printf("[PANIC] "+format+"\n", v...)
	panic("")
}

