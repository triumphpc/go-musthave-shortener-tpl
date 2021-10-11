package worker

import (
	"context"
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"runtime"
	"sync"
)

type Pool struct {
	// Workers pool
	workerPool []*Worker
	// Logger
	logger *zap.Logger
	// Database object
	db *sql.DB
	// Queue for marks as deleted
	queue *Queue
	// Waiting group for worker goroutines
	wg *sync.WaitGroup
	// Total counter
	total chan int
	// Has workers for work
	isAvail bool
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
}

// Task in queue updating link ids for user id
type Task struct {
	ids    []string
	userID string
}

// sqlUpdate for set delete flag
const sqlUpdate = `
	UPDATE storage.short_links 
	SET is_deleted=true 
	WHERE user_id=$1 
	AND (correlation_id = ANY($2) OR short=ANY($3))
`

// New Instance new pool
func New(ctx context.Context, db *sql.DB, l *zap.Logger) *Pool {
	p := &Pool{logger: l, db: db}

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
	p.isAvail = true

	// If we have some error in goroutine
	go func() {
		if err := g.Wait(); err != nil {
			p.logger.Info("Pool error", zap.Error(err))
		}
	}()

	// When all goroutines closed
	go func() {
		p.wg.Wait()
		p.isAvail = false
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

	return p
}

// newQueue Init new queue for workers
func (p *Pool) newQueue() *Queue {
	q := Queue{}
	// Condition for mutex use
	q.cond = sync.NewCond(&q.mu)
	return &q
}

// newWorker constructor
func (p *Pool) newWorker(id int) *Worker {
	p.logger.Info("Init new worker", zap.Int("id", id))
	return &Worker{id, p}
}

// loop listen for new tasks
func (w *Worker) loop(ctx context.Context) error {
	defer func() {
		w.pool.logger.Info("Close worker", zap.Int("worker id", w.id))
		// Close counter
		w.pool.wg.Done()

		select {
		case <-ctx.Done():
			w.pool.logger.Info("Aborting from ctx")
		}

	}()

	for {
		t := w.pool.queue.PopWait()
		// bunch update
		w.pool.logger.Info("Worker get new task", zap.Int("worker id", w.id))

		if err := w.bunchUpdateAsDeleted(ctx, t.ids, t.userID); err != nil {
			w.pool.logger.Info("Run to out from loop ")
			return err
		}
		// Write len for counter
		w.pool.total <- len(t.ids)
	}
}

// PopWait Get task for update from fanOut
func (q *Queue) PopWait() *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.arr) == 0 {
		// Goroutines to sleep
		// After awake again take mutex
		q.cond.Wait()
	}
	t := q.arr[0]
	q.arr = q.arr[1:]

	return t
}

// Push lock all goroutines for push task in queue
func (p *Pool) Push(ids []string, userID string) bool {
	// Check if workers has
	if !p.isAvail {
		return false
	}
	p.logger.Info("Push new task in queue")
	// New Task for updating
	t := Task{ids, userID}
	// Lock input queue
	p.queue.mu.Lock()
	defer p.queue.mu.Unlock()
	// Add to queue new Task for update
	p.queue.arr = append(p.queue.arr, &t)
	//awake PopWait in worker
	p.queue.cond.Signal()
	return true
}

// bunchUpdateAsDeleted  update as deleted
func (w *Worker) bunchUpdateAsDeleted(ctx context.Context, ids []string, userID string) error {
	return errors.New("TEST")
	//if len(ids) == 0 {
	//	return nil
	//}
	//
	//idsArr := pq.Array(ids)
	//_, err := w.pool.db.ExecContext(ctx, sqlUpdate, userID, idsArr, idsArr)
	//
	//return err
}
