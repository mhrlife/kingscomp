
test-integration:
	TEST_INTEGRATION=true go test ./... -v


serve:
	go run main.go serve

dev:generate
	go run main.go serve

ngrok:
	(source 'filename.env' && bash 'scriptname.sh')

generate:
	templ generate