package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const usage = `Usage: nanny <file/dir> <commands>

Examples:

    nanny . "go build nanny.go; echo 'rinse, repeat'"
    nanny README.markdown "markdown README.markdown > temp.html; open temp.html"

`

const debug = false

func debugf(msg string, args ...interface{}) {
	if debug {
		fmt.Printf(msg, args...)
	}
}

// watcher watches a file/dir for changes
type watcher struct {
	path string
}

func (w *watcher) watch() {
	startTime, _ := w.newestMod()
	for {
		time.Sleep(time.Second)
		currentTime, file := w.newestMod()
		if currentTime.After(startTime) {
			debugf("%s: changed at %v\n", file, currentTime)
			return
		}
	}
}

func (w *watcher) newestMod() (modTime time.Time, file string) {
	filepath.Walk(w.path, func(path string, info os.FileInfo, err error) error {
		if info.ModTime().After(modTime) {
			modTime = info.ModTime()
			file = path
		}
		return nil
	})
	return
}

// runner runs commands
type runner struct {
	shell string
	cmds  string
}

func (r *runner) run() {
	cmd := exec.Command(r.shell)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = strings.NewReader(r.cmds)
	err := cmd.Run()
	debugf("%s: exited with error: %v\n", r.cmds, err)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf(usage)
		os.Exit(1)
	}

	w := &watcher{os.Args[1]}
	_, err := os.Stat(w.path)
	if err != nil {
		error := err.Error()
		if pe, ok := err.(*os.PathError); ok {
			error = pe.Err.Error()
		}
		fmt.Printf("%s: %s\n", w.path, error)
		os.Exit(1)
	}

	shell := os.Getenv("SHELL") // todo add ComSpec for windows (with /C flag to exit on completion)
	if shell == "" {
		fmt.Printf("Missing SHELL environment variable\n")
		os.Exit(1)
	}

	r := &runner{shell, os.Args[2]}
	for {
		w.watch()
		r.run()
	}
}
