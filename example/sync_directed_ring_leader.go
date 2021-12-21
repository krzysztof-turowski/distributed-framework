package main

import (
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"text/tabwriter"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])

	log.SetOutput(ioutil.Discard)

	var results [4][2]int
	results[0][0], results[0][1] = directed_ring.RunItaiRodeh(n)
	results[1][0], results[1][1] = directed_ring.RunChangRoberts(n)
	results[2][0], results[2][1] = directed_ring.RunDolevKlaweRodehA(n)
	results[3][0], results[3][1] = directed_ring.RunDolevKlaweRodehB(n)

	fmt.Println("Results")

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	fmt.Fprintln(w, "AlgoName\tMessages\tRounds\t")

	fmt.Fprintf(w, "ItaiRodeh\t%d\t%d\t\n", results[0][0], results[0][1])
	fmt.Fprintf(w, "ChangRoberts\t%d\t%d\t\n", results[1][0], results[1][1])
	fmt.Fprintf(w, "DolevKlaweRodeh_A\t%d\t%d\t\n", results[2][0], results[2][1])
	fmt.Fprintf(w, "DolevKlaweRodeh_B\t%d\t%d\t\n", results[3][0], results[3][1])
	w.Flush()
}
