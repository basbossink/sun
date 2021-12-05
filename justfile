build:
	go build -o build/sun

clean:
    rm -Rf build

doc: (docs `git describe --tag` `git rev-parse HEAD`)

docs version hash:
    asciidoctor -D build -o index.html -r asciidoctor-diagram -a version={{version}} -a hash={{hash}} README.adoc

fumpt:
	gofumpt -w .
lint:
    golangci-lint run --enable-all

release: clean
    ./release.sh

rebuild: clean build

test:
	go test ./...

tcov:
    go test ./... -json -coverprofile cover.out |tparse -all 
    go tool cover -func=cover.out

