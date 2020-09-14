[![build](https://github.com/krisfromhbk/auto-backend-trainee-assignment/workflows/build/badge.svg)](https://github.com/krisfromhbk/auto-backend-trainee-assignment/actions?query=workflow%3Abuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/krisfromhbk/auto-backend-trainee-assignment)](https://goreportcard.com/report/github.com/krisfromhbk/auto-backend-trainee-assignment)
[![codecov](https://codecov.io/gh/krisfromhbk/auto-backend-trainee-assignment/branch/master/graph/badge.svg)](https://codecov.io/gh/krisfromhbk/auto-backend-trainee-assignment)

# URL Shortener
Yet another service creating short urls. Made with love for Avito Auto team.

## Run locally using Docker
The Docker image for [url-shortener](https://hub.docker.com/repository/docker/krisfromhbk/avito-auto) is available on DockerHub.

1. Either `git clone` this repository locally and `cd auto-backend-trainee-assignment/deployments`, or download a copy of the [docker-compose.yaml](deployments/docker-compose.yml) locally.

1. Ensure you have the most up-to-date Docker container images:

   ```bash
   docker-compose pull
   ```

1. Run the service on your local Docker:

   ```bash
   docker-compose up
   ```
## API reference
Service container exposes its API on 9000 port. It can be changed with environment variable `PORT` specified in [docker-compose.yaml](deployments/docker-compose.yml)

### Create short url

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"url": "https://some.host/path"}' \
  http://localhost:9000/api/shorten
```

Response: 'short' - generated short url path (e.g. jnegYbw) or HTTP error code with description.

### Get redirect for short url

```bash
curl http://localhost:9000/jnegYbw
```

Response: HTTP 301 redirect with location header set to source url or HTTP error code with description..

## Plans
- [x] Setup [dgraph-io/badger](https://github.com/dgraph-io/badger) as database.
- [x] Add structured logging with [uber-go/zap](https://github.com/uber-go/zap).
- [x] Implement built-in http server with two endpoints ("/api/shorten", "/**").
- [x] Rewrite built-in http server with [valyala/fasthttp](https://github.com/valyala/fasthttp).
- [x] A graceful shutdown mechanism via a goroutine.
- [x] Implement [failpoints](https://github.com/pingcap/failpoint) support for testing.
- [x] Add basic make targets to use failpoints.
- [x] Add command-line flags parsing.
- [ ] Set URL validation.
- [ ] Add custom short links support.
- [x] Setup [Github Actions](https://docs.github.com/en/actions) workflows.
- [ ] Prepare for automated load testing with [Yandex.Tank](https://github.com/yandex/yandex-tank)

## References
[Task definition](https://github.com/avito-tech/auto-backend-trainee-assignment)