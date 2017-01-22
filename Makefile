croshclip: croshclip.go
	go build croshclip.go

croshclip-static: croshclip.go
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s -w' croshclip.go

clean:
	rm -f croshclip
