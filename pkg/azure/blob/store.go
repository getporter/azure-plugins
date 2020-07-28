package blob

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"text/template"
	"time"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
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

func (s *Store) getTags(itemType string, group string) map[string]string {
	tags := map[string]string{}
	if itemType != "" {
		tags["type"] = itemType
	}

	if group != "" {
		switch itemType {
		case claim.ItemTypeClaims:
			tags["installation"] = group
		case claim.ItemTypeResults:
			tags["claim-id"] = group
		case claim.ItemTypeOutputs:
			tags["result-id"] = group
		}
	}

	return tags
}

func (s *Store) List(itemType string, group string) ([]string, error) {
	err := s.init()
	if err != nil {
		return nil, err
	}

	items, err := s.filterBlobsByTags(s.getTags(itemType, group))
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(items))
	for _, blobInfo := range items {
		// Return just the name portion of the blob, e.g. installations/INSTALLATION -> INSTALLATION
		fileName := path.Base(blobInfo.Name)
		names = append(names, fileName)
	}

	s.logger.Info(fmt.Sprintf("blob names: %s", strings.Join(names, ", ")))
	return names, nil
}

func (s *Store) filterBlobsByTags(tags map[string]string) ([]azblob.BlobItem, error) {
	service, err := s.buildServiceURL()
	if err != nil {
		return nil, err
	}

	// Search for blobs tagged with the specified group, e.g. installation=$GROUP
	// this gives us the claims in the installation
	tmplData := struct {
		Tags      map[string]string
		Container string
	}{
		Tags:      tags,
		Container: s.Container,
	}
	tmpl := `{{range $k, $v := .Tags}}"{{$k}}"='{{$v}}' AND {{end}}@container='{{.Container}}'`
	t, err := template.New("where").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	err = t.Execute(&buf, tmplData)
	if err != nil {
		return nil, err
	}

	opts := azblob.FilterBlobsByTagsOptions{
		Where: buf.String(),
	}
	s.logger.Info(fmt.Sprintf("where %s", opts.Where))
	result, err := service.FilterBlobsByTags(context.Background(), azblob.Marker{}, opts)
	if err != nil {
		s.logger.Error(errors.Wrapf(err, "error filtering blobs where %s", opts.Where).Error())
		return nil, err
	}

	s.logger.Info(fmt.Sprintf("matched %d blobs", len(result.Segment.BlobItems)))
	return result.Segment.BlobItems, nil
}

func (s *Store) listBlobs(prefix string) ([]azblob.BlobItem, error) {
	container, err := s.buildContainerURL()
	if err != nil {
		return nil, err
	}

	s.logger.Info(fmt.Sprintf("list %s/", prefix))
	opts := azblob.ListBlobsSegmentOptions{
		Prefix: prefix,
	}

	result, err := container.ListBlobsFlatSegment(context.Background(), azblob.Marker{}, opts)
	if err != nil {
		s.logger.Error(errors.Wrapf(err, "error listing blobs by prefix %s", prefix).Error())
		return nil, err
	}
	return result.Segment.BlobItems, nil
}

func (s *Store) Save(itemType string, group string, name string, data []byte) error {
	err := s.init()
	if err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("Save %s/ group=%q %s", itemType, group, name))
	tags := s.getTags(itemType, group)
	return s.saveBlob(s.buildRelBlobURL(itemType, name), tags, data)
}

func (s *Store) saveBlob(path string, tags map[string]string, data []byte) error {
	blob, err := s.buildBlockBlobURL(path)
	if err != nil {
		return err
	}
	opts := azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16}

	_, err = azblob.UploadBufferToBlockBlob(context.Background(), data, blob, opts)
	if err != nil {
		return err
	}

	if tags != nil && len(tags) > 0 {
		err = s.setTag(context.Background(), path, tags)
	}

	return err
}

func (s *Store) Read(itemType string, name string) ([]byte, error) {
	err := s.init()
	if err != nil {
		return nil, err
	}

	s.logger.Info(fmt.Sprintf("Read %s/ %s", itemType, name))
	data, err := s.getBlob(itemType, name)

	// Check if a migration should be attempted
	if itemType == "" && name == "schema" && err == crud.ErrRecordDoesNotExist {
		s.migrateToTaggedClaims()
	}

	return data, err
}

