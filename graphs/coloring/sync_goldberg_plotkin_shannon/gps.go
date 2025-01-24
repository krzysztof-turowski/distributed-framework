package sync_goldberg_plotkin_shannon
import(
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"fmt"
	"encoding/json"
	"log"
	"math"
)


func Run(n int, p float64) (int, int) {
	nodes, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	delta := 0
	id := 0
	for _,node := range nodes {
		delta = max(delta,node.GetOutChannelsCount()) //each node must know max degree in graph
	}
	for _, node := range nodes {
		log.Println("Node", node.GetIndex(), "about to run")
		go run(node,delta,id)
		id++
	}
	synchronizer.Synchronize(3)
	check(nodes)
	return synchronizer.GetStats()
}

func check(nodes []lib.Node) {
	for _, node := range nodes {
		our_color := getState(node).Color
		if(our_color > getState(node).Graph_size) {
			panic("too many colors!")
		}
		for _, neighbor := range node.GetOutNeighbors() {
			if getState(neighbor).Color == our_color {
				panic("two adjacent nodes have the same color!")
			}
		}
		
	}
}



func run(node lib.Node,delta int,id int) {
	node.StartProcessing()
	
	finish := initialize(node,delta,id)
	node.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		node.StartProcessing()
		finish = process(node)
		
		node.FinishProcessing(finish)
		
	}
}
func initialize(node lib.Node, delta int,id int) bool {
	
	setState(node, makeState(node.GetOutChannelsCount(), delta,node.GetSize(),id))
	return false
}

func process(node lib.Node) bool {
	s := getState(node)
	switch s.Stage {
	case pseudoforest_creation:
		s.choose_edge_for_next_pseudoforest(node)
		setState(node,s)
	case coloring_pseudoforests_six_colors:
		s.reduce_colors_logarythmically(node)
		setState(node,s)
	case coloring_pseudoforests_with_three_colors:
		s.remove_additional_three_colors(node)
		setState(node,s)
	case recoloring_graph:
		s.final_recoloring(node)
		setState(node,s)
	case colored:
		return s.Done
	default:
		panic(fmt.Sprint("Invalid stage ", s.Stage))
	}
	return s.Done
}

//main algorithm 
func (s *state)final_recoloring(node lib.Node) {
	if(s.Done) {
			
		return
	}	
	s.Iteration++
	if(s.Iteration == 1) {
		if _,exists := s.Pseudoforests_in[s.Current_pseudoforest];exists {
			
			s.Edges_left_for_taking = append(s.Edges_left_for_taking, s.Pseudoforests_in[s.Current_pseudoforest]...)
				if(s.Pseudoforests_out[s.Current_pseudoforest] != -1) {
					s.Edges_left_for_taking = append(s.Edges_left_for_taking, s.Pseudoforests_out[s.Current_pseudoforest])
				}
		}
	}
		if(s.Iteration > s.Delta) {
			s.Iteration = 1
			s.Current_pseudoforest--
			if(s.Current_pseudoforest < 0) {
				s.Current_pseudoforest = s.Delta
				s.K++
				
			} else {
				
			}
			if(s.K == 3) {
				s.Done = true
				s.Stage = colored
			}
			
			return 
		}
	
		if _,exists := s.Pseudoforests_in[s.Current_pseudoforest];!exists {
			return
			
		}
		
		


		for _,neighbour := range s.Edges_left_for_taking{
			send_message(node,neighbour,message{
				Color: s.Color,
				Sender_id: s.Id,
				Color_in_pseudoforest: s.Color_in_pseudoforest[s.Current_pseudoforest],
			})
		}
		for _,neighbour := range s.Edges_left_for_taking{

			

			if(s.Color_in_pseudoforest[s.Current_pseudoforest] == s.K && s.Color == s.Iteration - 1) {
				msg := receive_next_message(node,neighbour)
				if(msg.Color_in_pseudoforest <= s.K) {
					s.free_colors[msg.Color] = false
				}
				
			} else {
				receive_next_message(node,neighbour)
			}
		 }

		if(s.Color_in_pseudoforest[s.Current_pseudoforest] == s.K && s.Color == s.Iteration - 1) {
			for i := s.Delta; i >= 0;i--{
				if(s.free_colors[i]) {
					s.Color = i;
					return
				}
			}
		}
}


