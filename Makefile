all: lurker

lurker: *.go lib/*.go
	go build -o lurker

test: *.go lib/*.go
	go test . ./lib
