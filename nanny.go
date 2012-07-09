package main

import (
	"fmt"
	"io"
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

// watcher watches a file/dir for changes
type watcher struct {
	path string
}

func (w *watcher) watch() {
	startTime, _ := w.newestMod()
	for {
		time.Sleep(time.Second)
		currentTime, _ := w.newestMod()
		if currentTime.After(startTime) {
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
	cmd.Run() // todo check result!
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println(usage)
		os.Exit(1)
	}

	w := &watcher{os.Args[1]}
	_, err := os.Stat(w.path)
	if err != nil {
		error := err.Error()
		if pe, ok := err.(*os.PathError); ok {
			error = pe.Err.Error()
		}
		fmt.Println(w.path + ": " + error)
		os.Exit(1)
	}

	shell := os.Getenv("SHELL") // todo add ComSpec for windows (with /C flag to exit on completion)
	if shell == "" {
		fmt.Println("Missing SHELL environment variable")
		os.Exit(1)
	}

	r := &runner{shell, os.Args[2]}

	// quit if we recieve EOF on stdin, otherwise if started as a detached process
	// from another detached process, it will become a zombie
	go func() {
		bs := make([]byte, 1)
		for {
			_, err := os.Stdin.Read(bs)
			if err == io.EOF {
				os.Exit(0)
			}
		}
	}()

	for {
		w.watch()
		r.run()
	}
}