func (s *Store) Delete(itemType string, name string) error {
	err := s.init()
	if err != nil {
		return err
	}

	return s.deleteBlob(s.buildRelBlobURL(itemType, name))
}

// deleteBlob, allowing for it to already be gone
// Delete any associated tags at the same time
func (s *Store) deleteBlob(path string) error {
	blob, err := s.buildBlockBlobURL(path)
	if err != nil {
		return err
	}

	tagResp, err := blob.SetTags(context.Background(), nil)
	if err != nil {
		if tagResp.StatusCode() != http.StatusNotFound {
			return err
		}
	}
	blobResp, err := blob.Delete(context.Background(), azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})
	if err != nil {
		if blobResp.StatusCode() != http.StatusNotFound {
			return err
		}
	}
	return nil
}

func (s *Store) getBlob(itemType string, blobName string) ([]byte, error) {
	blobURL, err := s.buildBlobURL(itemType, blobName)
	if err != nil {
		return nil, err
	}

	resp, err := blobURL.Download(context.Background(), 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	if err != nil {
		if sErr, ok := err.(azblob.StorageError); ok && sErr.ServiceCode() == azblob.ServiceCodeBlobNotFound {
			return nil, crud.ErrRecordDoesNotExist
		}
		return nil, err
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, crud.ErrRecordDoesNotExist
	}

	bodyStream := resp.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})
	buff := bytes.Buffer{}
	_, err = buff.ReadFrom(bodyStream)

	return buff.Bytes(), err
}

// setTag on a blob and wait for the tag index to sync.
func (s *Store) setTag(ctx context.Context, blobName string, tags map[string]string) error {
	g := new(errgroup.Group)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	container, err := s.buildContainerURL()
	if err != nil {
		return err
	}

	blobURL := container.NewBlobURL(blobName)

	s.logger.Info(fmt.Sprintf("Setting tags on %s: %v", blobName, tags))
	g.Go(func() error {
		_, err = blobURL.SetTags(ctx, tags)
		return err
	})

	g.Go(func() error {
		return s.waitForTagInCache(ctx, blobName, tags)
	})

	return g.Wait()
}

// waitForMigratedTagCacheWarm waits until a query for the specified tags returns at least the
// specified number of blobs, indicating that the tag cache has caught up with the tags applied
// during the migration. Not sure why this is necessary, but it seems to avoid a race condition
// and nothing ends up being migrated.
func (s *Store) waitForTagInCache(parentCtx context.Context, blobName string, tags map[string]string) error {
	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
	defer cancel()
	s.logger.Info("Waiting for tags to sync...")
	for {
		select {
		case <-ctx.Done():
			return errors.New("Timed out waiting for the azure blob storage tags cache to warm up")
		default:
			cachedBlobs, _ := s.filterBlobsByTags(tags)
			for _, b := range cachedBlobs {
				if b.Name == blobName {
					s.logger.Info(fmt.Sprintf("Found %s, tag cache warm", blobName))
					return nil
				}
				time.Sleep(time.Second)
			}
		}
	}
}

func (s *Store) buildServiceURL() (azblob.ServiceURL, error) {
	rawURL := fmt.Sprintf("https://%s.blob.core.windows.net", s.Credential.AccountName())
	URL, err := url.Parse(rawURL)
	if err != nil {
		return azblob.ServiceURL{}, errors.Wrapf(err, "could not parse service URL %s", rawURL)
	}

	return azblob.NewServiceURL(*URL, s.Pipeline), nil
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

func (s *Store) buildRelBlobURL(itemType string, blobName string) string {
	return path.Join(itemType, blobName)
}

func (s *Store) buildBlockBlobURL(path string) (azblob.BlockBlobURL, error) {
	containerURL, err := s.buildContainerURL()
	if err != nil {
		return azblob.BlockBlobURL{}, err
	}

	url := containerURL.NewBlockBlobURL(path)
	return url, nil
}
