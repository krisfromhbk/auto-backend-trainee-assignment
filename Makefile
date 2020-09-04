SOURCE_FILES?=./...

export GO111MODULE := on
#GO := GO111MODULE=on go

FAILPOINT_ENABLE  := $$(find $$PWD/ -type d | grep -vE "(\.git|tools)" | xargs tools/bin/failpoint-ctl enable)
FAILPOINT_DISABLE := $$(find $$PWD/ -type d | grep -vE "(\.git|tools)" | xargs tools/bin/failpoint-ctl disable)

tools/bin/failpoint-ctl: go.mod
	go build -o $@ github.com/pingcap/failpoint/failpoint-ctl

failpoint-enable: tools/bin/failpoint-ctl
# Converting gofail failpoints...
	@$(FAILPOINT_ENABLE)

failpoint-disable: tools/bin/failpoint-ctl
# Restoring gofail failpoints...
	@$(FAILPOINT_DISABLE)

build:
	go build cmd/server/main.go
.PHONY: build

test:
	go test -v -failfast -coverprofile=coverage.txt -covermode=atomic $(SOURCE_FILES) -timeout=2m
.PHONY: test

cover: test
	go tool cover -html=coverage.txt
.PHONY: cover