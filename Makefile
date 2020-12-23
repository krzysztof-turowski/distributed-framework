all: check

synchronized_example:
	go run src/example/synchronized_ring_example.go 5

directed_chang_roberts_example:
	go run src/example/directed_chang_roberts.go 5

test:
	go test test -run . -v

benchmark:
	go test test -bench . -benchtime 10x -run Benchmark -v

check:
	@for DIR in ./src/*/ ; do echo "Directory: $$DIR"; golint $$DIR | grep -v "should have comment or be unexported" || true; done

format:
	gofmt -l -s -w .
