package edist

import (
	"crypto/md5"
	"errors"
	"os"
	"os/exec"
	"runtime"
)

func edit(filepath string) (updated bool, err error) {
	before, err := getMD5(filepath)
	if err != nil {
		return
	}
	err = openEditor(filepath)
	if err != nil {
		return
	}
	after, err := getMD5(filepath)
	if err != nil {
		return
	}
	updated = before != after
	return
}

func getMD5(filepath string) ([16]byte, error) {
	binary, err := os.ReadFile(filepath)
	if err != nil {
		return [16]byte{}, err
	}
	return md5.Sum(binary), nil
}

func openEditor(filepath string) error {
	cmd := exec.Command("vi", filepath)
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
