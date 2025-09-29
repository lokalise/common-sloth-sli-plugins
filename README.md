Reference repository: <https://github.com/slok/sloth-common-sli-plugins>

## Development

```
# Install dependencies
$ go mod tidy

# Run a test for a specific plugin
$ go test plugins/nginx-http/availability/plugin_test.go
```

## Generate rules (for testing)

```
docker run -v $(pwd):/home/nonroot ghcr.io/slok/sloth:v0.11.0 generate -p ./plugins -i example-slo-spec.yaml -o example-prometheus-servicelevel.yaml
```

## Usage

Sloth runs with a `git-sync` sidecar which will automatically pick up the latest changes from this repo.
A webhook is fired by `git-sync` which causes Sloth to reload the plugins and regenerate the Prometheus rules.
