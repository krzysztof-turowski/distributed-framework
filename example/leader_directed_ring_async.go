package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_higham_przytycka"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_itah_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_itai_rodeh_2"
)

func main() {
	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 1 {
		fmt.Println("Invalid size", n)
		return
	}

	switch algorithm := strings.ToLower(os.Args[2]); algorithm {
	case "higham_przytycka":
		async_higham_przytycka.Run(n)
	case "itai_rodeh":
		async_itah_rodeh.Run(n)
	case "itai_rodeh_2":
		async_itai_rodeh_2.Run(n)
	}

}
