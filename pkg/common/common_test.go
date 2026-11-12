package common

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func urlMust(u string) *url.URL {
	p, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return p
}

func TestExtractProjectIdFromUrl(t *testing.T) {
	assert.Equal(t, "my-org%2Fmy-group%2Fmy-proj", ExtractProjectIdFromUrl(
		urlMust("https://mygitlab.acme.tld/my-org/my-group/my-proj.git")))
	assert.Equal(t, "my-org%2Fmy-group%2Fmy-proj", ExtractProjectIdFromUrl(
		urlMust("https://mygitlab.acme.tld/my-org/my-group/my-proj")))
}
