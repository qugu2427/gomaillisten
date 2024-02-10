package main

import "fmt"

func main() {
	cfg := ListenConfig{
		nil,
		false,
		"0.0.0.0:25",
		24576,
		24576 * 1000,
		func(mail *Mail) {},
		func(lev LogLevel, msg string) {
			fmt.Println(msg)
		},
		[]string{"localhost", "bob.org"},
		"localhost",
	}
	err := Listen(cfg)
	if err != nil {
		fmt.Println(err)
	}
}
