# Mino Agent

基于Go语言的网关代理服务，支持网站访问权限限制，转发请求

## 特性

- [x] 代理访问
    - [x] HTTP/HTTPS
    - [x] Socks5
- [x] 自动PAC设置
- [x] Web服务
- [x] 开机自启
- [x] 自动更新
- [x] 热更新配置
- [ ] 权限验证
    - [ ] 启用IP验证
    - [ ] 用户名认证
- [ ] Web面板(desktop)
    - [x] 本地访问不验证权限 
- [ ] 访问控制
    - [ ] 域名黑白名单

## 移动端支持

- [ ] android **计划中**
- [ ] ios

## 多平台支持

从v0.2.1-alpha版本起，增加了对macOS的适配，并且原生支持M1！

## 使用

### 安装

```bash
go install dxkite.cn/mino/cmd/mino
```

### 命令行

`-addr :1080` 监听 `1080` 端口 支持 http/socks5 协议
`-upstream mino://127.0.0.1:8080`
`-pac_file conf/pac.txt` 启用PAC文件，自动设置系统Pac(windows)
```
mino -addr :1080 -pac_file conf/pac.txt -upstream mino://127.0.0.1:8080
```

`-addr :8080` 监听 `8080` 端口，支持 http/socks5/mino协议（需要配置加密密钥）
直连网络
使用公钥 `-cert_file conf/server.crt` 私钥 `-key_file conf/server.key` 加密连接
```
mino -addr :8080 -cert_file conf/server.crt  -key_file conf/server.key
```

### 使用配置

直接运行会加载  `mino.yml` 作为配置文件

```
mino
```

- 默认配置名 `mino.yml`

指定配置文件：
```
mino -c config.yaml
```

### 配置文件示例

```yaml
address: ":1080"
upstream: "mino://199.115.229.64:28648"
```
