package table

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccountName = "testaccountname"
	testAccountKey  = "dGVzdEFjY291bnRLZXkK"
)

func Test_NoClient(t *testing.T) {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   t.Name(),
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	config := azureconfig.Config{}

	conn := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if conn != "" {
		os.Setenv("AZURE_STORAGE_CONNECTION_STRING", "")
		defer os.Setenv("AZURE_STORAGE_CONNECTION_STRING", conn)
	}

	store := NewStore(config, logger, nil)
	requiredError := "environment variable AZURE_STORAGE_CONNECTION_STRING containing the azure storage connection string was not set:\nazureconfig.Config{EnvConnectionString:\"\", StorageAccount:\"\", StorageAccountResourceGroup:\"\", StorageAccountSubscriptionId:\"\", StorageCompressData:false, EnvAzurePrefix:\"\", Vault:\"\"}"

	_, err := store.Count("test", "test")
	require.Error(t, err, "store.Count should have returned an error")
	assert.EqualError(t, err, requiredError)

	_, err = store.List("test", "test")
	require.Error(t, err, "store.List should have returned an error")
	assert.EqualError(t, err, requiredError)

	_, err = store.Read("test", "test")
	require.Error(t, err, "store.Read should have returned an error")
	assert.EqualError(t, err, requiredError)

	err = store.Delete("test", "test")
	require.Error(t, err, "store.Delete should have returned an error")
	assert.EqualError(t, err, requiredError)

	err = store.Save("test", "test", "test", []byte{123})
	require.Error(t, err, "store.Save should have returned an error")
	assert.EqualError(t, err, requiredError)

}

func Test_Save(t *testing.T) {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   t.Name(),
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	config := azureconfig.Config{}

	conn := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if conn != "" {
		os.Setenv("AZURE_STORAGE_CONNECTION_STRING", "")
		defer os.Setenv("AZURE_STORAGE_CONNECTION_STRING", conn)
	}

	store := NewStore(config, logger, nil)

	data, err := ioutil.ReadFile("testdata/test.dat")
	require.NoError(t, err, "Reading test data should not cause an error")
	require.True(t, len(data) > 65536, "Test data should be larger than 65536 bytes")

	requiredError := fmt.Sprintf("Data exceeds maximum length for table storage for item: test/ group=\"test\" test length: %d", len(data))
	err = store.Save("test", "test", "test", data)
	require.Error(t, err, "store.Save should have returned an error")
	assert.EqualError(t, err, requiredError)

}

func Test_Schema(t *testing.T) {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   t.Name(),
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	config := azureconfig.Config{}
	table, rec := getTableReferenceAndRecorder(t, t.Name())
	defer rec.Stop()
	store := NewStore(config, logger, table)
	store.table = table

	data, err := store.Read("", "schema")
	if err == nil {
		err = store.Delete("schema", "schema")
		require.NoError(t, err, "Deleting schema should not cause an error")
		data, err = store.Read("", "schema")
	}

	require.Error(t, err, "Reading non-existant schema should cause an error")
	assert.EqualError(t, err, "File does not exist")
	assert.Equal(t, 0, len(data))

	err = store.Save("", "", "schema", []byte("schema data"))
	require.NoError(t, err, "Saving schema should not cause an error")

	data, err = store.Read("", "schema")
	require.NoError(t, err, "Reading schema should not cause an error")
	assert.NotEqual(t, 0, len(data))
	assert.Equal(t, string(data), "schema data")

}

func Test_NoGroup_NoData(t *testing.T) {

	testcases := []string{
		"installation1",
		"installation2",
		"installation3",
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   t.Name(),
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	config := azureconfig.Config{}
	table, rec := getTableReferenceAndRecorder(t, t.Name())
	defer rec.Stop()
	store := NewStore(config, logger, table)
	store.table = table

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			err := store.Save("installations", "", tc, nil)
			require.NoError(t, err, "Saving installation %s should not result in an error", tc)
		})
	}

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			data, err := store.Read("installations", tc)
			require.NoError(t, err, "Reading installation %s should not result in an error", tc)
			assert.Equal(t, []byte(nil), data)
		})
	}

	data, err := store.List("installations", "")
	require.NoError(t, err, "Listing installations should not result in an error")
	assert.Equal(t, 3, len(data))

	count, err := store.Count("installations", "")
	require.NoError(t, err, "Counting installations should not result in an error")
	assert.Equal(t, 3, count)

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			err := store.Delete("installations", tc)
			require.NoError(t, err, "Deleting installation %s should not result in an error", tc)
		})
	}
}
func Test_With_Group_And_Data(t *testing.T) {

	testcases := map[string][]struct {
		name string
		data []byte
	}{
		"test1": {
			{
				"claim1",
				[]byte("claim1"),
			},
			{
				"claim2",
				[]byte("claim2"),
			},
			{
				"claim3",
				[]byte("claim3"),
			},
		},
		"test2": {
			{
				"claim4",
				[]byte("claim4"),
			},
			{
				"claim5",
				[]byte("claim5"),
			},
			{
				"claim6",
				[]byte("claim6"),
			},
			{
				"claim7",
				[]byte("claim7"),
			},
			{
				"claim8",
				[]byte("claim8"),
			},
		},
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   t.Name(),
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	config := azureconfig.Config{}
	table, rec := getTableReferenceAndRecorder(t, t.Name())
	defer rec.Stop()
	store := NewStore(config, logger, table)
	store.table = table

	for group, tests := range testcases {
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				err := store.Save("claims", group, tc.name, tc.data)
				require.NoError(t, err, "Saving claim %s in group %s should not result in an error", tc.name, group)
			})
		}
	}

	for group, tests := range testcases {
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				data, err := store.Read("claims", tc.name)
				require.NoError(t, err, "Reading claim %s in group %s should not result in an error", tc.name, group)
				assert.Equal(t, tc.data, data)
			})
		}
	}

	for group, tests := range testcases {
		data, err := store.List("claims", group)
		require.NoError(t, err, "Listing claims for group %s should not result in an error", group)
		assert.Equal(t, len(tests), len(data))

		count, err := store.Count("claims", group)
		require.NoError(t, err, "Counting claims for group %s should not result in an error", group)
		assert.Equal(t, len(tests), count)
	}

	for group, tests := range testcases {
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				err := store.Delete("claims", tc.name)
				require.NoError(t, err, "Deleting claim %s in group %s should not result in an error", tc.name, group)
			})
		}
	}

}

