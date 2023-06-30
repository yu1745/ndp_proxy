package utils

import (
	"log"
	"os/exec"
	"strings"
)

func ExecCmdWithSpaces(command string) error {
	args := strings.Split(command, " ")
	log.Printf("exec cmd: %s\n", command)
	err := exec.Command(args[0], args[1:]...).Run()
	if err != nil {
		return err
	}
	return nil
}
