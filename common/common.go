package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/luoyancn/dubhe/conf"
	"github.com/luoyancn/dubhe/logging"

	"github.com/spf13/viper"
)

type callback func()

func Wait(fn ...callback) {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case s := <-sig:
		logging.LOG.Infof("Terminating Daemon: Recevied signal %v\n", s)
		for _, f := range fn {
			f()
		}
		os.Exit(-1)
	}
}

func ReadConfig(conf_path string, logger string, logbak int,
	logpath string, debug bool, verbose bool, cfg ...conf.Config) {
	viper.SetConfigFile(conf_path)
	if err := viper.ReadInConfig(); nil != err {
		fmt.Printf(
			"Failed to read config file %s: %v\n", conf_path, err)
		fmt.Printf(
			"Using most configurations with default value instead.\n")
		fmt.Printf(
			"Note: Default valus maybe not correct in product env\n")
	}

	for _, c := range cfg {
		c.OverWriteConfig()
	}

	logging.GetLogger(logger, logbak, logpath, debug)

	if verbose {
		for key, value := range viper.AllSettings() {
			settings := value.(map[string]interface{})
			for setting_key, setting_value := range settings {
				logging.LOG.Noticef("%s.%s\t%v\n",
					key, setting_key, setting_value)
			}
		}
	}
}
