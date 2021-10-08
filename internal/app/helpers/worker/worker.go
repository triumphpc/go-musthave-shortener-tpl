package worker

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	"go.uber.org/zap"
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
	// Make error group
	errCh chan error
	// Waiting group for worker goroutines
	wg *sync.WaitGroup
	// Total counter
	total int
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
func New(db *sql.DB, l *zap.Logger) *Pool {
	return &Pool{logger: l, db: db}
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
func (w *Worker) loop(ctx context.Context) {
	var defErr error

	defer func() {
		w.pool.logger.Info("Close worker", zap.Int("worker id", w.id))
		select {
		// Send all to chan
		case w.pool.errCh <- defErr:
		case <-ctx.Done():
			w.pool.logger.Info("Aborting from ctx")
		}
		// Close counter
		w.pool.wg.Done()
	}()

exit:
	for {
		t := w.pool.queue.PopWait()
		// bunch update
		w.pool.logger.Info("Worker get new task", zap.Int("worker id", w.id))

		if err := w.bunchUpdateAsDeleted(ctx, t.ids, t.userID); err != nil {
			defErr = err
			w.pool.logger.Info("Run to out from loop ")
			break exit
		}
		// On ranIn single channel (it's only example how we can use channels)
		// And show fanOut fanIn pattern
		inputCh := make(chan string)
		// Put in channel all ids
		go func() {
			for _, id := range t.ids {
				inputCh <- id
			}
			close(inputCh)
		}()
		// Redirection chan (only example)
		for range w.pool.fanIn(inputCh) {
			w.pool.total++
		}
	}
	w.pool.logger.Info("To end")
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
	if p.isAvail == false {
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

// fanIn create simple chan from worker pools (example only)
func (p *Pool) fanIn(ch <-chan string) (out chan string) {
	out = make(chan string)

	go func() {
		// wg example
		wg := &sync.WaitGroup{}

		for item := range ch {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				out <- id

			}(item)
		}
		wg.Wait()
		close(out)
	}()
	return out
}

// Run worker pool
func (p *Pool) Run(ctx context.Context) {
	p.logger.Info("Init new worker pool")
	// Init new worker pool
	p.workerPool = make([]*Worker, 0, runtime.NumCPU())
	// Init queue for tasks
	p.queue = p.newQueue()
	// Make workers
	for i := 0; i < runtime.NumCPU(); i++ {
		p.workerPool = append(p.workerPool, p.newWorker(i))
	}
	// Make error group chan
	p.errCh = make(chan error)
	// Run all workers in goroutines
	ctx, cancel := context.WithCancel(ctx)
	p.wg = &sync.WaitGroup{}

	for _, w := range p.workerPool {
		p.wg.Add(1)
		go w.loop(ctx)
	}
	p.isAvail = true

	// Listening error group chan
	go func() {
		if err := <-p.errCh; err != nil {
			p.logger.Info("Pool error", zap.Error(err))
			cancel()
		}
	}()
	// When all goroutines closed
	go func() {
		p.wg.Wait()
		p.logger.Info("Close error chan")
		close(p.errCh)
		p.isAvail = false
		// Out how many updates
		p.logger.Info("Total updated", zap.Int("count", p.total))
	}()
}

// bunchUpdateAsDeleted  update as deleted
func (w *Worker) bunchUpdateAsDeleted(ctx context.Context, ids []string, userID string) error {
	if len(ids) == 0 {
		return nil
	}
	// Update in transaction
	idsArr := pq.Array(ids)
	_, err := w.pool.db.ExecContext(ctx, sqlUpdate, userID, idsArr, idsArr)

	return err
}
