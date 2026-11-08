package sync_itai_rodeh_improved

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	. "github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/common"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type state struct {
	Phase  int
	W      string
	Status ModeType
}

func loadState(v lib.Node) state {
	var s state
	json.Unmarshal(v.GetState(), &s)
	return s
}

func saveState(v lib.Node, s state) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func elect(w string) ModeType {
	status := ModeType(Leader)
	for k := 1; k < len(w); k++ {
		r := w[k:] + w[:k]
		if r == w {
			return Unknown
		}
		if r < w {
			status = NonLeader
		}
	}
	return status
}

func process(v lib.Node, round int, q float64, rng *rand.Rand) bool {
	s := loadState(v)
	send := round > 0 && v.ReceiveMessage(0) != nil
	if n := v.GetSize(); round%n == 0 {
		if s.Phase > 0 {
			if s.Status = elect(s.W); s.Status != Unknown {
				log.Printf("Node %d: W = %s, phase %d -> %s", v.GetIndex(), s.W, s.Phase, s.Status)
				saveState(v, s)
				return true
			}
		}
		s.Phase, s.W = s.Phase+1, ""

		send = rng.Float64() < q/float64(n)
	}
	if send {
		s.W += "1"
		log.Println("Node", v.GetIndex(), "sends pebble to", v.GetOutNeighbors()[0].GetIndex())
		v.SendMessage(0, []byte{1})
	} else {
		s.W += "0"
		v.SendMessage(0, nil)
	}
	saveState(v, s)
	return false
}

func run(v lib.Node, q float64) {
	rng := rand.New(rand.NewSource(int64(v.GetIndex())))
	saveState(v, state{Status: Unknown})
	for round, finish := 0, false; !finish; round++ {
		v.StartProcessing()
		finish = process(v, round, q, rng)
		v.FinishProcessing(finish)
	}
}

func check(vertices []lib.Node) {
	leaders, phase := 0, loadState(vertices[0]).Phase
	for _, v := range vertices {
		s := loadState(v)
		if s.Status == Leader {
			leaders++
		} else if s.Status != NonLeader {
			panic(fmt.Sprintf("Node %d has state %s", v.GetIndex(), s.Status))
		}
		if s.Phase != phase {
			panic(fmt.Sprintf("Node %d finished in phase %d instead of %d", v.GetIndex(), s.Phase, phase))
		}
	}
	if leaders != 1 {
		panic(fmt.Sprintf("There are %d leaders on the directed ring", leaders))
	}
}

func Run(n int) (int, int) {
	return RunWithParameter(n, 1)
}

func RunWithParameter(n int, q float64) (int, int) {
	if q <= 0 || q >= float64(n) {
		panic(fmt.Sprintf("Parameter q = %g out of range (0, %d)", q, n))
	}
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)
	for _, v := range vertices {
		go run(v, q)
	}
	synchronizer.Synchronize(0)
	check(vertices)
	return synchronizer.GetStats()
}
