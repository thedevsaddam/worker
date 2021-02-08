package gworker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	packageName        = "gsworker"
	defaultConcurrency = 10
)

var ctx = context.Background()

var c = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

// Worker represents a background worker type
type Worker struct {
	mu      *sync.Mutex
	name    string
	jobs    chan *Job                // track concurrency limit
	jobFunc map[string]reflect.Value // jobs
	stop    chan os.Signal           // register worker to os signals
	process chan struct{}
	option  option
}

// New return a Worker instance
func New(name string, options ...OptionFunc) *Worker {
	//default option
	defaultOptions := option{
		concurrency: defaultConcurrency,
		logger:      log.New(os.Stderr, packageName+": ", log.Ldate|log.Ltime),
	}

	w := &Worker{
		mu:      &sync.Mutex{},
		option:  defaultOptions,
		stop:    make(chan os.Signal, 1),
		jobFunc: make(map[string]reflect.Value),
		jobs:    make(chan *Job, 1),
		process: make(chan struct{}, 100),
	}

	for _, option := range options {
		if err := option(w); err != nil {
			w.option.logger.Printf("%s: %s", packageName, err.Error())
		}
	}

	w.name = name
	w.option.redisClient = c
	return w
}

// RegisterJob add a new task to the worker
func (w *Worker) RegisterJob(name string, fn interface{}) error {

	if _, exist := w.jobFunc[name]; exist {
		return ErrJobConflict
	}

	vfn := reflect.ValueOf(fn)

	// if the fn is not a function then return error
	if vfn.Type().Kind() != reflect.Func {
		return ErrInvalidJob
	}

	// if the function does not return anything, we can't catch if an error occur or not
	if vfn.Type().NumOut() <= 0 {
		return ErrInvalidJobReturn
	}

	w.jobFunc[name] = vfn

	return nil
}

// SendJob add a new task to the worker
func (w *Worker) SendJob(j *Job) error {
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return w.option.redisClient.RPush(ctx, fmt.Sprintf("%s:%s", packageName, w.name), b).Err()
}

func (w *Worker) listenMessages() {
	for {
		result, err := w.option.redisClient.BLPop(ctx, 1*time.Second, fmt.Sprintf("%s:%s", packageName, w.name)).Result()
		if err != nil {
			if err != redis.Nil {
				log.Println(err)
			}
		}
		if result != nil {
			w.process <- struct{}{}
			j := &Job{}
			if err := json.Unmarshal([]byte(result[1]), j); err == nil {
				w.jobs <- j
			}
		}
	}
}

// Run start the backgournd worker for processing jobs
func (w *Worker) Run() {

	fmt.Println()
	fmt.Println(strings.Repeat("-", 34))
	fmt.Println("| Backgound job worker started...|")
	fmt.Println(strings.Repeat("-", 34))

	go func(lw *Worker) {
		w.listenMessages()
	}(w)

	signal.Notify(w.stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	for {

		select {

		case <-w.stop:
			fmt.Println()
			fmt.Println(strings.Repeat("-", 40))
			fmt.Println("| Shutting down background job worker! |")
			fmt.Println(strings.Repeat("-", 40))

			close(w.jobs)
			close(w.stop)

			os.Exit(0)

		case j := <-w.jobs:
			if j != nil {
				go func(job Job) {
					in := make([]reflect.Value, 0)
					for _, a := range job.Args {
						in = append(in, reflect.ValueOf(a.Value))
					}

					// call the fn with arguments
					out := []interface{}{}
					for _, o := range w.jobFunc[job.Name].Call(in) {
						out = append(out, o.Interface())
					}
					<-w.process
				}(*j)
			}
		}
	}
}
