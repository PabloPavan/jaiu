package service

import (
	"context"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

func endSubscriptions(ctx context.Context, repo ports.SubscriptionRepository, subscriptions []domain.Subscription, now time.Time) error {
	if repo == nil {
		return nil
	}

	for _, subscription := range subscriptions {
		if subscription.Status == domain.SubscriptionEnded {
			continue
		}
		subscription.Status = domain.SubscriptionEnded
		subscription.UpdatedAt = now
		if _, err := repo.Update(ctx, subscription); err != nil {
			return err
		}
	}

	return nil
}
