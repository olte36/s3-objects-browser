BINARY := s3browser
BUILD_DIR := out

.PHONY: build clean integration-test test

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) .

test:
	go test ./...

integration-test:
	S3BROWSER_AWS_INTEGRATION=1 go test -tags=integration -run TestAWSIntegration ./...

clean:
	rm -rf $(BUILD_DIR)
