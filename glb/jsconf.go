package glb

import (
	"ju"
)

const (
	JsConfName = "conf.js"
)

type Conf struct {
	TLS struct {
		Enable bool   `json:"enable"`
		Pem    string `json:"pem"`
		Key    string `json:"key"`
		Port   string `json:"port"`
	} `json:"tls"`
	HTTP struct {
		Enable bool   `json:"enable"`
		Port   string `json:"port"`
	} `json:"http"`
	WS struct {
		Enable bool   `json:"enable"`
		Port   string `json:"port"`
	} `json:"ws"`
	WSS struct {
		Enable bool   `json:"enable"`
		Port   string `json:"port"`
		Pem    string `json:"pem"`
		Key    string `json:"key"`
	} `json:"wss"`
	VideoServer struct {
		SourceList []struct {
			Path string `json:"path"`
			Hash string `json:"hash"`
		} `json:"source_list"`
	} `json:"video_server"`
}

func (cf *Conf) Save() bool {
	confFile := "./" + JsConfName
	return ju.SaveJsConfFile(confFile, &conf)
}
func init() {
	loadConf()
}

var conf Conf

// noinspection GoUnusedExportedFunction
func GetConf() *Conf {
	return &conf
}
func loadConf() {
	confFile := "./" + JsConfName
	if !ju.LoadJsConfFile(confFile, &conf) {
		conf.HTTP.Port = "80"
		conf.TLS.Port = "443"
		ju.SaveJsConfFile(confFile, &conf)
	}
	for i, source := range conf.VideoServer.SourceList {
		conf.VideoServer.SourceList[i].Path = PathToLinux(source.Path)
	}
}
