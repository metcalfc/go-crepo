.PHONY: all test clean-dir clean help

all: help

crepo: crepo.go ## Build crepo
	go build -o crepo crepo.go

test: clean-dir crepo ## Run tests
	./crepo validate
	./crepo init
	./crepo foreach git status
	./crepo check
	touch test/crepo/dirty
	./crepo check -v || true # should fail

clean-dir: ## Clean up test directory
	rm -rf test

clean: clean-dir ## Clean up test directory and binary
	rm -f crepo

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
