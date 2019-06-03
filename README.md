# TinySocks5
简单的socks5实现

# Install
```
git clone https://github.com/leiqiao0324/TinySocks5.git
```

# RUN
```
修改config.json里面的server为服务器ip
server
go run cmd/server.go

client
go run cmd/local.go
```
# TODO 
1. Redsocks  
重定向tcp连接到socks5代理，实现socks5透明代理
详情:[redsocks](https://github.com/darkk/redsocks)
2. 
