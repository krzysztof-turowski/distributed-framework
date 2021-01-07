all: check

synchronized_example:
	go run src/example/synchronized.go 5

directed_chang_roberts_example:
	go run src/example/directed_chang_roberts.go 5

directed_dolev_klawe_rodeh_example:
	go run src/example/directed_dolev_klawe_rodeh.go a 5
	go run src/example/directed_dolev_klawe_rodeh.go b 5

undirected_yoyo_example:
	go run src/example/undirected_yoyo_example_book.go
	go run src/example/undirected_yoyo_example_random.go 20 0.25

undirected_mesh_leader_example:
	go run src/example/undirected_mesh_leader.go 6 9

maximal_independent_set_luby_example:
	go run src/example/maximal_independent_set_luby.go 20 0.25

unit_test:
	go test leader/undirected_graph -v

test:
	go test test -run . -v

benchmark:
	go test test -bench . -benchtime 10x -run Benchmark -v

check:
	@for DIR in ./src/*/ ; do echo "Directory: $$DIR"; golint $$DIR | grep -v "should have comment or be unexported" || true; done

format:
	gofmt -l -s -w .
