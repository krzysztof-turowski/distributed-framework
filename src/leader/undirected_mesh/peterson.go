package undirected_mesh

import (
	"encoding/json"
	"lib"
	"log"
)

type modeType string

const (
	passive    modeType = "passive"
	forwarding          = "forwarding"
	leader              = "leader"
	candidate           = "candidate"
)

type stateMesh struct {
	Min                   int
	Status                modeType
	Neighbours            []modeType
	Candidates            []int
	Forwarders            []int
	Passives              []int
	PassRight             *messageMesh
	PassLeft              *messageMesh
	Broadcast             *messageMesh
	NeighboursToBroadcast []int
	Leader                int
}

type messageMesh struct {
	Id   int
	Mode modeType
}

func processMesh(v lib.Node, round int, rounds int) bool {
	s := getStateMesh(v)
	//broadcast phase
	if round > rounds {
		for _, i := range s.NeighboursToBroadcast {
			sendMessageMesh(v, i, s.Broadcast)
		}
    done := (s.Broadcast != nil)
		NewNeighboursToBroadcast := make([]int, 0)
		for _, i := range s.NeighboursToBroadcast {
			msg := receiveMessageMesh(v, i)
			if msg != nil {
				s.Broadcast = msg
				if msg.Mode == leader {
					s.Leader, s.Status = msg.Id, passive
				}
			} else {
				NewNeighboursToBroadcast = append(NewNeighboursToBroadcast, i)
			}
		}
		s.NeighboursToBroadcast = NewNeighboursToBroadcast
		setStateMesh(v, &s)
		return done
	}
	if s.Status == passive {
		return false
	}
	if s.Status == candidate {
		sendMessageMesh(v, 0, s.PassLeft)
		sendMessageMesh(v, 1, s.PassRight)
		msg1 := receiveMessageMesh(v, 0)
		if msg1 != nil {
			if msg1.Id <= s.Min {
				s.Min, s.PassRight = msg1.Id, msg1
			} else {
				s.PassRight = nil
			}
		} else {
			s.PassRight = msg1
		}
		msg2 := receiveMessageMesh(v, 1)
		s.PassLeft = msg2
		if msg2 != nil {
			if msg2.Id <= s.Min {
				s.Min, s.PassLeft = msg2.Id, msg2
			} else {
				s.PassLeft = nil
			}
		} else {
			s.PassLeft = msg2
		}
	}
	//left - Candidates[0] or Forwarders[0]
	//right - Candidates[1] or Forwarders[0] or Forwarders[1]
	if s.Status == forwarding {
		switch len(s.Candidates) {
		case 2:
			sendMessageMesh(v, s.Candidates[0], s.PassLeft)
			sendMessageMesh(v, s.Candidates[1], s.PassRight)
			s.PassRight = receiveMessageMesh(v, s.Candidates[0])
			s.PassLeft = receiveMessageMesh(v, s.Candidates[1])
		case 1:
			if len(s.Forwarders) == 1 {
				sendMessageMesh(v, s.Candidates[0], s.PassLeft)
				sendMessageMesh(v, s.Forwarders[0], s.PassRight)
				s.PassRight = receiveMessageMesh(v, s.Candidates[0])
				s.PassLeft = receiveMessageMesh(v, s.Forwarders[0])
			}
		case 0:
			sendMessageMesh(v, s.Forwarders[0], s.PassLeft)
			sendMessageMesh(v, s.Forwarders[1], s.PassRight)
			s.PassRight = receiveMessageMesh(v, s.Forwarders[0])
			s.PassLeft = receiveMessageMesh(v, s.Forwarders[1])
		}
	}
	if round == rounds && s.Min == v.GetIndex() && s.Status == candidate {
		s.Status, s.Leader = leader, v.GetIndex()
		s.Broadcast = &messageMesh{s.Min, leader}
	}
	setStateMesh(v, &s)
	return false
}

