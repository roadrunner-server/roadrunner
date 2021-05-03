clean:
	rm -rf rr-jobbroadcast
install: all
	cp rr-broadcast /usr/local/bin/rr-broadcast
uninstall: 
	rm -f /usr/local/bin/rr-broadcast
test:
	composer update
	go test -v -race -cover
