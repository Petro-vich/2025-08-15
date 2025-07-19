APP = myapp
SRC_DIR = cmd/main.go

all: clean run

build:
	@go build -o bin/$(APP) $(SRC_DIR)

run: build
	@./bin/$(APP)

test:
	go test ./internal/...

clean_all:
	@rm -f bin/$(APP)
	@rm -f tmp/*.pdf tmp/*.jpg tmp/*.jpeg archives/*.zip

clean:
	@rm -f bin/$(APP)

.PHONY: build run clean clean_all test

