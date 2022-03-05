// Package repository implement interface for Repository
package repository

import (
	"context"

	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
)

// Repository interface for working with global repository
// go:generate mockery --name=Repository --inpackage
type Repository interface {
	// LinkByShort get original link from all storage
	LinkByShort(short shortlink.Short) (string, error)
	// Save link to repository
	Save(userID user.UniqUser, url string) (shortlink.Short, error)
	// BunchSave save mass urls and generate shorts
	BunchSave(userID user.UniqUser, urls []shortlink.URLs) ([]shortlink.ShortURLs, error)
	// LinksByUser return all user links
	LinksByUser(userID user.UniqUser) (shortlink.ShortLinks, error)
	// Clear storage
	Clear() error
	// BunchUpdateAsDeleted set flag as deleted
	BunchUpdateAsDeleted(ctx context.Context, ids []string, userID string) error
}
