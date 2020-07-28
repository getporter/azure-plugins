package blob

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Add tags to each claim so that when the migration runs,
// it can use the new code only which operates exclusively on tags:
// type = claims

func (s *Store) migrateToTaggedClaims() error {
	s.logger.Warn("Tagging claims data in preparation for claims migration")

	// List all claims
	blobs, err := s.listBlobs("claims/")
	if err != nil {
		return errors.Wrap(err, "Unable to list claims for migration to the new azure plugin storage format")
	}

	tags := map[string]string{
		"type": "claims",
	}

	g := new(errgroup.Group)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, blob := range blobs {
		g.Go(func() error {
			return s.setTag(ctx, blob.Name, tags)
		})
	}

	return g.Wait()
}
