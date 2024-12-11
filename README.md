# video-serv
A Simple Video Server

在centos 添加80端口通过防火墙
```bash
sudo firewall-cmd --zone=public --add-port=80/tcp --permanent
```
重启防火墙
```bash
sudo firewall-cmd --reload
```
