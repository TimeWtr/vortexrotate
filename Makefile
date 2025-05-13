.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: ut
ut:
	@go test -race ./...

.PHONY: clean
clean:
	@cd tests && rm *.gz || true  && rm *.zst || true && rm *.snappy || true

.PHONY: check
check:
	@$(MAKE) --no-print-directory tidy
	@$(MAKE) --no-print-directory ut
	@$(MAKE) --no-print-directory clean