all: check

example:
	go run src/example_simulator.go src/lib_*.go src/example_client.go 5

check:
	golint src/*.go
