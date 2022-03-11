// Package worker contain logic for works pool
package worker

import (
	"context"
	"runtime"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/repository"
)

type Pool struct {
	// Workers pool
	workerPool []*Worker
	// Logger
	logger *zap.Logger
	// Queue for marks as deleted
	queue *Queue
	// Waiting group for worker goroutines
	wg *sync.WaitGroup
	// Total counter
	total chan int
	// Storage of users
	storage repository.Repository
}

type IPool interface {
	// Push ids for working
	Push(ids []string, userID string)
	// Run init pool of workers
	Run(ctx context.Context)
}

// Worker for process
type Worker struct {
	id   int
	pool *Pool
}

// Queue for tasks in worker
type Queue struct {
	arr  []*Task
	mu   sync.Mutex
	cond *sync.Cond
	stop bool
}

// Task in queue updating link ids for user id
type Task struct {
	ids    []string
	userID string
}

// New Instance new pool
func New(ctx context.Context, l *zap.Logger, s repository.Repository) (*Pool, func()) {
	p := &Pool{logger: l, storage: s}

	p.logger.Info("Init new worker pool")
	// Init new worker pool
	p.workerPool = make([]*Worker, 0, runtime.NumCPU())
	// Init queue for tasks
	p.queue = p.newQueue()
	// Make workers
	for i := 0; i < runtime.NumCPU(); i++ {
		p.workerPool = append(p.workerPool, p.newWorker(i))
	}
	// Run all workers in goroutines
	ctx, cancel := context.WithCancel(ctx)
	// Make error group for goroutines
	g, _ := errgroup.WithContext(ctx)
	p.wg = &sync.WaitGroup{}

	for _, w := range p.workerPool {
		p.wg.Add(1)
		worker := w
		f := func() error {
			return worker.loop(ctx)
		}
		g.Go(f)
	}
	// If we have some error in goroutine
	go func() {
		if err := g.Wait(); err != nil {
			p.logger.Info("Pool error", zap.Error(err))
		}
	}()
	// When all goroutines closed
	go func() {
		p.wg.Wait()
		close(p.total)
		// close context
		cancel()
	}()
	// monitor for total updates
	p.total = make(chan int)
	go func() {
		total := 0
		for c := range p.total {
			total = total + c
		}
		// Out how many updates
		p.logger.Info("Total updated", zap.Int("count", total))
	}()

	return p, p.Close
}

// newQueue Init new queue for workers
func (p *Pool) newQueue() *Queue {
	q := Queue{}
	// Condition for mutex use
	q.cond = sync.NewCond(&q.mu)
	q.stop = false
	return &q
}

// newWorker constructor
func (p *Pool) newWorker(id int) *Worker {
	p.logger.Info("Init new worker", zap.Int("id", id))
	return &Worker{id, p}
}

// Close grace shutdown handler
func (p *Pool) Close() {
	p.queue.close()
}

// close pool for clients
func (q *Queue) close() {
	q.cond.L.Lock()
	q.stop = true
	// Broadcast that workers must close
	q.cond.Broadcast()
	q.cond.L.Unlock()
}

// loop listen for new tasks
func (w *Worker) loop(ctx context.Context) error {
	defer func() {
		w.pool.logger.Info("Close worker", zap.Int("worker id", w.id))
		// Close counter
		w.pool.wg.Done()
		// Close queue
		w.pool.queue.close()

		<-ctx.Done()
		w.pool.logger.Info("Aborting from ctx", zap.Int("worker id", w.id))
	}()

	for {
		t, ok := w.pool.queue.PopWait()
		// Check if pool available
		if !ok {
			return nil
		}
		// bunch update
		w.pool.logger.Info("Worker get new task", zap.Int("worker id", w.id))

		if err := w.pool.storage.BunchUpdateAsDeleted(ctx, t.ids, t.userID); err != nil {
			w.pool.logger.Info("Run to out from loop ")
			return err
		}
		// Write len for counter
		w.pool.total <- len(t.ids)
	}
}

// PopWait Get task for update from fanOut
func (q *Queue) PopWait() (*Task, bool) {
	q.cond.L.Lock()

	for len(q.arr) == 0 && !q.stop {
		// Goroutines to sleep
		// After awake again take mutex
		q.cond.Wait()
	}

	if q.stop {
		q.cond.L.Unlock()
		return nil, false
	}

	// If pool is available
	t := q.arr[0]
	q.arr = q.arr[1:]

	q.cond.L.Unlock()

	return t, true
}

// Push lock all goroutines for push task in queue
func (p *Pool) Push(ids []string, userID string) bool {
	// Check if workers has
	if p.queue.stop {
		return false
	}

	// Lock input queue
	p.queue.cond.L.Lock()
	defer p.queue.cond.L.Unlock()

	p.logger.Info("Push new task in queue")
	// New Task for updating
	t := Task{ids, userID}
	// Add to queue new Task for update
	p.queue.arr = append(p.queue.arr, &t)
	//awake PopWait in worker
	p.queue.cond.Signal()
	return true
}

//// bunchUpdateAsDeleted  update as deleted
//func (w *Worker) bunchUpdateAsDeleted(ctx context.Context, ids []string, userID string) error {
//	if len(ids) == 0 {
//		return nil
//	}
//
//	idsArr := pq.Array(ids)
//	_, err := w.pool.db.ExecContext(ctx, sqlUpdate, userID, idsArr, idsArr)
//
//	return err
//}
