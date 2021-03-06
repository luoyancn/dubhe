package logging

import (
	"os"
	"sync"

	gologging "github.com/op/go-logging"

	"github.com/luoyancn/dubhe/logging/config"
	"github.com/luoyancn/dubhe/logging/rotate"
)

// Please use this Variable after call the function GetLogger !!!
// Otherwise ,null pointer exception would be occured
var LOG *gologging.Logger

var format_std = gologging.MustStringFormatter(
	"%{color}%{time:2006-01-02 15:04:05.000000000}" +
		" %{level:-8.8s} %{shortfile} %{shortfunc}" +
		" %{color:reset} %{message}",
)

var format_file = gologging.MustStringFormatter(
	"%{time:2006-01-02 15:04:05.000000000} %{level:-8.8s} pid:%{pid} " +
		" [%{shortfile}] [%{shortfunc}] %{message}",
)

var init_once sync.Once

var once sync.Once

const STD_ENABLED = 1
const FILE_ENABLED = 2

func init() {
	init_once.Do(func() {
		LOG = gologging.MustGetLogger("")
		std_backend := gologging.NewLogBackend(os.Stdout, "", 0)
		std_back_formater := gologging.NewBackendFormatter(
			std_backend, format_std)
		gologging.SetBackend(std_back_formater)
	})
}

// Initail the global logger
// logger: The name of logger
// logback: The mode of logging
// logpath: The log file path
// debug: The log level. Default is INFO
func GetLogger(logger string, logback int, logpath string, debug bool) {
	once.Do(func() {
		LOG = gologging.MustGetLogger(logger)
		std_backend := gologging.NewLogBackend(os.Stdout, "", 0)
		std_back_formater := gologging.NewBackendFormatter(
			std_backend, format_std)
		switch logback {
		case 2:
			file_back_formater := get_file_logger(logpath, logger)
			gologging.SetBackend(file_back_formater)
			break
		case 3:
			file_back_formater := get_file_logger(logpath, logger)
			gologging.SetBackend(
				file_back_formater, std_back_formater)
			break
		default:
			gologging.SetBackend(std_back_formater)
			break
		}

		if debug {
			gologging.SetLevel(gologging.DEBUG, "")
		} else {
			gologging.SetLevel(gologging.INFO, "")
		}
	})
}

func get_file_logger(logpath string, logger string) gologging.Backend {
	/*
		logfile, err := os.OpenFile(path.Join(logpath, logger),
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			LOG.Panicf("Cannot create the log file :%v\n", err)
		}
	*/
	logfile, err := rotate.New(logpath, logger)
	if nil != err {
		LOG.Panicf("Cannot create the log file :%v\n", err)
	}
	logfile.SetKeep(config.LOG_KEEP)
	logfile.SetMax(config.LOG_MAXSIZE)

	file_backend := gologging.NewLogBackend(logfile, "", 0)
	file_back_formater := gologging.NewBackendFormatter(
		file_backend, format_file)
	return file_back_formater
}
