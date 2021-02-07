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
	mu     *sync.Mutex
	name   string
	jobs   map[string]reflect.Value // jobs
	stop   chan os.Signal           // register worker to os signals
	limit  chan struct{}            // track concurrency limit
	option option
}

// New return a Worker instance
func New(name string, options ...OptionFunc) *Worker {
	//default option
	defaultOptions := option{
		concurrency: defaultConcurrency,
		logger:      log.New(os.Stderr, packageName+": ", log.Ldate|log.Ltime),
	}

	w := &Worker{
		mu:     &sync.Mutex{},
		option: defaultOptions,
		stop:   make(chan os.Signal, 1),
		limit:  make(chan struct{}, defaultConcurrency),
		jobs:   make(map[string]reflect.Value),
	}

	for _, option := range options {
		if err := option(w); err != nil {
			w.option.logger.Printf("%s: %s", packageName, err.Error())
		}
	}
	if w.option.concurrency > 0 {
		w.limit = make(chan struct{}, w.option.concurrency)
	}

	w.name = name
	w.option.redisClient = c
	return w
}

// RegisterJob add a new task to the worker
func (w *Worker) RegisterJob(name string, fn interface{}) error {

	if _, exist := w.jobs[name]; exist {
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

	w.jobs[name] = vfn

	return nil
}

// SendJob add a new task to the worker
func (w *Worker) SendJob(j *Job) error {
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return w.option.redisClient.LPush(ctx, fmt.Sprintf("%s:%s", packageName, w.name), b).Err()
}

// Run start the backgournd worker for processing jobs
func (w *Worker) Run() {

	// fLog := func(format string, v ...interface{}) {
	// 	if w.option.debug {
	// 		w.option.logger.Printf(format, v...)
	// 	}
	// }

	fmt.Println()
	fmt.Println(strings.Repeat("-", 34))
	fmt.Println("| Backgound job worker started...|")
	fmt.Println(strings.Repeat("-", 34))

	signal.Notify(w.stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)
	for {

		select {

		case <-w.stop:
			fmt.Println()
			fmt.Println(strings.Repeat("-", 40))
			fmt.Println("| Shutting down background job worker! |")
			fmt.Println(strings.Repeat("-", 40))

			close(w.limit)
			close(w.stop)

			os.Exit(0)

		default:
			w.limit <- struct{}{}
			go func() {
				for { //TODO:: in cofusion
					result, err := w.option.redisClient.BRPop(ctx, 1*time.Second, fmt.Sprintf("%s:%s", packageName, w.name)).Result()
					if err != nil {
						if err != redis.Nil {
							log.Println(err)
						}
					}
					if result != nil {
						j := Job{}
						json.Unmarshal([]byte(result[1]), &j)
						in := make([]reflect.Value, 0)
						for _, a := range j.Args {
							in = append(in, reflect.ValueOf(a.Value))
						}

						// call the fn with arguments
						out := []interface{}{}
						for _, o := range w.jobs[j.Name].Call(in) {
							out = append(out, o.Interface())
						}
					}
				}
				<-w.limit

			}()
		}

	}
}
