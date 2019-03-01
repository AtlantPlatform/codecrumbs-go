all:

build:
	go build github.com/AtlantPlatform/codecrumbs-go/cmd/cc-go

install:
	go install github.com/AtlantPlatform/codecrumbs-go/cmd/cc-go

examples:
	cc-go -p "Example Go" \
		--dir examples/example-go/ \
		--entry examples/example-go/cmd/example_api/main.go \
		--out examples/output/api.md

	cc-go render examples/output/api.md

.PHONY: build install examples
