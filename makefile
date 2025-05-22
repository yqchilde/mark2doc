.PHONY: release build

release: build upx compress

GO_FLAGS = -ldflags="-s -w"

build:
	@$(MAKE) --no-print-directory \
    build-darwin-amd64 build-darwin-arm64 \
    build-windows-amd64 build-windows-arm64 \
    build-linux-amd64 build-linux-arm64

build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${GO_FLAGS} -o build/darwin-amd64/mark2doc

build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ${GO_FLAGS} -o build/darwin-arm64/mark2doc

build-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${GO_FLAGS} -o build/windows-amd64/mark2doc.exe

build-windows-arm64:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build ${GO_FLAGS} -o build/windows-arm64/mark2doc.exe

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${GO_FLAGS} -o build/linux-amd64/mark2doc

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${GO_FLAGS} -o build/linux-arm64/mark2doc

upx:
	upx --best --lzma build/darwin-amd64/mark2doc; \
    upx --best --lzma build/windows-amd64/mark2doc.exe; \
    upx --best --lzma build/linux-amd64/mark2doc;

compress:
	@$(MAKE) --no-print-directory \
    compress-darwin-amd64 compress-darwin-arm64 \
    compress-windows-amd64 compress-windows-arm64 \
    compress-linux-amd64 compress-linux-arm64

compress-darwin-amd64:
	tar -czvf build/mark2doc-darwin-amd64.tar.gz -C build/darwin-amd64/ mark2doc; \
    rm -rf build/darwin-amd64

compress-darwin-arm64:
	tar -czvf build/mark2doc-darwin-arm64.tar.gz -C build/darwin-arm64/ mark2doc; \
    rm -rf build/darwin-arm64

compress-windows-amd64:
	tar -czvf build/mark2doc-windows-amd64.tar.gz -C build/windows-amd64/ mark2doc.exe; \
    rm -rf build/windows-amd64

compress-windows-arm64:
	tar -czvf build/mark2doc-windows-arm64.tar.gz -C build/windows-arm64/ mark2doc.exe; \
    rm -rf build/windows-arm64

compress-linux-amd64:
	tar -czvf build/mark2doc-linux-amd64.tar.gz -C build/linux-amd64/ mark2doc; \
    rm -rf build/linux-amd64

compress-linux-arm64:
	tar -czvf build/mark2doc-linux-arm64.tar.gz -C build/linux-arm64/ mark2doc; \
    rm -rf build/linux-arm64