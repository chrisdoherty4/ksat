package ksat

import "context"

// Task is a runnable unit of work.
type Task interface {
	Run(context.Context) error
}

// Func is a convinience type that allows consumers to define inline functions as a Task.
type Func func(context.Context) error

func (fn Func) Run(ctx context.Context) error {
	return fn(ctx)
}

// ErrorHandler is used to handle errors originating from Tasks.
type ErrorHandler func(error)

type task struct {
	t Task
	e ErrorHandler
}

// List is a list of runnable tasks. Tasks are executed concurrently.
type List struct {
	tasks []task
}

// Add adds t to the list of tasks to be run.
func (l *List) Add(t Task, e ErrorHandler) {
	l.tasks = append(l.tasks, task{t: t, e: e})
}

// Run runs all tasks in l. Tasks are executed out of order and concurrently. When an error occurs
// the registered error handler is called. When Run() completes, l is reset.
func (l *List) Run(ctx context.Context) {
	for _, tsk := range l.tasks {
		go func(task Task, handler ErrorHandler) {
			if err := task.Run(ctx); err != nil {
				handler(err)
			}
		}(tsk.t, tsk.e)
	}
	l.tasks = nil
}

// Chain is a list of chained tasks. Tasks are executed in the order they are added.
type Chain struct {
	tasks []Task
}

// Add adds all t's to c in the order they are specified.
func (c *Chain) Add(t ...Task) {
	c.tasks = append(c.tasks, t...)
}

// Run runs all tasks in c. Tasks are run in the order they were added. If a task fails the error
// is returned and subsequent tasks are not run. When Run() completes, c is reset.
func (c *Chain) Run(ctx context.Context) error {
	for _, tsk := range c.tasks {
		if err := tsk.Run(ctx); err != nil {
			return err
		}
	}
	c.tasks = nil
	return nil
}
