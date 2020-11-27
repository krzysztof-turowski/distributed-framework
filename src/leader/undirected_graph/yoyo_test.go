package undirected_graph

import (
	"flag"
	"io/ioutil"
	"log"
	"testing"
)

var isLogOn = flag.Bool("log", false, "Log output to screen")

func checkLogOutput() {
	if !*isLogOn {
		log.SetOutput(ioutil.Discard)
	}
}

func TestYoYo(t *testing.T) {
	checkLogOutput()
	RunYoYo(1000)
}
