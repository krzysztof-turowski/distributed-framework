package main

import (
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_chang_roberts"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_dolev_klawe_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_itai_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_peterson"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"text/tabwriter"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])

	log.SetOutput(ioutil.Discard)

	var results [5][2]int
	results[0][0], results[0][1] = sync_itai_rodeh.Run(n)
	results[1][0], results[1][1] = sync_chang_roberts.Run(n)
	results[2][0], results[2][1] = sync_dolev_klawe_rodeh.RunA(n)
	results[3][0], results[3][1] = sync_dolev_klawe_rodeh.RunB(n)
	results[4][0], results[4][1] = sync_peterson.Run(n)

	fmt.Println("Results")

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	fmt.Fprintln(w, "AlgoName\tMessages\tRounds\t")

	fmt.Fprintf(w, "ItaiRodeh\t%d\t%d\t\n", results[0][0], results[0][1])
	fmt.Fprintf(w, "ChangRoberts\t%d\t%d\t\n", results[1][0], results[1][1])
	fmt.Fprintf(w, "DolevKlaweRodeh_A\t%d\t%d\t\n", results[2][0], results[2][1])
	fmt.Fprintf(w, "DolevKlaweRodeh_B\t%d\t%d\t\n", results[3][0], results[3][1])
	fmt.Fprintf(w, "Peterson\t%d\t%d\t\n", results[4][0], results[4][1])
	w.Flush()
}
