all: lurker

lurker: *.go
	go mod download
	go build -o lurker

test: *.go
	go test .


install: lurker
	/usr/bin/install -m 755 -D lurker /usr/local/bin/

install-systemd: install
	/usr/bin/install -m 644 -D systemd/lurker /etc/default/
	/usr/bin/install -m 644 -D systemd/lurker.service /lib/systemd/system/
