package common

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractProjectIdFromUrl(t *testing.T) {
	t.Run("http", func(t *testing.T) {
		t.Run("invalid", func(t *testing.T) {
			assert.Equal(t, "", ExtractProjectIdFromUrl(
				"ftp://mygitlab.acme.tld/my-org/my-group/my-proj.git"))
			assert.Equal(t, "", ExtractProjectIdFromUrl("http://localhost"))
			assert.Equal(t, "", ExtractProjectIdFromUrl("localhost"))
		})
		t.Run("OK", func(t *testing.T) {
			assert.Equal(t, "my-org/my-group/my-proj", ExtractProjectIdFromUrl(
				"https://mygitlab.acme.tld/my-org/my-group/my-proj.git"))
			assert.Equal(t, "my-org/my-group/my-proj", ExtractProjectIdFromUrl(
				"https://mygitlab.acme.tld/my-org/my-group/my-proj"))
		})
	})
	t.Run("git", func(t *testing.T) {
		t.Run("invalid", func(t *testing.T) {
			assert.Equal(t, "", ExtractProjectIdFromUrl(
				"git@mygitlab.acme.tld/my-org/my-group/my-proj.git"))
		})
		t.Run("ok", func(t *testing.T) {
			assert.Equal(t, "my-org/my-group/my-proj", ExtractProjectIdFromUrl(
				"git@mygitlab.acme.tld:my-org/my-group/my-proj.git"))
			assert.Equal(t, "my-org/my-group/my-proj", ExtractProjectIdFromUrl(
				"git@mygitlab.acme.tld:my-org/my-group/my-proj"))
		})
	})
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
