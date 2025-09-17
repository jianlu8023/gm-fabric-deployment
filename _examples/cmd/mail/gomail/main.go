package main

import (
	"gopkg.in/gomail.v2"
)

func main() {
	m := gomail.NewMessage()
	m.SetHeader("From", "sender@gmail.com")
	m.SetHeader("To", "recipient@gmail.com")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "<b>Hello!</b>")
	m.Attach("demo.txt")
	d := gomail.NewDialer("smtp.gmail.com", 587, "sender@gmail.com", "password")
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}

}
