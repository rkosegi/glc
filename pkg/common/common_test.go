package common

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractProjectIdFromUrl(t *testing.T) {
	assert.Equal(t, "my-org%2Fmy-group%2Fmy-proj", ExtractProjectIdFromUrl(
		"https://mygitlab.acme.tld/my-org/my-group/my-proj.git"))
	assert.Equal(t, "my-org%2Fmy-group%2Fmy-proj", ExtractProjectIdFromUrl(
		"https://mygitlab.acme.tld/my-org/my-group/my-proj"))

	assert.Equal(t, "my-org%2Fmy-group%2Fmy-proj", ExtractProjectIdFromUrl(
		"git@gitlab.acme.tld:dnation/my-org/my-group/my-proj.git"))
}

func TestFindGitRemoteUrl(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		d := filepath.Join(t.TempDir(), ".git")
		assert.NoError(t, os.MkdirAll(d, 0750))
		cmd := exec.Command("git", "init", d)
		assert.NoError(t, cmd.Run())
		cmd = exec.Command("git", "remote", "add", "origin", "git@mygitlab.acme.tld:rkosegi/something.git")
		cmd.Dir = d
		assert.NoError(t, cmd.Run())
		u, err := FindGitRemoteUrl(d)
		assert.NoError(t, err)
		assert.Equal(t, "git@mygitlab.acme.tld:rkosegi/something.git", u)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := FindGitRemoteUrl(t.TempDir())
		assert.Error(t, err)
	})
}

func TestCompressText(t *testing.T) {
	in := "Aptent quis, tellus eu ante malesuada"
	assert.Equal(t, "Aptent…suada", CompressText(in, 10))
	assert.Equal(t, "Aptent quis…suada", CompressText(in, 15))
	assert.Equal(t, "Aptent quis, tel…suada", CompressText(in, 20))
	assert.Equal(t, "Aptent quis, tellus e…suada", CompressText(in, 25))
	assert.Equal(t, in, CompressText(in, len(in)))
}
