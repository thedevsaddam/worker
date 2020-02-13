package worker

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
)

// Worker represents a background worker type
type Worker struct {
	id          uint64
	queue       []func()  // list of functions
	concurrency uint      // no concurrent task
	jobs        chan bool // jobs
	stop        chan os.Signal
}

// New return a Worker instance
func New(concurrency uint) *Worker {
	return &Worker{
		queue:       make([]func(), 0),
		concurrency: concurrency,
		jobs:        make(chan bool, concurrency),
		stop:        make(chan os.Signal, 1),
	}
}

// Register add a new task to the worker
func (b *Worker) Register(f func()) *Worker {
	b.queue = append(b.queue, f)
	return b
}

// Run start the backgournd worker for processing jobs
func (b *Worker) Run() {
	fmt.Println()
	fmt.Println(strings.Repeat("-", 34))
	fmt.Println("| Backgound job worker started...|")
	fmt.Println(strings.Repeat("-", 34))

	signal.Notify(b.stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)
	for {
		select {

		case <-b.stop:
			fmt.Println()
			fmt.Println(strings.Repeat("-", 40))
			fmt.Println("| Shutting down background job worker! |")
			fmt.Println(strings.Repeat("-", 40))

			close(b.jobs)
			close(b.stop)
			os.Exit(0)

		default:
			// process JOB in FIFO
			if len(b.queue) > 0 {
				b.jobs <- true
				f := b.queue[0]       // asssign first task to worker
				b.queue = b.queue[1:] // remove first task from queue

				go func(b *Worker, f func()) {
					atomic.AddUint64(&b.id, 1)
					log.Println("Processing: Task ", b.id)

					f()

					log.Println("Completed: Task ", b.id)
					<-b.jobs
				}(b, f)
			}
			//slow down little bit
			// time.Sleep(100 * time.Millisecond)
		}
	}
}