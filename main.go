package main

import (
	"os"
	"time"
)

var (
	Version    string
	CommitHash string
)

func run() int {
	now := time.Now()
	currentYear := now.Year()
	name := os.Args[0]
	env := newEnv(name)
	cmd := newCmdParser(name, os.Stderr)
	backend, err := newFSBackend(env)
	if err != nil {
		return 1
	}
	storage := newStorage(env, backend, currentYear)
	output := newOutput(os.Stdout)
	app := newApp(
		name,
		env,
		cmd,
		storage,
		output,
		Version,
		CommitHash,
		now,
		currentYear,
	)
	return app.run()
}

func main() {
	os.Exit(run())
}
