package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/thedevsaddam/gworker"
)

var counter int32

func SendEmail(v string) error {
	atomic.AddInt32(&counter, 1)
	time.Sleep(1 * time.Second)
	fmt.Println("Counter:", atomic.LoadInt32(&counter))
	return nil
}

func main() {
	w := gworker.New("emailer", gworker.WithConcurrency(5))
	if err := w.RegisterJob("send_email", SendEmail); err != nil {
		panic(err)
	}

	go func() {
		for i := 1; i <= 5; i++ {
			w.SendJob(&gworker.Job{
				Name: "send_email",
				Args: []gworker.Arg{
					{Type: "string", Value: "Hello"},
				},
			})
		}
	}()

	w.Run()
}
