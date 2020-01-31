package blob

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

var _ crud.Store = &Store{}

// Store implements the backing store for claims in azure blob storage
type Store struct {
	logger hclog.Logger
	config azureconfig.Config

	Container string
	CredentialSet
}

func NewStore(cfg azureconfig.Config, l hclog.Logger) *Store {
	return &Store{
		config: cfg,
		logger: l,
	}
}

func (s *Store) init() error {
	// TODO: should we allow the container to be configurable?
	s.Container = "porter"

	creds, err := GetCredentials(s.config, s.logger)
	if err != nil {
		return err
	}
	s.CredentialSet = creds

	return nil
}

func (s *Store) List(itemType string) ([]string, error) {
	err := s.init()
	if err != nil {
		return nil, err
	}

	container, err := s.buildContainerURL()
	if err != nil {
		return nil, err
	}

	var claims []string
	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := container.ListBlobsFlatSegment(context.Background(), marker,
			azblob.ListBlobsSegmentOptions{Prefix: itemType}) // Filter by item type
		if err != nil {
			return nil, err
		}

		marker = listBlob.NextMarker

		for _, blobInfo := range listBlob.Segment.BlobItems {
			claimName := strings.TrimPrefix(blobInfo.Name, itemType+"/")
			claims = append(claims, claimName)
		}
	}

	return claims, nil
}

func (s *Store) Save(itemType string, name string, data []byte) error {
	err := s.init()
	if err != nil {
		return err
	}

	blob, err := s.buildBlockBlobURL(itemType, name)
	if err != nil {
		return err
	}
	opts := azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16}

	_, err = azblob.UploadBufferToBlockBlob(context.Background(), data, blob, opts)
	return err
}

func (s *Store) Read(itemType string, name string) ([]byte, error) {
	err := s.init()
	if err != nil {
		return nil, err
	}

	return s.getBlob(itemType, name)
}

func (s *Store) Delete(itemType string, name string) error {
	err := s.init()
	if err != nil {
		return err
	}

	blob, err := s.buildBlockBlobURL(itemType, name)
	if err != nil {
		return err
	}

	_, err = blob.Delete(context.Background(), azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})
	return err
}

func (s *Store) getBlob(itemType string, blobName string) ([]byte, error) {
	blobURL, err := s.buildBlobURL(itemType, blobName)
	if err != nil {
		return nil, err
	}

	resp, err := blobURL.Download(context.Background(), 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return nil, err
	}

	bodyStream := resp.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})
	buff := bytes.Buffer{}
	_, err = buff.ReadFrom(bodyStream)

	return buff.Bytes(), err
}

func (s *Store) buildContainerURL() (azblob.ContainerURL, error) {
	rawURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", s.Credential.AccountName(), s.Container)
	URL, err := url.Parse(rawURL)
	if err != nil {
		return azblob.ContainerURL{}, errors.Wrapf(err, "could not parse container URL %s", rawURL)
	}

	return azblob.NewContainerURL(*URL, s.Pipeline), nil
}

func (s *Store) buildBlobURL(itemType string, blobName string) (azblob.BlobURL, error) {
	containerURL, err := s.buildContainerURL()
	if err != nil {
		return azblob.BlobURL{}, err
	}

	url := containerURL.NewBlobURL(path.Join(itemType, blobName))
	return url, nil
}

func (s *Store) buildBlockBlobURL(itemType string, blobName string) (azblob.BlockBlobURL, error) {
	containerURL, err := s.buildContainerURL()
	if err != nil {
		return azblob.BlockBlobURL{}, err
	}

	url := containerURL.NewBlockBlobURL(path.Join(itemType, blobName))
	return url, nil
}
