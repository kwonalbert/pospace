all:	install

install:
	go install ./posgraph
	go install ./prover
	go install ./verifier

clean:
	go clean ./...

nuke:
	go clean -i ./...
