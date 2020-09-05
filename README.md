[![build](https://github.com/krisfromhbk/auto-backend-trainee-assignment/workflows/build/badge.svg)](https://github.com/krisfromhbk/auto-backend-trainee-assignment/actions?query=workflow%3Abuild)
[![codecov](https://codecov.io/gh/krisfromhbk/auto-backend-trainee-assignment/branch/master/graph/badge.svg)](https://codecov.io/gh/krisfromhbk/auto-backend-trainee-assignment)

# Plans
- [x] Setup [dgraph-io/badger](https://github.com/dgraph-io/badger) as database.
- [x] Add structured logging with [uber-go/zap](https://github.com/uber-go/zap).
- [x] Implement built-in http server with two endpoints ("/api/shorten", "/**").
- [x] Rewrite built-in http server with [valyala/fasthttp](https://github.com/valyala/fasthttp).
- [x] Implement [failpoints](https://github.com/pingcap/failpoint) support for testing.
- [x] Add basic make targets to use failpoints.
- [ ] Add command-line flags parsing.
- [ ] Set URL validation.
- [ ] Add custom short links support.
- [ ] Setup [Github Actions](https://docs.github.com/en/actions) workflows.
- [ ] Prepare for load testing with [Yandex.Tank](https://github.com/yandex/yandex-tank)

# References
[Task definition](https://github.com/avito-tech/auto-backend-trainee-assignment)