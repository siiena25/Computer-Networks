/*package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
)

var (
	user = flag.String("u", "kostyunina", "User name")
	password = flag.String("pass", "12345", "Private key file")
	//host1 = flag.String("h", "185.20.226.174", "Host")
	host2 = flag.String("h", "localhost", "Host")
	//port1 = flag.Int("p", 22, "Port")
	port2 = flag.Int("p", 2231, "Port")
)

func main() {
	flag.Parse()

	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", *host2, *port2)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		panic(err)
	}

	for {
		session, err := client.NewSession()
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer session.Close()

		fmt.Print("> ")
		command, _, err := bufio.NewReader(os.Stdin).ReadLine()
		if err != nil {
			fmt.Println(err)
			continue
		}

		b, err := session.Output(string(command))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Print(string(b))
	}
}*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"
)

var (
	user = flag.String("u", "kostyunina", "User name")
	password = flag.String("pass", "12345", "Private key file")
)

func executeCmd(cmd, hostname string, config *ssh.ClientConfig) string {
	t0 := time.Now()

	conn, _ := ssh.Dial("tcp", hostname, config)
	session, _ := conn.NewSession()
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	t1 := time.Now()
	fmt.Printf("Time for %s: %v\n", hostname, t1.Sub(t0))

	return "Output: " + stdoutBuf.String()
}

func main() {
	flag.Parse()

	cmd := "ls"
	var hosts = [2]string{"localhost:2231", "localhost:2232"}

	results := make(chan string, 10)

	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	for _, hostname := range hosts {
		go func(hostname string) {
			results <- executeCmd(cmd, hostname, config)
		}(hostname)
	}

	for i := 0; i < len(hosts); i++ {
		select {
		case res := <-results:
			fmt.Print(res)
			fmt.Print("\n")
		}
	}
}