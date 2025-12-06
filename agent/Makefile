.PHONY: build build-all clean

# 默认构建当前平台
build:
	go build -ldflags="-s -w" -o xboard-agent .

# 构建所有平台
build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o xboard-agent-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o xboard-agent-linux-arm64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o xboard-agent-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o xboard-agent-darwin-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o xboard-agent-windows-amd64.exe .

clean:
	rm -f xboard-agent xboard-agent-*
