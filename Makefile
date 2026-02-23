# Makefile para easy-git

BINARY=easy-git

.PHONY: build run tidy clean install

## build: compila o binário
build:
	go build -ldflags="-s -w" -o $(BINARY) .

## run: roda em modo dev
run:
	go run .

## tidy: baixa e organiza dependências
tidy:
	go mod tidy

## install: instala globalmente em /usr/local/bin
install: build
	sudo mv $(BINARY) /usr/local/bin/

## clean: remove binário gerado
clean:
	rm -f $(BINARY)

## cross: build para linux, mac e windows
cross:
	GOOS=linux   GOARCH=amd64 go build -o dist/$(BINARY)-linux .
	GOOS=darwin  GOARCH=amd64 go build -o dist/$(BINARY)-mac .
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY).exe .
