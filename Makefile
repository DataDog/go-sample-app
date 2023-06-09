.PHONY: clean users notes

notes:
	go build -o ./bin/notes ./services/notes

users:
	go build  -o ./bin/users ./services/users

clean:
	rm -f ./bin/* ./data/*
