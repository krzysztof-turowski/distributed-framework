all: check

example:
	go run src/example/example_undirected.go 5

chang_roberts:
	go run src/leader/directed_ring/chang_roberts.go 5

check:
	@for DIR in ./src/*/ ; do echo "Directory: $$DIR"; golint $$DIR | grep -v "should have comment or be unexported" || true; done
