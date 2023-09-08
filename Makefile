GIT_COMMIT := $(shell git rev-list -1 HEAD) 


build:
	@go build -ldflags "-X main.GitCommit=$(GIT_COMMIT)"

run: build
	./service-monitorer

