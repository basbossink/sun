doc: (docs `git describe --tags --exact-match`)
docs version:
    asciidoctor -D build -r asciidoctor-diagram -a version={{version}} README.adoc
clean:
    rm -Rf build
release: clean
    ./release.sh
