package main

import (
	"fmt"

	"github.com/thedevsaddam/gworker"
)

func main() {
	w := gworker.New("emailer")
	for i := 1; i <= 10; i++ {
		fmt.Println("sending job", i)
		w.SendJob(&gworker.Job{
			Name: "send_email",
			Args: []gworker.Arg{
				{Type: "string", Value: "Hello"},
			},
		})
	}

}
