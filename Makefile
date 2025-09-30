.PHONY: test
test:
	mkdir -p coverage/covdata
# Use the new binary format to ensure integration tests and cross-package calls are counted towards coverage
# https://go.dev/blog/integration-test-coverage
# -p 1 disable parallel testing in favor of streaming log output - https://github.com/golang/go/issues/24929#issuecomment-384484654
	go test -race -cover -covermode atomic -v -vet=all -timeout 15m -p 1\
		./... \
		-args -test.gocoverdir="${PWD}/coverage/covdata" \
		| ts -s
# NB: ts command requires moreutils package; awk trick from https://stackoverflow.com/a/25764579 doesn't stream output
	go tool covdata percent -i=./coverage/covdata
	# Convert to old text format for coveralls upload
	go tool covdata textfmt -i=./coverage/covdata -o ./coverage/covprofile
	go tool cover -html=./coverage/covprofile -o ./coverage/coverage.html

.PHONY: short-test
short-test:
	go test -v -short ./...

#NB: CI uses the golangci-lint Github action, not this target
.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.5 run -v

.PHONY: checks
checks: check_tidy check_vuln check_modern

.PHONY: check_vuln
check_vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
# if we use more tools, we can switch to go tool -modfile=tools.mod
# there is good discussion at https://news.ycombinator.com/item?id=42845323

check_tidy:
	go mod tidy
	# Verify that `go mod tidy` didn't introduce any changes. Run go mod tidy before pushing.
	git diff --exit-code --stat go.mod go.sum

.PHONY: check_modern
check_modern:
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@v0.20.0 ./...
# non-zero exit status on issues found
# nb: modernize is not part of golangci-lint yet - https://github.com/golangci/golangci-lint/issues/686	
