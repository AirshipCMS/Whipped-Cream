
build:
	go build -o bin/Whipped-Cream .

build-linux64:
	env GOOS=linux GOARCH=amd64 go build -o bin/Whipped-Cream .

run:
	./bin/Whipped-Cream

dev: build run

