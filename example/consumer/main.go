package main

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/thedevsaddam/gworker"
)

var counter int32
var t = time.Now()

func SendEmail(v string) error {
	atomic.AddInt32(&counter, 1)
	time.Sleep(500 * time.Millisecond)
	log.Println("Counter:", atomic.LoadInt32(&counter))
	fmt.Println("Time taken:", time.Since(t))
	return nil
}

func main() {
	w := gworker.New("emailer")
	if err := w.RegisterJob("send_email", SendEmail); err != nil {
		panic(err)
	}

	w.Run()
}
