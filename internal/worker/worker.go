package worker

import (
	"context"
	"log"

	"ecommerce-api-go/internal/mail"
	"ecommerce-api-go/internal/queue"
)

type Worker struct {
	queue   *queue.Queue
	mailer  *mail.Mailer
}

func New(q *queue.Queue, m *mail.Mailer) *Worker {
	return &Worker{
		queue:  q,
		mailer: m,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	log.Println("Starting worker...")

	// Start consumers for each queue
	errCh := make(chan error, 2)

	go func() {
		errCh <- w.queue.Consume(ctx, "emails", w.handleEmailJob)
	}()

	go func() {
		errCh <- w.queue.Consume(ctx, "jobs", w.handleGenericJob)
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		log.Println("Worker shutting down...")
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (w *Worker) handleEmailJob(job queue.Job) error {
	log.Printf("Processing email job: %s", job.Type)

	switch job.Type {
	case queue.JobOrderConfirmation:
		return w.handleOrderConfirmation(job)
	case queue.JobPaymentFailed:
		return w.handlePaymentFailed(job)
	case queue.JobOrderStatusChanged:
		return w.handleOrderStatusChanged(job)
	default:
		log.Printf("Unknown email job type: %s", job.Type)
		return nil
	}
}

func (w *Worker) handleGenericJob(job queue.Job) error {
	log.Printf("Processing generic job: %s", job.Type)
	return nil
}

func (w *Worker) handleOrderConfirmation(job queue.Job) error {
	var p queue.OrderConfirmationPayload
	if err := decodePayload(job.Payload, &p); err != nil {
		return err
	}

	return w.mailer.SendOrderConfirmation(
		p.Email,
		p.FirstName,
		p.LastName,
		p.OrderID,
		p.Total,
	)
}

func (w *Worker) handlePaymentFailed(job queue.Job) error {
	var p queue.PaymentFailedPayload
	if err := decodePayload(job.Payload, &p); err != nil {
		return err
	}

	return w.mailer.SendPaymentFailed(
		p.Email,
		p.FirstName,
		p.OrderID,
		p.Total,
		p.Reason,
	)
}

func (w *Worker) handleOrderStatusChanged(job queue.Job) error {
	var p queue.OrderStatusChangedPayload
	if err := decodePayload(job.Payload, &p); err != nil {
		return err
	}

	return w.mailer.SendOrderStatusUpdate(
		p.Email,
		p.FirstName,
		p.OrderID,
		p.OldStatus,
		p.NewStatus,
	)
}
