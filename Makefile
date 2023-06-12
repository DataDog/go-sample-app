.PHONY: all clean data users notes

all: clean data users notes

notes:
	go build -o ./bin/notes ./services/notes

users:
	go build  -o ./bin/users ./services/users

data:
	mkdir -p data

clean:
	rm -f ./bin/* ./data/*