func (s *state)remove_additional_three_colors(node lib.Node) {

	s.Iteration++
	
	if _,exists := s.Pseudoforests_in[s.Current_pseudoforest];!exists {
		if(s.coloring_3_previous_color == -1) {
			s.coloring_3_previous_color = 1
			
		} else {
			s.coloring_3_previous_color = -1
			s.Color_to_remove--
		}
		if(s.Color_to_remove == 2) {
			s.Color_to_remove = 5
			s.Current_pseudoforest++
		}
		if(s.Current_pseudoforest > s.Delta) {
			s.Current_pseudoforest--
			s.Stage = recoloring_graph
	
		} 
		
		return
	}

	if(s.coloring_3_previous_color == -1) {
		
		for _,child := range s.Pseudoforests_in[s.Current_pseudoforest] {
		
			send_message(node,child,message{Color: s.Color_in_pseudoforest[s.Current_pseudoforest],Sender_id: s.Id})
		} 
		var parent_color int
		if(s.Pseudoforests_out[s.Current_pseudoforest] == -1) {
			if(s.Color_in_pseudoforest[s.Current_pseudoforest] != 0) {
				parent_color = 0
			} else {
				parent_color = 1
			}
		} else {

			msg := receive_next_message(node,s.Pseudoforests_out[s.Current_pseudoforest])
			parent_color = msg.Color
			
		}
		s.coloring_3_previous_color = s.Color_in_pseudoforest[s.Current_pseudoforest]
		if(s.coloring_3_previous_color == -1) {
			panic("no color in pseudoforest -1")
		}
		s.Color_in_pseudoforest[s.Current_pseudoforest] = parent_color
		
	} else {
		for _,child := range s.Pseudoforests_in[s.Current_pseudoforest] {

			send_message(node,child,message{Color: s.Color_in_pseudoforest[s.Current_pseudoforest],Sender_id: s.Id})
		} 
		var parent_color int
		if(s.Pseudoforests_out[s.Current_pseudoforest] == -1) {
			if(s.Color_in_pseudoforest[s.Current_pseudoforest] != 0) {
				parent_color = 0
			} else {
				parent_color = 1
			}
		} else {
	
			parent_color = receive_next_message(node,s.Pseudoforests_out[s.Current_pseudoforest]).Color
		}
		if(s.Color_in_pseudoforest[s.Current_pseudoforest] == s.Color_to_remove) {
			if(s.coloring_3_previous_color != 0 && parent_color != 0) {
				s.Color_in_pseudoforest[s.Current_pseudoforest] = 0
			} else if(s.coloring_3_previous_color != 1 && parent_color != 1) {
				s.Color_in_pseudoforest[s.Current_pseudoforest] = 1
			} else {
				s.Color_in_pseudoforest[s.Current_pseudoforest] = 2
			}
		}
		s.coloring_3_previous_color = -1
		s.Color_to_remove--
		if(s.Color_to_remove == 2) {
			s.Color_to_remove = 5
			s.Current_pseudoforest++
			if(s.Current_pseudoforest > s.Delta) {
				s.Current_pseudoforest--
				s.Stage = recoloring_graph
			}

		}
	}
}
func (s *state)reduce_colors_logarythmically(node lib.Node) {

	if _,exists := s.Pseudoforests_in[s.Current_pseudoforest];!exists {
		s.Max_possible_color_used_in_coloring = compute_2_log_2_c(s.Max_possible_color_used_in_coloring)
		if(s.Max_possible_color_used_in_coloring <= 6) {

			s.Current_pseudoforest++
			s.Max_possible_color_used_in_coloring = s.Graph_size
		} 
		if(s.Current_pseudoforest > s.Delta) {
			s.Current_pseudoforest = 0
			s.Stage = coloring_pseudoforests_with_three_colors
			s.Color_to_remove = 5
		} 
		return
	}

	for _,child := range s.Pseudoforests_in[s.Current_pseudoforest] {

		send_message(node,child,message{Color: s.Color_in_pseudoforest[s.Current_pseudoforest],Sender_id: s.Id})
	} 
	
	var parent_color int
	if(s.Pseudoforests_out[s.Current_pseudoforest] == -1) {
		
		if(s.Color_in_pseudoforest[s.Current_pseudoforest] != 0) {
			parent_color = 0
		} else {
			parent_color = 1
		}
	} else {
		send_message(node,s.Pseudoforests_out[s.Current_pseudoforest],message{})

		msg :=  receive_next_message(node,s.Pseudoforests_out[s.Current_pseudoforest])
		parent_color =msg.Color
		
	}
	for _,child := range s.Pseudoforests_in[s.Current_pseudoforest] {
		receive_next_message(node,child)
	}
	different_bit_position := first_different_bit(parent_color,s.Color_in_pseudoforest[s.Current_pseudoforest])
	s.Color_in_pseudoforest[s.Current_pseudoforest] = (different_bit_position<<1) | ((s.Color_in_pseudoforest[s.Current_pseudoforest]>>different_bit_position) & 1)
	s.Max_possible_color_used_in_coloring = compute_2_log_2_c(s.Max_possible_color_used_in_coloring)
	if(s.Max_possible_color_used_in_coloring <= 6) {
		s.Current_pseudoforest++
		s.Max_possible_color_used_in_coloring = s.Graph_size
	} 
	
}



