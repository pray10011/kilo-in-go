clean:
	rm -f kilo

build:
	go build -o kilo

run:
	./kilo

all: clean build run