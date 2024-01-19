GO_PATH := $(shell go env GOPATH)

.PHONY : all
all : dep lint

lint: check-lint dep
	golangci-lint run --timeout=5m -c .golangci.yml

check-lint:
	@which golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_PATH)/bin v1.55.2

dep:
	@go mod tidy
	@go mod download
	@go mod vendor


define increment
	$(eval v := $(shell git describe --tags --abbrev=0 | sed -Ee 's/^v|-.*//'))
	$(eval n := $(shell echo $(v) | awk -F. -v OFS=. -v f=$1 '{ $$f++ } 1'))
	@git tag -a v$(n) -m "Bumped to version $(n), $(m)"
	@git push --set-upstream origin $(git branch --show-current)
	@git push --tags
	@echo "Updating version $(v) to $(n)"
endef

define minor
	$(eval v := $(shell git describe --tags --abbrev=0 | sed -Ee 's/^v|-.*//'))
	$(eval n := $(shell echo $(v) | awk -F. -v OFS=. -v f=$1 '{ $$f++ } 1' | awk -F. '{print $$1 "." $$2 "." 0 }'))

	@git tag -a v$(n) -m "Bumped to version $(n)"
	@git push --set-upstream origin $(@git branch --show-current) --tags
	@echo "Updating version $(v) to $(n)"
endef

version:
	@git describe --tags --abbrev=0

git:
	@git add .
	@git commit -m "$m"
	@git push -u origin ${BRANCH}

release: dep build git patch

major:
	$(call increment,1,major)
minor:
	$(call increment,2,minor)
patch:
	$(call increment,3,path)
new-minor:
	$(call minor,2,minor)
