all: test benchmark check

example: runners_example leader_directed_hypercube_example leader_directed_ring_example leader_undirected_ring_example leader_undirected_graph_example leader_undirected_mesh_example orientation_example graphs_mst_example graphs_mis_example graphs_ds_example consensus_example

runners_example:
	go run example/synchronized.go 5

leader_directed_hypercube_example:
	go run example/leader_directed_hypercube_sync_hyperelect.go 6

leader_directed_ring_example:
	go run example/leader_directed_ring_sync_all.go 10
	go run example/leader_directed_ring_sync_chang_roberts.go 10
	go run example/leader_directed_ring_sync_dolev_klawe_rodeh.go a 10
	go run example/leader_directed_ring_sync_dolev_klawe_rodeh.go b 10
	go run example/leader_directed_ring_sync_peterson.go 10
	go run example/leader_directed_ring_sync_itai_rodeh.go 10
	go run example/leader_directed_ring_async_itai_rodeh.go 10

leader_undirected_ring_example:
	go run example/leader_undirected_ring_sync_hirschberg_sinclair.go 10
	go run example/leader_undirected_ring_sync_franklin.go 10
	go run example/leader_undirected_ring_async_stages_with_feedback.go 10
	go run example/leader_undirected_ring_async_franklin.go 10
	go run example/leader_undirected_ring_async_probabilistic_franklin.go 10 3

leader_undirected_graph_example:
	go run example/leader_undirected_graph_sync_yoyo.go 20 0.25

leader_undirected_mesh_example:
	go run example/leader_undirected_mesh_sync_peterson.go 6 9

graphs_ds_example:
	go run example/graphs_ds_sync_lrg.go 10 0.70
	go run example/graphs_ds_sync_kuhn_wattenhofer.go 101 0.05 4

graphs_mst_example:
	go run example/graphs_mst_sync_ghs.go 10 30 100

graphs_mis_example:
	go run example/graphs_mis_sync_luby.go 20 0.25

consensus_example:
	go run example/consensus_sync_ben_or.go 11 2  0 1 0 1 0 1 1 0 0 0 1  1 2  Random
	go run example/consensus_sync_ben_or.go 11 2  0 1 0 1 0 1 1 0 0 0 1  1 2  Optimal
	go run example/consensus_sync_phase_king.go 10 3  0 1 0 1 0 1 1 0 0 0  1 2 3  Random
	go run example/consensus_sync_phase_king.go 10 3  0 1 0 1 0 1 1 0 0 0  1 2 3  Optimal
	go run example/consensus_sync_single_bit.go 9 2  0 1 0 1 0 1 1 0 0  1 2  Random
	go run example/consensus_sync_single_bit.go 9 2  0 1 0 1 0 1 1 0 0  1 2  Optimal

orientation_example:
	go run example/orientation_async_syrotiuk_pachl.go 10

unit_test:
	go test ./leader/undirected_graph/sync_yoyo -v
	go test ./graphs/mst/sync_ghs -v

test:
	go test ./test -run . -v

benchmark:
	go test ./test -bench . -benchtime 10x -run Benchmark -v -timeout 30m

check:
	@go vet `go list ./... | grep -v example`

format:
	gofmt -l -s -w .

.PHONY: all unit_test test benchmark
