package blob

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"net/url"

	"get.porter.sh/plugin/azure/pkg/azure/credentials"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

var _ crud.Store = &Store{}

// Store implements the backing store for claims in azure blob storage
type Store struct {
	logger    hclog.Logger
	Container string
	credentials.CredentialSet
}

func (s *Store) init() error {
	s.Container = "porter"

	creds, err := credentials.GetCredentials()
	if err != nil {
		return err
	}
	s.CredentialSet = creds

	return nil
}

func (s *Store) List() ([]string, error) {
	err := s.init()
	if err != nil {
		return nil, err
	}

	container, err := s.buildContainerURL(s.Container)
	if err != nil {
		return nil, err
	}

	var claims []string
	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := container.ListBlobsFlatSegment(context.Background(), marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return nil, err
		}

		marker = listBlob.NextMarker

		for _, blobInfo := range listBlob.Segment.BlobItems {
			claims = append(claims, blobInfo.Name)
		}
	}

	return claims, nil
}

func (s *Store) Store(name string, data []byte) error {
	err := s.init()
	if err != nil {
		return err
	}

	container, err := s.buildContainerURL(s.Container)
	if err != nil {
		return err
	}

	blob := container.NewBlockBlobURL(name)
	opts := azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16}

	_, err = azblob.UploadBufferToBlockBlob(context.Background(), data, blob, opts)
	return err
}

func (s *Store) Read(name string) ([]byte, error) {
	err := s.init()
	if err != nil {
		return nil, err
	}

	return s.getBlob(s.Container, name)
}

func (s *Store) Delete(name string) error {
	err := s.init()
	if err != nil {
		return err
	}

	container, err := s.buildContainerURL(s.Container)
	if err != nil {
		return err
	}

	container, err = s.buildContainerURL(s.Container)
	if err != nil {
		return err
	}

	blob := container.NewBlockBlobURL(name)

	_, err = blob.Delete(context.Background(), azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})
	return err
}

func (s *Store) getBlob(containerName string, blobName string) ([]byte, error) {
	containerURL, err := s.buildContainerURL(containerName)
	if err != nil {
		return nil, err
	}

	blobURL := containerURL.NewBlobURL(blobName)

	resp, err := blobURL.Download(context.Background(), 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return nil, err
	}

	bodyStream := resp.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})
	buff := bytes.Buffer{}
	_, err = buff.ReadFrom(bodyStream)

	return buff.Bytes(), err
}

func (s *Store) buildContainerURL(containerName string) (azblob.ContainerURL, error) {
	rawURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", s.Credential.AccountName(), containerName)
	URL, err := url.Parse(rawURL)
	if err != nil {
		return azblob.ContainerURL{}, errors.Wrapf(err, "could not parse container URL %s", rawURL)
	}

	return azblob.NewContainerURL(*URL, s.Pipeline), nil
}
