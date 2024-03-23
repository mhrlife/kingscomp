
test-integration:
	TEST_INTEGRATION=true go test ./... -v

test:
	 TEST_INTEGRATION=false go test ./... -v

serve:
	go run main.go serve

dev:generate
	go run main.go serve

generate:
	templ generate