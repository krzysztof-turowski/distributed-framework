package test

import (
	"fmt"
	"testing"

	wildfire "github.com/krzysztof-turowski/distributed-framework/aggregation/wild_fire"
)

func TestWildFireAggregationQuery(t *testing.T) {
	checkLogOutput()
	a, b := wildfire.Run(5000, 0.01, 0.2, "max", 50, 2)
	fmt.Println(a, b)
}
