# tcpcheck

一个使用 Go 编写的，用于检查网络 TCP 连接稳定性的工具

## 协议

客户端消息长度均为 4 字节，服务器消息均为固定长度

1. RoundTrip

```text
CLIENT: RDTR
SERVER: OK
```

2. Transfer 10m

```text
CLIENT: T10M
SERVER: <10m bytes>
```

Server will close the connection after `T10M` action

## 许可证

Guo Y.K., MIT License