func Test_With_Compressed_Data(t *testing.T) {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   t.Name(),
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	config := azureconfig.Config{
		StorageCompressData: true,
	}
	table, rec := getTableReferenceAndRecorder(t, t.Name())
	defer rec.Stop()
	store := NewStore(config, logger, table)
	store.table = table

	data, err := ioutil.ReadFile("testdata/test.dat")
	require.NoError(t, err, "Reading test data should not cause an error")
	require.True(t, len(data) > 65536, "Test data should be larger than 65536 bytes")
	err = store.Save("claims", "largedata", t.Name(), data)
	require.NoError(t, err, "Saving compressed date > 65536 should succeed")

}

func getTableReferenceAndRecorder(t *testing.T, cassetteName string) (*storage.Table, *recorder.Recorder) {
	testMode := os.Getenv("RECORDER_MODE")
	mode := recorder.ModeReplaying
	if strings.EqualFold(testMode, "record") {
		mode = recorder.ModeRecording
	}
	connectionString := fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=core.windows.net", testAccountName, testAccountKey)
	if mode == recorder.ModeRecording {
		connectionString = os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
		if connectionString == "" {
			t.Fatal("Test Recording Mode requires a valid azure connection string in environment variable AZURE_STORAGE_CONNECTION_STRING")
		}
		createTableIfNotExists(t, connectionString)
	}
	rec, err := recorder.NewAsMode(fmt.Sprintf("testdata/%s", cassetteName), mode, nil)
	assert.NoError(t, err, "Expected no Error when creating recorder")
	rec.AddFilter(func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	})
	rec.SetMatcher(func(r *http.Request, i cassette.Request) bool {
		return compareMethods(r, i) &&
			compareURLs(r, i) &&
			compareBodies(r, i)
	})
	client, err := storage.NewClientFromConnectionString(connectionString)
	assert.NoError(t, err, "Expected no Error when creating storage client")
	client.HTTPClient = &http.Client{Transport: rec}
	tableServiceClient := client.GetTableService()
	return tableServiceClient.GetTableReference(tableName), rec
}

func createTableIfNotExists(t *testing.T, connectionString string) {
	client, err := storage.NewClientFromConnectionString(connectionString)
	assert.NoError(t, err, "Expected no Error when creating storage client")
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   t.Name(),
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	config := azureconfig.Config{}
	tableServiceClient := client.GetTableService()
	store := NewStore(config, logger, tableServiceClient.GetTableReference(tableName))
	err = store.CreateTableIfNotExists()
	assert.NoError(t, err, "Expected no Error when checking if table exists")

}

func compareMethods(r *http.Request, i cassette.Request) bool {
	return strings.EqualFold(r.Method, i.Method)
}

func compareURLs(r *http.Request, i cassette.Request) bool {
	cassetteURL, err := url.Parse(i.URL)
	if err != nil {
		return false
	}
	// This does not support recording tests using the storage emulator on windows (format of the url is http://127.0.0.1:<port-number-for-service>/<account-name>/<resource-path> )
	cassetteURL.Host = fmt.Sprintf("%s.table.core.windows.net", testAccountName)
	return r.URL.String() == cassetteURL.String()
}

func compareBodies(r *http.Request, i cassette.Request) bool {
	body := bytes.Buffer{}
	if r.Body != nil {
		if _, err := body.ReadFrom(r.Body); err != nil {
			return false
		}
		r.Body = ioutil.NopCloser(&body)
	}

	return body.String() == i.Body
}
