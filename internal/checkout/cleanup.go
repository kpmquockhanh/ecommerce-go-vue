package checkout

import (
	"context"
	"log"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"

	"ecommerce-api-go/internal/repositories"
)

type CleanupJob struct {
	sessionRepo repositories.CheckoutSessionRepository
	interval    time.Duration
	ttl         time.Duration
}

func NewCleanupJob(sessionRepo repositories.CheckoutSessionRepository, interval, ttl time.Duration) *CleanupJob {
	return &CleanupJob{
		sessionRepo: sessionRepo,
		interval:    interval,
		ttl:         ttl,
	}
}

func (j *CleanupJob) Start(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	log.Printf("Checkout cleanup job started (interval=%v, ttl=%v)", j.interval, j.ttl)

	for {
		select {
		case <-ctx.Done():
			log.Println("Checkout cleanup job stopped")
			return
		case <-ticker.C:
			j.run(ctx)
		}
	}
}

func (j *CleanupJob) run(ctx context.Context) {
	sessions, err := j.sessionRepo.FindAbandoned(ctx, j.ttl, 100)
	if err != nil {
		log.Printf("CleanupJob: failed to find abandoned sessions: %v", err)
		return
	}

	if len(sessions) == 0 {
		return
	}

	log.Printf("CleanupJob: found %d abandoned checkout sessions", len(sessions))

	for _, session := range sessions {
		if session.PaymentIntentID == "" {
			continue
		}

		pi, err := paymentintent.Get(session.PaymentIntentID, nil)
		if err != nil {
			log.Printf("CleanupJob: failed to get PI %s: %v", session.PaymentIntentID, err)
			continue
		}

		if pi.Status == stripe.PaymentIntentStatusRequiresPaymentMethod ||
			pi.Status == stripe.PaymentIntentStatusRequiresConfirmation {
			_, cancelErr := paymentintent.Cancel(session.PaymentIntentID, nil)
			if cancelErr != nil {
				log.Printf("CleanupJob: failed to cancel PI %s: %v", session.PaymentIntentID, cancelErr)
				continue
			}
			log.Printf("CleanupJob: cancelled orphaned PI %s for user %d", session.PaymentIntentID, session.UserID)
		}

		if err := j.sessionRepo.UpdateStatus(ctx, session.IdempotencyKey, session.UserID, "expired"); err != nil {
			log.Printf("CleanupJob: failed to expire session %s: %v", session.IdempotencyKey, err)
		}
	}

	expired, err := j.sessionRepo.ExpireOldSessions(ctx, j.ttl)
	if err != nil {
		log.Printf("CleanupJob: failed to expire old sessions: %v", err)
		return
	}
	if expired > 0 {
		log.Printf("CleanupJob: expired %d old checkout sessions", expired)
	}
}
