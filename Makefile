all: users notes


notes:
	go build ./cmd/notes

users:
	go build ./cmd/users

clean:
	rm -f notes users data/*

