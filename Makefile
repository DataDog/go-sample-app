.PHONY: all clean data users notes licenses verify-licenses verify-dd-headers

all: clean data users notes verify-licenses verify-dd-headers

notes:
	go build -o ./bin/notes ./services/notes

users:
	go build  -o ./bin/users ./services/users

data:
	mkdir -p data

clean:
	rm -f ./bin/* ./data/*

licenses: bin/go-licenses
	tools/make-licenses.sh

verify-licenses: bin/go-licenses
	# TODO: verify-licenses is noop bc the copyright column in LICENSE-3rdparty.csv is manually generated right now.
	# This can be automated, see https://github.com/DataDog/orchestrion/tree/v0.1.0/tools/copyrights
	# tools/verify-licenses.sh

bin/go-licenses:
	mkdir -p $(PWD)/bin
	GOBIN=$(PWD)/bin go install github.com/google/go-licenses@v1.6.0

verify-dd-headers:
	go run tools/header_check.go
