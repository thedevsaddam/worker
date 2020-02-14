package main

import (
	"log"
	"os"
	"time"

	"github.com/thedevsaddam/worker"
)

func main() {

	logger := log.New(os.Stderr, "XXX", log.Ltime)
	b := worker.New(worker.WithConcurrency(2), worker.WithDebug(), worker.WithLogger(logger))

	duration := 1 * time.Second

	log.Println("Enqueue task 1")
	b.Register(func() {
		log.Println("processing: tom 1")
		time.Sleep(duration)
		log.Println("completed: tom 1")
	})

	log.Println("Enqueue task 2")
	b.Register(func() {
		log.Println("processing: jerry 2")
		time.Sleep(duration)
		log.Println("completed: jerry 2")
	})

	log.Println("Enqueue task 3")
	b.Register(func() {
		log.Println("processing: john 3")
		time.Sleep(duration)
		log.Println("completed: john 3")
	})

	log.Println("Enqueue task 4")
	b.Register(func() {
		log.Println("processing: joe 4")
		time.Sleep(duration)
		log.Println("completed: joe 4")
	})

	log.Println("Enqueue task 5")
	b.Register(func() {
		log.Println("processing: jack 5")
		time.Sleep(duration)
		log.Println("completed: jack 5")
	})

	log.Println("Enqueue task 6")
	b.Register(func() {
		log.Println("processing: sparrow 6")
		time.Sleep(duration)
		log.Println("completed: sparrow 6")
	})

	log.Println("Enqueue task 7")
	b.Register(func() {
		log.Println("processing: witcher 7")
		time.Sleep(duration)
		log.Println("completed: witcher 7")
	})

	// run background worker
	b.Run()
}
