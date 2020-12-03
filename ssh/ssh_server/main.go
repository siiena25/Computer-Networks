package main

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os/exec"
	"strings"
)

func main() {
	ssh.Handle(func(s ssh.Session) {
		command := s.Command()
		if len(command) == 0 {
			print("")
		} else {
			exe, _ := exec.Command(command[0], command[1:]...).Output()
			io.WriteString(s, string(exe))
		}

		term := terminal.NewTerminal(s, "")
		for {
			line, err := term.ReadLine()
			if err != nil {
				break
			}
			in := strings.Split(line, " ")
			if in[0] == "" {
				continue
			}
			if in[0] == "exit" {
				break
			}
			exe := exec.Command(in[0], in[1:]...)
			out, err := exe.Output()
			if err != nil {
				fmt.Println(err)
			}
			term.Write(out)
		}
	})
	ssh.ListenAndServe(":2231", nil,
		ssh.PasswordAuth(func(context ssh.Context, password string) bool {
			return context.User() == "kostyunina" && password == "12345"
		}),
	)
}