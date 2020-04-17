gofwd: cmd.go duoauth.go examples.go geoip.go nics.go
	go build -tags netgo -ldflags '-extldflags "-static" -s -w'

clean:
	rm gofwd
