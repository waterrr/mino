package mino

import (
	"crypto/tls"
	"crypto/x509"
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/proto"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"net"
)

const (
	Version1 = 0x01
)

type packType uint8

const (
	msgRequest packType = iota
	msgResponse
)

var ErrAuth = errors.New("auth error")

type Server struct {
	net.Conn
	// 公玥文件
	CertFile string
	// 私玥文件
	KeyFile string
	// 请求信息
	r *RequestMessage
}

// 握手
func (conn *Server) Handshake(auth proto.BasicAuthFunc) (err error) {
	cert, er := tls.LoadX509KeyPair(conn.CertFile, conn.KeyFile)
	if er != nil {
		_ = conn.Close()
		return er
	}
	conn.Conn = tls.Server(conn.Conn, &tls.Config{Certificates: []tls.Certificate{cert}})
	if _, p, er := readPack(conn); er != nil {
		_ = conn.Close()
		return er
	} else {
		req := new(RequestMessage)
		if er := req.unmarshal(p); er != nil {
			_ = conn.Close()
			return er
		}
		if auth != nil {
			if auth(&proto.AuthInfo{
				Username:     req.Username,
				Password:     req.Password,
				RemoteAddr:   conn.RemoteAddr().String(),
				HardwareAddr: req.MacAddress,
			}) {
			} else {
				_ = conn.Close()
				return ErrAuth
			}
		}
		conn.r = req
	}
	return
}

// 获取链接信息
func (conn *Server) Info() (network, address string, err error) {
	switch conn.r.Network {
	case NetworkUdp:
		network = "udp"
	default:
		network = "tcp"
	}
	return network, conn.r.Address, nil
}

// 发送错误
func (conn *Server) SendError(err error) error {
	if e, ok := err.(tlsError); ok {
		return writeRspMsg(conn, uint8(e), e.Error())
	}
	return writeRspMsg(conn, unknownError, err.Error())
}

// 发送连接成功
func (conn *Server) SendSuccess() error {
	return writeRspMsg(conn, succeeded, "OK")
}

type Client struct {
	net.Conn
	// 认证公玥
	RootCa string
	// 用户名
	Username string
	// 密码
	Password string
}

func (conn *Client) Handshake() (err error) {
	cfg, er := conn.cfgGen()
	if er != nil {
		return er
	}
	conn.Conn = tls.Client(conn.Conn, cfg)
	return
}

func (conn *Client) cfgGen() (cfg *tls.Config, err error) {
	if len(conn.RootCa) == 0 {
		cfg = &tls.Config{InsecureSkipVerify: true}
	} else {
		pool := x509.NewCertPool()
		caCrt, e := ioutil.ReadFile(conn.RootCa)
		if e != nil {
			return nil, e
		}
		pool.AppendCertsFromPEM(caCrt)
		cfg = &tls.Config{RootCAs: pool}
	}
	return
}

func (conn *Client) Connect(network, address string) (err error) {
	m := new(RequestMessage)
	switch network {
	case "udp":
		m.Network = NetworkUdp
	default:
		m.Network = NetworkTcp
	}
	m.Address = address
	m.Username = conn.Username
	m.Password = conn.Password
	m.MacAddress = getHardwareAddr()
	if er := writePack(conn, msgRequest, m.marshal()); er != nil {
		return er
	}
	if _, p, er := readPack(conn); er != nil {
		return er
	} else {
		rsp := new(ResponseMessage)
		if er := rsp.unmarshal(p); er != nil {
			return er
		}
		if rsp.Code != succeeded {
			if rsp.Code == unknownError {
				return errors.New(rsp.Message)
			}
			return tlsError(rsp.Code)
		}
	}
	return
}

type Checker struct {
}

const (
	// TLS握手记录
	TlsRecordTypeHandshake uint8 = 22
)

// 判断是否为HTTP协议
func (d *Checker) Check(r io.Reader) (bool, error) {
	// 读3个字节
	buf := make([]byte, 3)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false, err
	}
	if buf[0] != TlsRecordTypeHandshake {
		return false, nil
	}
	// 0300~0305
	if buf[1] != 0x03 {
		return false, nil
	}
	if buf[2] > 0x05 {
		return false, nil
	}
	return true, nil
}

type Protocol struct {
}

func (c *Protocol) Name() string {
	return "mino"
}

// 创建HTTP接收器
func (c *Protocol) Server(conn net.Conn, config config.Config) proto.Server {
	return &Server{
		Conn:     conn,
		CertFile: config.String(mino.KeyCertFile),
		KeyFile:  config.String(mino.KeyKeyFile),
	}
}

// 创建HTTP请求器
func (c *Protocol) Client(conn net.Conn, config config.Config) proto.Client {
	return &Client{
		Conn:     conn,
		Username: config.String(mino.KeyUsername),
		Password: config.String(mino.KeyPassword),
		RootCa:   config.String(mino.KeyRootCa),
	}
}

func (c *Protocol) Checker(config config.Config) proto.Checker {
	return &Checker{}
}

// 获取Mac地址
func getHardwareAddr() []net.HardwareAddr {
	h := []net.HardwareAddr{}
	if its, _ := net.Interfaces(); its != nil {
		for _, it := range its {
			if it.Flags&net.FlagLoopback == 0 {
				h = append(h, it.HardwareAddr)
			}
		}
	}
	return h
}

// 写入包
func writePack(w io.Writer, typ packType, p []byte) (err error) {
	buf := make([]byte, 4)
	buf[0] = Version1
	buf[1] = byte(typ)
	binary.BigEndian.PutUint16(buf[2:], uint16(len(p)))
	buf = append(buf, p...)
	_, err = w.Write(buf)
	return
}

// 写信息
func writeRspMsg(w io.Writer, code uint8, msg string) (err error) {
	m := &ResponseMessage{Code: code, Message: msg}
	if er := writePack(w, msgResponse, m.marshal()); er != nil {
		return er
	}
	return nil
}

// 读取包
func readPack(r io.Reader) (typ packType, p []byte, err error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, nil, err
	}
	typ = packType(buf[1])
	l := binary.BigEndian.Uint16(buf[2:])
	p = make([]byte, l)
	if _, err := io.ReadFull(r, p); err != nil {
		return 0, nil, err
	}
	return
}

func init() {
	proto.Add(&Protocol{})
}