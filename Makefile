.PHONY: setup
setup:
	@sh ./scripts/setup.sh

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: ut
ut:
	@go test -race -v ./...

.PHONY: lint
lint:
	@golangci-lint run -c ./scripts/lint/.golangci.yml ./...

.PHONY: clean
clean:
	@cd tests && ls | grep -v "test.log" | grep -v "test.reset.log" | xargs rm -rf
	@cd tests && rm -rf *.snappy && rm -rf *.zst && rm -rf *.gz

.PHONY: check
check:
	@true
#	@$(MAKE) --no-print-directory setup
#	@$(MAKE) --no-print-directory tidy
#	@$(MAKE) --no-print-directory ut
#	@$(MAKE) --no-print-directory clean