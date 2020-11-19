package main

import (
  "os"
  "strconv"
  "leader/directed_ring"
)

func main() {
  n, _ := strconv.Atoi(os.Args[len(os.Args) - 1])
  directed_ring.RunItaiRodeh(n)
}
