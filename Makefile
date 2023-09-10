GIT_COMMIT := $(shell git rev-list -1 HEAD) 


build:
	@go build -ldflags "-X main.GitCommit=$(GIT_COMMIT)"

run: build
	./service-monitorer

backendMock: 
	go run ./cmd/service/main.go -failedComponents db,sfirm -okComponents tripica,mako -address localhost:1234


