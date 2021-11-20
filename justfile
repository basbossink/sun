doc: (docs `git describe --tags --exact-match`)
docs version:
    asciidoctor -D build -o index.html -r asciidoctor-diagram -a version={{version}} README.adoc
clean:
    rm -Rf build
release: clean
    ./release.sh
