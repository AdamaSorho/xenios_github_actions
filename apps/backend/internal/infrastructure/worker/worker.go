package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// JobHandler is a function that processes a job and returns an error if processing fails.
type JobHandler func(ctx context.Context, job *entities.Job) error

// Worker polls the job queue and processes jobs using registered handlers.
type Worker struct {
	queue        repository.JobQueue
	handlers     map[entities.JobType]JobHandler
	pollInterval time.Duration
	jobTimeout   time.Duration
	mu           sync.RWMutex
	running      bool
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewWorker creates a new Worker with the given job queue and options.
func NewWorker(queue repository.JobQueue, pollInterval, jobTimeout time.Duration) *Worker {
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	if jobTimeout <= 0 {
		jobTimeout = 5 * time.Minute
	}

	return &Worker{
		queue:        queue,
		handlers:     make(map[entities.JobType]JobHandler),
		pollInterval: pollInterval,
		jobTimeout:   jobTimeout,
		stopCh:       make(chan struct{}),
	}
}

// RegisterHandler registers a handler for a specific job type.
func (w *Worker) RegisterHandler(jobType entities.JobType, handler JobHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[jobType] = handler
}

// RegisteredJobTypes returns the list of job types this worker handles.
func (w *Worker) RegisteredJobTypes() []entities.JobType {
	w.mu.RLock()
	defer w.mu.RUnlock()

	types := make([]entities.JobType, 0, len(w.handlers))
	for jt := range w.handlers {
		types = append(types, jt)
	}
	return types
}

// Start begins the worker polling loop in a goroutine.
// It is safe to call Start multiple times; subsequent calls are no-ops.
func (w *Worker) Start(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.stopCh = make(chan struct{})
	w.mu.Unlock()

	w.wg.Add(1)
	go w.pollLoop(ctx)
}

// Stop signals the worker to stop and waits for the current job to finish.
func (w *Worker) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	close(w.stopCh)
	w.mu.Unlock()

	w.wg.Wait()
}

// IsRunning returns whether the worker is currently running.
func (w *Worker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

func (w *Worker) pollLoop(ctx context.Context) {
	defer w.wg.Done()

	log.Println("Worker started, polling for jobs...")

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			log.Println("Worker stopping...")
			return
		case <-ctx.Done():
			log.Println("Worker context cancelled, stopping...")
			return
		case <-ticker.C:
			w.processNextJob(ctx)
		}
	}
}

func (w *Worker) processNextJob(ctx context.Context) {
	jobTypes := w.RegisteredJobTypes()
	if len(jobTypes) == 0 {
		return
	}

	job, err := w.queue.Dequeue(ctx, jobTypes)
	if err != nil {
		log.Printf("Error dequeuing job: %v", err)
		return
	}
	if job == nil {
		return // No jobs available
	}

	log.Printf("Processing job %s (type=%s, attempt=%d/%d)", job.ID, job.Type, job.Attempt, job.MaxAttempts)

	w.mu.RLock()
	handler, exists := w.handlers[job.Type]
	w.mu.RUnlock()

	if !exists {
		log.Printf("No handler for job type %s, marking as failed", job.Type)
		if failErr := w.queue.Fail(ctx, job.ID, "no handler registered for job type"); failErr != nil {
			log.Printf("Error marking job %s as failed: %v", job.ID, failErr)
		}
		return
	}

	// Execute the handler with a timeout
	jobCtx, cancel := context.WithTimeout(ctx, w.jobTimeout)
	defer cancel()

	if err := handler(jobCtx, job); err != nil {
		log.Printf("Job %s failed: %v", job.ID, err)
		if failErr := w.queue.Fail(ctx, job.ID, err.Error()); failErr != nil {
			log.Printf("Error marking job %s as failed: %v", job.ID, failErr)
		}
		return
	}

	log.Printf("Job %s completed successfully", job.ID)
	if err := w.queue.Complete(ctx, job.ID); err != nil {
		log.Printf("Error marking job %s as complete: %v", job.ID, err)
	}
}
