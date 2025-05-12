.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: ut
ut:
	@go test -race ./...

.PHONY: cleanup
cleanup:
	@cd tests && rm test.l.gz && rm test.l.zst

.PHONY: check
check:
	@$(MAKE) --no-print-directory tidy
	@$(MAKE) --no-print-directory ut
	@$(MAKE) --no-print-directory cleanup