PACKAGE=marathon-consul-discovery_linux_amd64.zip
BINARY=/opt/gopath/bin/marathon-consul-discovery

$(BINARY): deps
	go build ./...

deps:
	go get ./...

$(PACKAGE): $(BINARY)
	zip -j $@ $<

package: $(PACKAGE)
