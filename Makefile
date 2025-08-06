APP_NAME=weex-watchdog

.PHONY: build dev clean

build:
	GOOS=windows GOARCH=amd64 go build -o bin/$(APP_NAME)_windows_amd64.exe main.go \
	&& GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME)_linux_amd64 main.go \
	&& GOOS=darwin GOARCH=amd64 go build -o bin/$(APP_NAME)_darwin_amd64 main.go \
	&& GOOS=darwin GOARCH=arm64 go build -o bin/$(APP_NAME)_darwin_arm64 main.go \

clean:
	rm -f $(APP_NAME)_windows_amd64.exe $(APP_NAME)_linux_amd64 $(APP_NAME)_darwin_amd64 $(APP_NAME)_darwin_arm64