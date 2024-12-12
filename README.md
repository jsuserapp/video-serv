# video-serv
A Simple Video Server

下载
```bash
git clone https://github.com/jsuserapp/video-serv
```
安装
```bash
go build
```
默认使用 80 端口, 如果 80 已经被占用, 可以在 conf.js 的 http 字段配置新的端口

局域网访问, windows 可能需要把可执行文件添加到允许通过防火墙.

在centos 添加80端口通过防火墙
```bash
sudo firewall-cmd --zone=public --add-port=80/tcp --permanent
```
重启防火墙
```bash
sudo firewall-cmd --reload
```
