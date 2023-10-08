GIT_COMMIT := $(shell git rev-list -1 HEAD) 
MAKEFLAGS += -j2 # run 2 targets in parallel. For the 'backendMocks' command.

build:
	@go build -ldflags "-X main.GitCommit=$(GIT_COMMIT)"

run: build
	./service-monitorer -config config.example.yaml

debug: build
	./service-monitorer -loglevel debug -config config.example.yaml

backendMock: 
	go run ./cmd/service_mock/main.go -failedComponents db,sfirm,tripica,pdfgenerator,jenkins -address localhost:1234

backendMock2: 
	go run ./cmd/service_mock/main.go -failedComponents payment,kubernetes -address localhost:1235

backendMocks:  backendMock backendMock2