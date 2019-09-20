package app

import (
	"log"
	"os"
	"io"
)

var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func InitLog() {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Fatal to open error log file", err)
	}

	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Llongfile)
	Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Llongfile)
	Error = log.New(io.MultiWriter(os.Stderr, logFile), "ERROR: ", log.Ldate|log.Ltime|log.Llongfile)
}
