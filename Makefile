.PHONY: FORCE


build: go-build
.PHONY: build

lint: go-lint
.PHONY: lint

test: go-test
.PHONY: test

all-test: go-all-test
.PHONY: all-test


go-build: FORCE
	./build.sh

go-lint: FORCE
	./lint.sh

go-test: FORCE
	./test.sh report

go-all-test: FORCE
	./test.sh all
