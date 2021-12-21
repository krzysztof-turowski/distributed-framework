package test

import (
	"flag"
	"io/ioutil"
	"log"
)

var isLogOn = flag.Bool("log", false, "Log output to screen")

func checkLogOutput() {
	if !*isLogOn {
		log.SetOutput(ioutil.Discard)
	}
}
