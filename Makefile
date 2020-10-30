all: check

example:
	go run src/example/example_simulator.go src/example/example_client.go 5

check:
	@for DIR in ./src/*/ ; do echo "Directory: $$DIR"; golint $$DIR | grep -v "should have comment or be unexported" || true; done
	