func (s *state) choose_edge_for_next_pseudoforest(node lib.Node ) {
	s.Iteration++
	if(len(s.Edges_left_for_taking) == 0) {
		
		if(s.Iteration > s.Delta) {
			s.Iteration = 0
			s.Current_pseudoforest = 0
			s.Stage = coloring_pseudoforests_six_colors
		}
		return
	}
	s.Number_of_pseudoforests++
	s.Pseudoforests_in[s.Current_pseudoforest] =make([]int, 0)
	s.Pseudoforests_out[s.Current_pseudoforest] = -1
	s.Color_in_pseudoforest[s.Current_pseudoforest] = s.Id
	edge_taken := s.Edges_left_for_taking[0]
	for _,edge := range s.Edges_left_for_taking {
		msg := message{
			Sender_id: s.Id,
			Edge_taken: edge_taken==edge,
		}
		
		send_message(node,edge,msg)
	}
	edges_left_in_graph_after_all_this := make([]int,0)
	for _,edge := range s.Edges_left_for_taking {
		msg := receive_next_message(node,edge)

		if(msg.Edge_taken) {
			if(edge == edge_taken) {
				if(msg.Sender_id > s.Id) {
					s.Pseudoforests_out[s.Current_pseudoforest] = edge_taken
				
				} else {
					s.Pseudoforests_in[s.Current_pseudoforest] = append(s.Pseudoforests_in[s.Current_pseudoforest], edge)
				}
			} else {
				s.Pseudoforests_in[s.Current_pseudoforest] = append(s.Pseudoforests_in[s.Current_pseudoforest], edge)
			}
		} else {
			if(edge == edge_taken) {
				s.Pseudoforests_out[s.Current_pseudoforest] = edge_taken

			} else {
				edges_left_in_graph_after_all_this = append(edges_left_in_graph_after_all_this, edge)
			}
		}

	}
	s.Edges_left_for_taking = edges_left_in_graph_after_all_this
	
	if(len(s.Edges_left_for_taking) == 0) {
		s.Current_pseudoforest = 0

	} else {
		s.Current_pseudoforest++
		
	}
	


	

}

//initializers
type message struct {
	Edge_taken bool
	Color int
	Leader_candidate int
	Color_in_pseudoforest int
	Sender_id int
}


type stage byte

const (
	pseudoforest_creation stage = iota
	colored
	coloring_pseudoforests_six_colors
	coloring_pseudoforests_with_three_colors
	recoloring_graph
)

type state struct {
	Stage stage
	Neighbors []int
	Delta int
	Graph_size int
	Color int
	Number_of_pseudoforests int
	Pseudoforests_in map[int][]int
	Pseudoforests_out map[int]int
	Edges_left_for_taking []int
	Done bool
	Max_possible_color_used_in_coloring int
	Current_pseudoforest int
	Color_in_pseudoforest []int
	Color_to_remove int
	coloring_3_previous_color int
	free_colors map[int]bool
	Id  int
	Iteration int
	selected_edge int
	K int
}
func setState(node lib.Node, s state) {
	representation, _ := json.Marshal(s)
	node.SetState(representation)
}

func getState(node lib.Node) state {
	var s state
	json.Unmarshal(node.GetState(), &s)
	return s
}
func makeState(degree int, delta int,size int,id int) state {
	s := state{
		
		Neighbors: make([]int,0),
		Pseudoforests_in: make(map[int][]int),
		Pseudoforests_out: make(map[int]int),
		Edges_left_for_taking: make([]int, 0),
		Color_in_pseudoforest: make([]int,delta + 1),
		free_colors: make(map[int]bool),
		Number_of_pseudoforests:  0,
		Done: false,
		Stage: pseudoforest_creation,
		Max_possible_color_used_in_coloring: size,
		Current_pseudoforest: 0,
		Color_to_remove: 5,
		coloring_3_previous_color: -1,
		Color: id,
		Delta: delta,
		Id : id,
		Iteration:0,
		Graph_size: size,
		K: 1,

	
	}
	if(degree == 0) {
		
		s.Done = true
		s.Stage = colored
	}
	for i := 0; i < degree; i++ {
		s.Neighbors = append(s.Neighbors, i)
		s.Edges_left_for_taking = append(s.Edges_left_for_taking, i)
		s.free_colors[i] = true
	}
	return s
}



//utilities






func first_different_bit(a, b int) int {
	
	xor := a ^ b
	
	
	for i := 0; xor != 0; i++ {
		if xor&1 == 1 {
			return i
		}
		xor >>= 1
	}
	

	return -1 
}










func send_message(node lib.Node,receiver int,msg message){
	//fmt.Println("Sending")
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Error serializing message: %v\n", err)
		return
	}
	node.SendMessage(receiver,data)
}

func receive_next_message(node lib.Node,sender int) message {
		data := node.ReceiveMessage(sender)
		var msg message
		json.Unmarshal(data, &msg);
		return msg
	
}


func compute_2_log_2_c( n int) int {
	log2 := math.Log2(float64(n))
	ceilLog2 := math.Ceil(log2)
	return 2 * int(ceilLog2)
}








