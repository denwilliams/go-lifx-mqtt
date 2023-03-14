package logging

import (
	"log"
	"os"
)

var (
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

func Init() {
	debugLogger = log.New(os.Stdout, "DEBG: ", log.Ldate|log.Ltime)
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	warningLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "ERRO: ", log.Ldate|log.Ltime)
}

func Debug(format string, v ...interface{}) {
	debugLogger.Printf(format+"\n", v...)
}

func Info(format string, v ...interface{}) {
	infoLogger.Printf(format+"\n", v...)
}

func Warn(format string, v ...interface{}) {
	warningLogger.Printf(format+"\n", v...)
}

func Error(format string, v ...interface{}) {
	errorLogger.Printf(format+"\n", v...)
}
