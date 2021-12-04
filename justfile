doc: (docs `git describe --tag` `git rev-parse HEAD`)
docs version hash:
    asciidoctor -D build -o index.html -r asciidoctor-diagram -a version={{version}} -a hash={{hash}} README.adoc
clean:
    rm -Rf build
tcov:
    go test ./... -json -coverprofile cover.out |tparse -all 
    go tool cover -func=cover.out
release: clean
    ./release.sh
