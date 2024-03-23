
test-integration:
	TEST_INTEGRATION=true go test ./... -v

test:
	 TEST_INTEGRATION=false go test ./... -v

serve:
	go run main.go serve

scale:
	bash test-scale.sh

dev:generate
	go run main.go serve

generate:
	templ generate