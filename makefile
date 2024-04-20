#Test
test:
	go test -count=1 -p=1 ./... -v | grep -v "no test"

#Speed
speed:
	hyperfine --warmup 1 'go run main.go'