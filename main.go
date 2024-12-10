package main

import (
	"VideoServ/httpserv"
	"bufio"
	"ju"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	KeyVideoPath = "video_path"
	KeyHttpPort  = "http_port"
)

func main() {
	httpserv.StartHttpServ()
}
func readLine() []string {
	var lines []string
	file, err := os.Open(".conf")
	if ju.CheckFailure(err) {
		return lines
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
func readConf() *Conf {
	var conf Conf
	lines := readLine()
	if len(lines) == 0 {
		return &conf
	}
	for _, line := range lines {
		pos := strings.Index(line, "=")
		if pos == -1 {
			continue
		}
		key := line[:pos]
		if key == KeyVideoPath {
			conf.VideoPath = line[pos+1:]
		} else if key == KeyHttpPort {
			conf.HttpPort = line[pos+1:]
		}
	}
	return &conf
}

type Conf struct {
	VideoPath string
	HttpPort  string
}

func IsExistingDir(path string) string {
	if path == "" {
		return "请设置有效的视频文件夹路径"
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "指定的视频文件夹不存在"
		}
		return err.Error()
	}
	if info.IsDir() {
		return ""
	}

	return "指定的视频文件夹不能是一个文件"
}

// GetLocalIPAddresses 获取并返回本机的所有非环回 IPv4 地址
func getLocalIPAddresses() ([]string, error) {
	var ips []string
	interfaces, err := net.Interfaces() // 获取所有网络接口
	if err != nil {
		return nil, err
	}

	// 遍历所有接口
	for _, iface := range interfaces {
		// 检查网络接口是否激活并且不是回环接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue // 不是激活状态或者是回环接口，跳过
		}

		// 获取接口的所有地址
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		// 处理接口的每一个地址
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 仅保留 IPv4 地址
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}

	return ips, nil
}

func startHttpServ() {
	conf := readConf()
	ers := IsExistingDir(conf.VideoPath)
	if ers != "" {
		ju.LogRed(ers)
		return
	}
	if conf.HttpPort == "" {
		conf.HttpPort = "80"
	}
	fs := http.FileServer(http.Dir(conf.VideoPath))
	http.Handle("/", fs)

	ips, err := getLocalIPAddresses()
	for _, ip := range ips {
		ju.LogGreen("开始启动视频服务在", ip, ":", conf.HttpPort)
	}
	// 启动服务器
	err = http.ListenAndServe(":"+conf.HttpPort, nil)
	if ju.CheckFailure(err) {
		ju.LogRed("启动失败, 是否重复启动, 或者端口被占用?")
		time.Sleep(time.Hour)
	}
}
