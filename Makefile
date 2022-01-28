all: test benchmark check

example: sync_directed_hypercube_leader_example sync_directed_ring_leader_example sync_undirected_graph_leader_example sync_undirected_mesh_leader_example sync_mst_example sync_mis_example sync_consensus_example async_undirected_ring_leader_example

runners:
	go run example/synchronized.go 5

async_undirected_ring_leader_example:
	go run example/undirected_ring_leader_test.go 10

sync_directed_hypercube_leader_example:
	go run example/sync_directed_hypercube_leader.go 6

sync_directed_ring_leader_example:
	go run example/sync_directed_ring_leader.go 10
	go run example/sync_directed_ring_chang_roberts.go 10
	go run example/sync_directed_ring_dolev_klawe_rodeh.go a 10
	go run example/sync_directed_ring_dolev_klawe_rodeh.go b 10
	go run example/sync_directed_ring_peterson.go 10

sync_undirected_ring_leader_example:
	go run example/sync_undirected_ring_hirschberg_sinclair.go 10

sync_undirected_graph_leader_example:
	go run example/sync_undirected_graph_yoyo.go 20 0.25

sync_undirected_mesh_leader_example:
	go run example/sync_undirected_mesh_peterson_leader.go 6 9

sync_mst_example:
	go run example/sync_mst_ghs.go 10 30 100

sync_mis_example:
	go run example/sync_mis_luby.go 20 0.25

sync_consensus_example:
	go run example/sync_consensus_ben_or.go 6 1
	go run example/sync_consensus_phase_king.go 10 3 0 1 0 1 0 1 1 0 0 0

unit_test:
	go test ./leader/undirected_graph -v
	go test ./graphs/mst -v

test:
	go test ./test -run . -v

benchmark:
	go test ./test -bench . -benchtime 10x -run Benchmark -v

check:
	@go vet `go list ./... | grep -v example`

format:
	gofmt -l -s -w .

.PHONY: all unit_test test benchmark
