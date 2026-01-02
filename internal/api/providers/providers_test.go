package providers_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/homedepot/arcade/internal/api"
	"github.com/homedepot/arcade/internal/api/providers"
	"github.com/stretchr/testify/assert"
)

func TestLoadTokenizersFromDir(t *testing.T) {
	t.Run("directory does not exist", func(t *testing.T) {
		_, err := providers.LoadTokenizersFromDir("i-dont-exist", time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "i-dont-exist")
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("no files exist", func(t *testing.T) {
		dir := t.TempDir()

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no token providers found in directory")
	})

	t.Run("file exists with bad json", func(t *testing.T) {
		dir := t.TempDir()

		path := filepath.Join(dir, "cred.json")
		assert.NoError(t, os.WriteFile(path, []byte("{"), 0o600))

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected end of JSON input")
	})

	t.Run("file exists without specifying a name", func(t *testing.T) {
		dir := t.TempDir()

		path := filepath.Join(dir, "provider.json")
		assert.NoError(t, os.WriteFile(path, []byte(`{}`), 0o600))

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `no "name" found`)
	})

	t.Run("duplicate credential exists (case-insensitive)", func(t *testing.T) {
		dir := t.TempDir()

		assert.NoError(t, os.WriteFile(filepath.Join(dir, "p1.json"), []byte(`{"type":"google","name":"google-test"}`), 0o600))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "p2.json"), []byte(`{"type":"google","name":"Google-Test"}`), 0o600))

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate token provider")
		assert.Contains(t, strings.ToLower(err.Error()), "google-test")
	})

	t.Run("microsoft missing clientId", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, `{
			"type": "microsoft",
			"name": "test",
			"clientSecret": "clientSecret",
			"resource": "resource",
			"loginEndpoint": "loginEndpoint"
		}`)

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "microsoft")
		assert.Contains(t, err.Error(), "clientId")
	})

	t.Run("microsoft missing clientSecret", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, `{
			"type": "microsoft",
			"name": "test",
			"clientId": "clientId",
			"resource": "resource",
			"loginEndpoint": "loginEndpoint"
		}`)

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "microsoft")
		assert.Contains(t, err.Error(), "clientSecret")
	})

	t.Run("microsoft missing resource", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, `{
			"type": "microsoft",
			"name": "test",
			"clientId": "clientId",
			"clientSecret": "clientSecret",
			"loginEndpoint": "loginEndpoint"
		}`)

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "microsoft")
		assert.Contains(t, err.Error(), "resource")
	})

	t.Run("microsoft missing loginEndpoint", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, `{
			"type": "microsoft",
			"name": "test",
			"clientId": "clientId",
			"clientSecret": "clientSecret",
			"resource": "resource"
		}`)

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "microsoft")
		assert.Contains(t, err.Error(), "loginEndpoint")
	})

	t.Run("rancher missing username", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, `{
			"type": "rancher",
			"name": "test",
			"password": "password",
			"url": "url"
		}`)

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "rancher")
		assert.Contains(t, err.Error(), "username")
	})

	t.Run("rancher missing password", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, `{
			"type": "rancher",
			"name": "test",
			"username": "username",
			"url": "url"
		}`)

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "rancher")
		assert.Contains(t, err.Error(), "password")
	})

	t.Run("rancher missing url", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, `{
			"type": "rancher",
			"name": "test",
			"username": "username",
			"password": "password"
		}`)

		_, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "rancher")
		assert.Contains(t, err.Error(), "url")
	})

	t.Run("it succeeds", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, `{"type":"google","name":"google-test"}`)

		tokenizers, err := providers.LoadTokenizersFromDir(dir, time.Second*api.DefaultTimeoutSeconds)
		assert.NoError(t, err)
		assert.Len(t, tokenizers, 1)
		_, ok := tokenizers["google-test"]
		assert.True(t, ok)
	})
}

func write(t *testing.T, dir, json string) {
	t.Helper()
	path := filepath.Join(dir, "provider.json")
	assert.NoError(t, os.WriteFile(path, []byte(json), 0o600))
}
