package common

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseAzureProfile(t *testing.T) {

	files := []string{"profile_with_bom.json", "profile_without_bom.json"}
	for _, filename := range files {
		testName := fmt.Sprintf("parsing %s", filename)
		t.Run(testName, func(t *testing.T) {
			testdata := path.Join("testdata", filename)
			subscriptionId, err := getCurrentAzureSubscriptionFromProfile(testdata)
			assert.NoError(t, err, "Expected no error parsing Azure Profile")
			assert.Equal(t, "8b5ab980-0253-40d6-b22a-61b3f9d94491", subscriptionId, "Expected Subscription not found parsing Azure Profile")
		})
	}
}
