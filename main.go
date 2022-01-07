package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"runtime"
)

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
	return nil
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}
