package repository

import (
	"context"
	"log"
	"sync"

	"github.com/xenios/backend/internal/domain/entities"
)

// AsyncAuditRepository wraps an AuditRepository and performs LogEvent calls
// asynchronously via a buffered channel. Query calls are passed through
// synchronously. This prevents audit logging from blocking API responses.
type AsyncAuditRepository struct {
	inner   AuditRepositoryInterface
	eventCh chan *entities.AuditEvent
	wg      sync.WaitGroup
	stopCh  chan struct{}
	mu      sync.Mutex
	running bool
}

// AuditRepositoryInterface matches the domain AuditRepository interface.
// This avoids importing the domain/repository package from the adapter layer
// (which would create an import cycle since adapter already depends on domain).
type AuditRepositoryInterface interface {
	LogEvent(ctx context.Context, event *entities.AuditEvent) error
	Query(ctx context.Context, filter entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error)
}

// NewAsyncAuditRepository creates an async wrapper with the given buffer size.
func NewAsyncAuditRepository(inner AuditRepositoryInterface, bufferSize int) *AsyncAuditRepository {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	return &AsyncAuditRepository{
		inner:   inner,
		eventCh: make(chan *entities.AuditEvent, bufferSize),
		stopCh:  make(chan struct{}),
	}
}

// Start begins processing events from the channel.
func (r *AsyncAuditRepository) Start() {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	r.stopCh = make(chan struct{})
	r.mu.Unlock()

	r.wg.Add(1)
	go r.processLoop()
}

// Stop signals the processor to stop and waits for all pending events to drain.
func (r *AsyncAuditRepository) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.running = false
	close(r.stopCh)
	r.mu.Unlock()

	r.wg.Wait()
}

// LogEvent enqueues the event for async processing. If the buffer is full,
// the event is logged synchronously as a fallback.
func (r *AsyncAuditRepository) LogEvent(ctx context.Context, event *entities.AuditEvent) error {
	select {
	case r.eventCh <- event:
		return nil
	default:
		// Buffer full — log synchronously as fallback to avoid data loss
		log.Printf("audit buffer full, logging synchronously")
		return r.inner.LogEvent(ctx, event)
	}
}

// Query delegates to the inner repository synchronously.
func (r *AsyncAuditRepository) Query(ctx context.Context, filter entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	return r.inner.Query(ctx, filter)
}

func (r *AsyncAuditRepository) processLoop() {
	defer r.wg.Done()

	for {
		select {
		case event := <-r.eventCh:
			if err := r.inner.LogEvent(context.Background(), event); err != nil {
				log.Printf("async audit log error: %v", err)
			}
		case <-r.stopCh:
			// Drain remaining events
			for {
				select {
				case event := <-r.eventCh:
					if err := r.inner.LogEvent(context.Background(), event); err != nil {
						log.Printf("async audit log drain error: %v", err)
					}
				default:
					return
				}
			}
		}
	}
}
