doc: (docs `git describe --tag` `git rev-parse HEAD`)
docs version hash:
    asciidoctor -D build -o index.html -r asciidoctor-diagram -a version={{version}} -a hash={{hash}} README.adoc
clean:
    rm -Rf build
release: clean
    ./release.sh
