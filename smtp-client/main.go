package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strings"
)

type Data struct {
	Host string
	Port string
	From string
}

func main() {
	file, err := ioutil.ReadFile("data_mail.json")
	if err != nil {
		log.Fatal(err)
	}
	var data Data
	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Fatal(err)
	}
	var password string
	fmt.Print("Введите пароль: ")
	fmt.Scanf("%s", &password)
	auth := smtp.PlainAuth("", data.From, password, data.Host)

	fmt.Print("Введите отправителя: ")
	in := bufio.NewReader(os.Stdin)
	input, _, _ := in.ReadLine()
	to := strings.Split(string(input), " ")

	message := ""
	fmt.Print("Введите тему: ")
	input, _, _ = in.ReadLine()
	message += "Subject: " + string(input) + "\r\n\r\n"
	fmt.Print("Введите сообщение: ")
	input, _, _ = in.ReadLine()
	message += string(input) + "\r\n"

	fmt.Println("Отправление...")
	fmt.Println(data.Host + ":" + data.Port)
	err = smtp.SendMail(data.Host + ":" + data.Port, auth, data.From, to, []byte(message))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Письмо отправилось")
}