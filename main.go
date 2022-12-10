package main

import (
	"log"
	"os"
)

func main() {
	conf, errs := loadConf()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Print(err.Error())
		}
		os.Exit(1)
	}
	_ = conf
}
