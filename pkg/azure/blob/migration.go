package blob

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Add tags to each claim so that when the migration runs,
// it can use the new code only which operates exclusively on tags:
// type = claims

func (s *Store) migrateToTaggedClaims() error {
	// List all claims
	blobs, err := s.listBlobs("claims/")
	if err != nil {
		return errors.Wrap(err, "unable to list claims for migration to the new azure plugin storage format")
	}

	container, err := s.buildContainerURL()
	if err != nil {
		return err
	}

	var bigErr *multierror.Error
	for _, blob := range blobs {
		tags := map[string]string{
			"type": "claims",
		}

		blobURL := container.NewBlobURL(blob.Name)
		_, err = blobURL.SetTags(context.Background(), tags)
		if err != nil {
			err = multierror.Append(bigErr, err)
		}
	}

	return bigErr.ErrorOrNil()
}