func newMeshState(Id int, State modeType, Neighbours []modeType, Candidates []int, Forwarders []int, Passives []int, PassRight *messageMesh, PassLeft *messageMesh, NeighbourCount []int) *stateMesh {
	return &stateMesh{
		Min:                   Id,
		Status:                State,
		Neighbours:            Neighbours,
		Candidates:            Candidates,
		Forwarders:            Forwarders,
		Passives:              Passives,
		PassRight:             PassRight,
		PassLeft:              PassLeft,
		Broadcast:             nil,
		NeighboursToBroadcast: NeighbourCount,
		Leader:                -1,
	}
}

func initializeMesh(v lib.Node) bool {
	var state modeType
	var PassRight *messageMesh
	var PassLeft *messageMesh
	switch v.GetOutChannelsCount() {
	case 2:
		state = candidate
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			msg := messageMesh{v.GetIndex(), candidate}
			sendMessageMesh(v, i, &msg)
		}
		PassRight = &messageMesh{v.GetIndex(), candidate}
		PassLeft = &messageMesh{v.GetIndex(), candidate}
	case 3:
		state = forwarding
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			msg := messageMesh{v.GetIndex(), forwarding}
			sendMessageMesh(v, i, &msg)
		}
		PassRight = nil
		PassLeft = nil
	case 4:
		state = passive
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			msg := messageMesh{v.GetIndex(), passive}
			sendMessageMesh(v, i, &msg)
		}
		PassRight = nil
		PassLeft = nil
	}
	Neighbours, Candidates, Forwarders, Passives := buildNeighboursMesh(v)
	NeighboursToBroadcast := make([]int, 0)
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		NeighboursToBroadcast = append(NeighboursToBroadcast, i)
	}
	setStateMesh(v, newMeshState(v.GetIndex(), state, Neighbours, Candidates, Forwarders, Passives, PassRight, PassLeft, NeighboursToBroadcast))
	return false
}

func buildNeighboursMesh(v lib.Node) ([]modeType, []int, []int, []int) {
	Neighbours := make([]modeType, 0)
	Candidates := make([]int, 0)
	Forwarders := make([]int, 0)
	Passives := make([]int, 0)
	for i := 0; i < v.GetInChannelsCount(); i++ {
		msg := receiveMessageMesh(v, i)
		Neighbours = append(Neighbours, msg.Mode)
		switch msg.Mode {
		case candidate:
			Candidates = append(Candidates, i)
		case forwarding:
			Forwarders = append(Forwarders, i)
		case passive:
			Passives = append(Passives, i)
		}
	}
	return Neighbours, Candidates, Forwarders, Passives
}

func receiveMessageMesh(v lib.Node, index int) *messageMesh {
	var msg messageMesh
	in := v.ReceiveMessage(index)
	if in != nil {
		json.Unmarshal(in, &msg)
		return &msg
	}
	return nil
}

func setStateMesh(v lib.Node, state *stateMesh) {
	data, _ := json.Marshal(*state)
	v.SetState(data)
}

func getStateMesh(v lib.Node) stateMesh {
	var s stateMesh
	in := v.GetState()
	json.Unmarshal(in, &s)
	return s
}

func sendMessageMesh(v lib.Node, index int, msg *messageMesh) {
	if msg != nil {
		data, _ := json.Marshal(msg)
		v.SendMessage(index, data)
	} else {
		v.SendMessage(index, nil)
	}
}

func runPeterson(v lib.Node, rounds int) {
	v.StartProcessing()
	finish := initializeMesh(v)
	v.FinishProcessing(finish)
	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processMesh(v, round, rounds)
		v.FinishProcessing(finish)
	}
}

func checkPeterson(vertices []lib.Node) {
	var lead_node lib.Node
	for _, v := range vertices {
		s := getStateMesh(v)
		if s.Status == leader && lead_node != nil {
			panic("multiple leaders")
		}
		if s.Status == leader {
			lead_node = v
		}
	}
	for _, v := range vertices {
		s := getStateMesh(v)
		if s.Leader == -1 {
			panic("node has no leader")
		}
		if s.Leader != lead_node.GetIndex() {
			panic("node has a wrong leader")
		}
	}
	log.Println("all good, leader id is: ", lead_node.GetIndex())
}

func RunPeterson(vertices []lib.Node, synchronizer lib.Synchronizer, rounds int) {
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runPeterson(v, rounds)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkPeterson(vertices)
}
