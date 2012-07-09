package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"
)

func fileExists(name string) bool {
	fi, _ := os.Stat(name)
	return fi != nil
}

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}

func createTempFile(t *testing.T, name string) (*os.File, func()) {
	tmp := path.Dir(os.TempDir())
	fullName := filepath.Join(tmp, name)
	dir := path.Dir(fullName)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		t.Fatalf("Failed to mkdir `%s`: %s\n", dir, err)
	}
	file, err := os.Create(fullName)
	if err != nil {
		t.Fatalf("Failed to create file `%s`: %s\n", file, err)
	}
	return file, func() {
		for {
			os.Remove(fullName)
			fullName = path.Dir(fullName)
			if fullName == tmp {
				break
			}
		}
	}
}

func runNanny(t *testing.T, target, command string) (kill chan bool) {
	kill = make(chan bool)
	cmd := exec.Command("go", "run", "nanny.go", target, command)
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Couldn't get stdin: %s", err)
	}
	go func() {
		<-kill
		in.Close()
	}()
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Couldn't run nanny: %s", err)
	}
	return
}

func TestCommandsNotRunIfWatchedFileNotModified(t *testing.T) {
	outName := tempFile("out_file")
	os.Remove(outName)

	file, cleanup := createTempFile(t, "watched_file")
	defer cleanup()

	kill := runNanny(t, file.Name(), "echo modified >> "+outName)

	time.Sleep(time.Second) // give time to run
	time.Sleep(time.Second) // give time to run

	if fileExists(outName) {
		t.Errorf("nanny ran command despite no change to target: %s", outName)
	}
	kill <- true
}

func TestCommandsRunOnceWhenWatchedFileModifiedOnce(t *testing.T) {
	outName := tempFile("out_file")
	os.Remove(outName)

	file, cleanup := createTempFile(t, "watched_file")
	defer cleanup()

	kill := runNanny(t, file.Name(), "echo modified >> "+outName)

	time.Sleep(time.Second)  // give time to run
	file.WriteString("mod1") // modify watched file
	time.Sleep(time.Second)  // give time to run

	if !fileExists(outName) {
		t.Errorf("nanny should have run and created target: %s", outName)
	}
	bs, err := ioutil.ReadFile(outName)
	if err != nil {
		t.Errorf("error reading target file: %s", err)
	}
	expect := []byte(`modified
`)
	if !bytes.Equal(expect, bs) {
		t.Errorf("target file contents wrong, got `%s`, want `%s`", bs, expect)
	}

	kill <- true
}

func TestCommandRunOnceAfterEachChangeToWatchedFile(t *testing.T) {
	outName := tempFile("out_file")
	os.Remove(outName)

	file, cleanup := createTempFile(t, "watched_file")
	defer cleanup()

	kill := runNanny(t, file.Name(), "echo modified >> "+outName)

	time.Sleep(time.Second)  // give time to run
	file.WriteString("mod1") // modify watched file
	time.Sleep(time.Second)  // give time to run

	bs, _ := ioutil.ReadFile(outName)
	expect := []byte(`modified
`)
	if !bytes.Equal(expect, bs) {
		t.Errorf("target file contents wrong, got `%s`, want `%s`", bs, expect)
	}

	//	time.Sleep(time.Second)  // give time to run
	file.WriteString("mod2") // modify watched file
	time.Sleep(time.Second)  // give time to run

	bs, _ = ioutil.ReadFile(outName)
	expect = []byte(`modified
modified
`)
	if !bytes.Equal(expect, bs) {
		t.Errorf("target file contents wrong, got `%s`, want `%s`", bs, expect)
	}

	//	time.Sleep(time.Second)  // give time to run
	file.WriteString("mod3") // modify watched file
	time.Sleep(time.Second)  // give time to run

	bs, _ = ioutil.ReadFile(outName)
	expect = []byte(`modified
modified
modified
`)
	if !bytes.Equal(expect, bs) {
		t.Errorf("target file contents wrong, got `%s`, want `%s`", bs, expect)
	}

	kill <- true
}

func TestWatchedFileMayBeDir(t *testing.T) {
	outName := tempFile("out_file")
	os.Remove(outName)

	file, cleanup := createTempFile(t, "watched_file")
	defer cleanup()

	kill := runNanny(t, path.Dir(file.Name()), "echo modified >> "+outName)

	time.Sleep(time.Second)  // give time to run
	file.WriteString("mod1") // modify watched file
	time.Sleep(time.Second)  // give time to run

	if !fileExists(outName) {
		t.Errorf("nanny should have run and created target: %s", outName)
	}
	bs, err := ioutil.ReadFile(outName)
	if err != nil {
		t.Errorf("error reading target file: %s", err)
	}
	expect := []byte(`modified
`)
	if !bytes.Equal(expect, bs) {
		t.Errorf("target file contents wrong, got `%s`, want `%s`", bs, expect)
	}

	kill <- true
}

func TestDirIsWatchedRecursively(t *testing.T) {
	outName := tempFile("out_file")
	os.Remove(outName)

	file, cleanup := createTempFile(t, "foo/bar/watched_file")
	defer cleanup()

	kill := runNanny(t, path.Dir(file.Name()), "echo modified >> "+outName)

	time.Sleep(time.Second)  // give time to run
	file.WriteString("mod1") // modify watched file
	time.Sleep(time.Second)  // give time to run

	if !fileExists(outName) {
		t.Errorf("nanny should have run and created target: %s", outName)
	}
	bs, err := ioutil.ReadFile(outName)
	if err != nil {
		t.Errorf("error reading target file: %s", err)
	}
	expect := []byte(`modified
`)
	if !bytes.Equal(expect, bs) {
		t.Errorf("target file contents wrong, got `%s`, want `%s`", bs, expect)
	}

	kill <- true
}
