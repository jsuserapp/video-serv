package main

import (
	"VideoServ/httpserv"
)

const (
	KeyVideoPath = "video_path"
	KeyHttpPort  = "http_port"
)

func main() {
	httpserv.StartHttpServ()
}
