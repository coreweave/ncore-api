package s3

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPresigner_GetObject(t *testing.T) {

	os.Setenv("AWS_ACCESS_KEY_ID", "my-test-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "my-secret-value")
	os.Setenv("AWS_REGION", "default")
	presigner := NewPresigner("https://object.ord1.coreweave.com")
	request, err := presigner.GetObject("ncore-images", "img-2202.iso", 900)

	assert.NoError(t, err)
	t.Logf("request: %s", request.URL)
}
