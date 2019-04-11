package utilities

import (
	"github.com/vjeantet/jodaTime"
	"log"
	"os"
	"time"
)

var (
	Log *log.Logger
)

func NewLog(logpath string) {
	println("LogFile: " + logpath)

	if Exists(logpath) {
		//os.Remove("log_file")
		date := jodaTime.Format("YYYY_MM_dd_HH_mm_ss", time.Now())
		os.Rename(".\\log\\log.txt", ".\\log\\log-"+date+".txt")
	}

	file, err := os.Create(logpath)
	if err != nil {
		panic(err)
	}
	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
}
