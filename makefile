#Test
test:
	go test -count=1 -p=1 ./... -v | grep -v "no test"

#Speed
speed:
	/usr/bin/time -v go run main.go 