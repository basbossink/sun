version := `git describe --tags --exact-match`
hash := `git rev-parse HEAD`
build:
    go build -ldflags="-X main.Version={{version}} -X main.CommitHash={{hash}}"
docs:
    asciidoctor -a version={{version}} README.adoc