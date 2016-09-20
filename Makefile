#Format is MAJOR . MINOR . PATCH

VERSION=1.3.0
GO_VERSION=1.6

release: package-linux package-windows package-darwin

package-linux: build-linux tar-linux

package-windows: build-windows tar-windows

package-darwin: build-darwin tar-darwin

build-linux:
	GOOS=linux GOARCH=amd64 go build -o shipyardctl

tar-linux:
	tar -zcvf shipyardctl-$(VERSION).linux.amd64.go$(GO_VERSION).tar.gz shipyardctl

build-windows:
	GOOS=windows GOARCH=amd64 go build -o shipyardctl.exe

tar-windows:
	tar -zcvf shipyardctl-$(VERSION).windows.amd64.go$(GO_VERSION).tar.gz shipyardctl

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o shipyardctl

tar-darwin:
	tar -zcvf shipyardctl-$(VERSION).darwin.amd64.go$(GO_VERSION).tar.gz shipyardctl

clean:
	rm shipyardctl shipyardctl.exe *.tar.gz