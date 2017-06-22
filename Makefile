all: build run

build:
	go build -o ./go-astivid

run:
	./go-astivid -v