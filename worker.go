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

const (
	packageName        = "worker"
	defaultConcurrency = 100
)

// Worker represents a background worker type
type Worker struct {
	option option
	id     uint64
	queue  []func()  // list of functions
	jobs   chan bool // jobs
	stop   chan os.Signal
}

// New return a Worker instance
func New(options ...OptionFunc) *Worker {
	//default option
	defaultOptions := option{
		concurrency: defaultConcurrency,
		logger:      log.New(os.Stderr, packageName+": ", log.Ldate|log.Ltime),
	}

	w := &Worker{
		option: defaultOptions,
		queue:  make([]func(), 0),
		jobs:   make(chan bool, defaultOptions.concurrency),
		stop:   make(chan os.Signal, 1),
	}

	for _, option := range options {
		if err := option(w); err != nil {
			w.option.logger.Printf("%s: %s", packageName, err.Error())
		}
	}

	if w.option.concurrency > 0 {
		w.jobs = make(chan bool, w.option.concurrency)
	}

	return w
}

// Register add a new task to the worker
func (w *Worker) Register(j func()) *Worker {
	w.queue = append(w.queue, j)
	return w
}

// Run start the backgournd worker for processing jobs
func (w *Worker) Run() {

	fLog := func(format string, v ...interface{}) {
		if w.option.debug {
			w.option.logger.Printf(format, v...)
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 34))
	fmt.Println("| Backgound job worker started...|")
	fmt.Println(strings.Repeat("-", 34))

	signal.Notify(w.stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)
	for {

		select {

		case <-w.stop:
			close(w.jobs)
			close(w.stop)

			fmt.Println()
			fmt.Println(strings.Repeat("-", 40))
			fmt.Println("| Shutting down background job worker! |")
			fmt.Println(strings.Repeat("-", 40))

			os.Exit(0)

		default:
			// process JOB in FIFO
			if len(w.queue) > 0 {
				w.jobs <- true
				f := w.queue[0]       // asssign first task to worker
				w.queue = w.queue[1:] // remove first task from queue

				go func(w *Worker, f func()) {
					atomic.AddUint64(&w.id, 1)

					fLog("Processing: Task %d\n", w.id)

					f()

					fLog("Completed: Task %d\n", w.id)
					<-w.jobs
				}(w, f)
			}

		}

	}
}
