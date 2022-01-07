package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type sticky struct {
	path     string
	name     string
	contents []fs.DirEntry
}

func stickiesDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	ret := filepath.Join(home, "Library/Containers/com.apple.Stickies/Data/Library/Stickies")
	return ret, nil
}

func isRTFD(f fs.DirEntry) bool {
	return f.IsDir() && filepath.Ext(f.Name()) == ".rtfd"
}

func isRTFTextData(f fs.DirEntry) bool {
	return f.Name() == "TXT.rtf"
}

func listStickies() ([]*sticky, error) {
	stickies := make([]*sticky, 0)
	baseDir, err := stickiesDataDir()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !isRTFD(file) {
			continue
		}
		rtfdDir := filepath.Join(baseDir, file.Name())
		rtfdFiles, err := os.ReadDir(rtfdDir)
		if err != nil {
			return nil, err
		}
		sticky := &sticky{
			path:     rtfdDir,
			name:     file.Name(),
			contents: rtfdFiles,
		}
		stickies = append(stickies, sticky)
	}
	return stickies, nil
}

func openEditor(filepath string) error {
	cmd := exec.Command("vi", "./foo")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func restartStickies() error {
	if err := exec.Command("killall", "Stickies").Run(); err != nil {
		return err
	}
	if err := exec.Command("open", "-a", "stickies").Run(); err != nil {
		return err
	}
	return nil
}

func checkOS() error {
	if runtime.GOOS != "darwin" {
		return errors.New("unsupported os")
	}
	return nil
}

func run(args []string) error {
	if err := checkOS(); err != nil {
		return err
	}
	stickies, err := listStickies()
	if err != nil {
		return err
	}
	for _, sticky := range stickies {
		fmt.Println(sticky.path)
	}
	return err
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}
