package netcat

import (
	"log"
	"os"
	"time"
)

// time format is using underscore instead of space to seperate date and time
// double underscore is used to seperate datepart and timepart
func getTimeStampofServer() string {
	// we are using underscores instead of - and : and space because
	// some server have problems with processing spaces in file names
	time := time.Now().Format("2006_01_02__15_04_05")
	return time
}

// creating a new file for logging
func createLoggingFile() *os.File {
	fname := getTimeStampofServer() + ".log"
	file, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(err)
	}
	return file
}

// logger is a function that logs the messages to a file
func logger(file *os.File) {
	log.SetOutput(file)
}

// when the server starts = run the createNewLogger function
// createNewLogger creates a new file for logging
func CreateNewLogger() {
	file := createLoggingFile()
	logger(file)
}
