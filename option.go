package gworker

import "github.com/go-redis/redis/v8"

// OptionFunc represents a contract for option func, it basically set options to Worker instance options
type OptionFunc func(*Worker) error

// option describes type for providing configuration options to Worker
type option struct {
	redisClient *redis.Client
	concurrency uint
	debug       bool
	logger      Logger
}

// WithConcurrency set concurrency of the worker
func WithConcurrency(c uint) OptionFunc {
	return func(w *Worker) error {
		w.option.concurrency = c
		return nil
	}
}

// WithDebug enable debug mode
func WithDebug() OptionFunc {
	return func(w *Worker) error {
		w.option.debug = true
		return nil
	}
}

// WithLogger pass custom logger
func WithLogger(l Logger) OptionFunc {
	return func(w *Worker) error {
		w.option.logger = l
		return nil
	}
}
