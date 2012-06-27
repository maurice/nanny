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

func newWatcher(path string) *watcher {
	_, err := os.Stat(path)
	if pe, ok := err.(*os.PathError); ok {
		fmt.Printf("%s: %s\n", pe.Path, pe.Err)
		os.Exit(1)
	}
	return &watcher{path}
}

func (w *watcher) watch() {
	// todo use inotify here rather than poll
	startTime, _ := w.newestMod()
	for {
		select {
		case <-time.Tick(time.Second):
			currentTime, file := w.newestMod()
			if currentTime.After(startTime) {
				debugf("%s: changed at %v\n", file, currentTime)
				return
			}
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
	cmds string
}

func newRunner(cmds string) *runner {
	return &runner{cmds}
}

func (r *runner) run() {
	cmd := exec.Command(os.Getenv("SHELL"))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = strings.NewReader(r.cmds)
	err := cmd.Run()
	debugf("%s: exited with error: %v\n", cmd, err)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf(usage)
		os.Exit(1)
	}

	w := newWatcher(os.Args[1])
	r := newRunner(os.Args[2])

	for {
		w.watch()
		r.run()
	}
}
