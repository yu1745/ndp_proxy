# ndp_proxy

使得应用可以绑定并使用指定段内的任意ipv6

## 注意
- 仅linux可用，依赖libpcap和iproute2
- 家宽环境下使用时设置的网段建议不要包含第一个/120段，因为很多路由器的dhcpv6策略是给v6分配和v4一样的地址，比如一台电脑的v4是192.168.1.12，路由器v6前缀是240e::,那么电脑会被分配到240e::12作为v6地址，此时如果使用了第一个/120段会导致ip冲突
- 同时应注意避开slaac分配使用的地址
# 安装
```bash
apt install -y libpcap-dev libpcap0.8 iproute2
go install github.com/yu1745/ndp_proxy/cmd/ndp_proxy@latest
```
非apt包管理的系统请自行替换成对应的命令
# 使用
```
ndp_proxy -i eth0 -p 240e:1234::/112
-i 网卡名称
-p ip段，格式如2401::/64所示
```
# 原理
首先使用sysctl配置net.ipv6.ip_nonlocal_bind为1，允许绑定不属于自己的地址

然后执行ip r add local [ip段] dev lo命令配置路由表，使得要绑定的块的包不被操作系统扔掉，能够正常的到达用户空间

然后使用pcap监听数据链路层，筛选出Neighbor Solicitation包，伪造Neighbor Advertisement包回应路由器，使路由器认为该地址是本机持有的

随后就可以使用参数中的段中的任意ip发起连接
