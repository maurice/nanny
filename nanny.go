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

Example:

    nanny . "go build foo.go; echo 'rinse, repeat"

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

func (w *watcher) watch(changed chan<- bool) {
	// todo use inotify here rather than poll
	startTime, _ := w.newestMod()
	for {
		select {
		case <-time.Tick(time.Second):
			currentTime, file := w.newestMod()
			if currentTime.After(startTime) {
				debugf("%s: changed at %v\n", file, currentTime)
				changed <- true
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
	rawCmds []string
}

func newRunner(cmdsArg string) *runner {
	cmds := strings.Split(cmdsArg, ";")
	for i, rawCmd := range cmds {
		cmds[i] = strings.Trim(rawCmd, " ")
	}
	return &runner{cmds}
}

func (r *runner) run() {
	for _, rawCmd := range r.rawCmds {
		debugf("%s: running now\n", rawCmd)
		err := runCommand(rawCmd)
		if err != nil {
			debugf("%s: exited with error: %v\n", rawCmd, err)
		} else {
			debugf("%s: exited ok\n", rawCmd)
		}
	}
}

// starts the command and waits for it to exit synchronously
func runCommand(rawCmd string) error {
	args := strings.Split(rawCmd, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf(usage)
		os.Exit(1)
	}

	w := newWatcher(os.Args[1])
	r := newRunner(os.Args[2])

	changed := make(chan bool)
	for {
		go w.watch(changed)
		select {
		case <-changed:
			r.run()
		}
	}
}
