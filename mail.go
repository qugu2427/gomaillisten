package main

import (
	"net"

	spf "github.com/qugu2427/gospf"
)

type MailHandler = func(mail *Mail)

type Mail struct {
	SenderAddr net.Addr
	MailFrom   string
	Raw        string
	SpfResult  spf.Result
}
