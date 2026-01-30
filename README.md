# hui_lv_server
## centos 安装 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./deploy/hui_lv_server_linux main.go