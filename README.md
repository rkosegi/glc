# Yet another GitLab CLI

## Build

You need [Go runtime](https://go.dev/dl/) and [Goreleaser](https://github.com/goreleaser/goreleaser) in the path

```shell
make build
```

## Usage

This program depends on 2 environment variables:

- `GITLAB_TOKEN` - API token used for authentication. Can be overridden with flag `--gitlab-token-env`
- `CI_API_V4_URL` - API endpoint URL like `https://gitlab.ci.acme.tld/api/v4`. Can be overridden with `--gitlab-endpoint-env`.

### Revision difference

```shell
glc revdiff --from v0.0.1 --to v0.10.2 --project-id 123 --log-level debug --output diff.yaml
```
