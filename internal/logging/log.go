package logging

import "log"

func Init() {
	log.SetFlags(0)
}

func Debug(format string, v ...interface{}) {
	log.Printf("[DEBG] "+format+"\n", v...)
}

func Info(format string, v ...interface{}) {
	log.Printf("[INFO] "+format+"\n", v...)
}

func Warn(format string, v ...interface{}) {
	log.Printf("[WARN] "+format+"\n", v...)
}

func Error(format string, v ...interface{}) {
	log.Printf("[ERRO] "+format+"\n", v...)
}
