package mypool

import (
	"context"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Task = func(ctx context.Context) error

type Executor interface {
	Run(ctx context.Context, threadNumber int) error
	Push(task Task) error
}

type Pool struct {
	logger *zap.Logger
	tasks  chan Task
}

func New(logger *zap.Logger, size int) *Pool {
	return &Pool{
		logger: logger,
		tasks:  make(chan Task, size),
	}
}

func (p *Pool) Push(task Task) error {
	p.tasks <- task
	return nil
}

func (p *Pool) Run(ctx context.Context, threadNumber int) error {
	group, currentCtx := errgroup.WithContext(ctx)

	for i := 0; i < threadNumber; i++ {
		group.Go(func() error {
			for {
				select {
				case task := <-p.tasks:
					if err := task(currentCtx); err != nil {
						p.logger.Error("Task executed with error", zap.Error(err))
						return err
					}

				case <-currentCtx.Done():
					return ctx.Err()
				}
			}
		})
	}

	p.logger.Info("Worker pool ran with", zap.Int(" thread of number", threadNumber))

	return group.Wait()
}
