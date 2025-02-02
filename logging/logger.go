package logging

import (
	"log"
	"os"
)

var logger *log.Logger
var filename = "${HOME}/.jerminal/logs/logs.txt"

func Logger() *log.Logger {
    if logger == nil {
        logfile, err := os.OpenFile(os.ExpandEnv(filename), os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0666)
        if err != nil {
            panic(err)
        }
        logger = log.New(logfile, "[jerminal]", log.Ldate | log.Ltime | log.Lshortfile)
        return logger
    }
    return logger
}

