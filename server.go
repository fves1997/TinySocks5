package Socks5

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	SocksVersion = 0x05

	CmdConnect = 0x01
	CmdBind    = 0x02
	CmdUdp     = 0x03

	ATypeIpv4   = 0x01
	ATypeDomain = 0x03
	ATypeIpv6   = 0x04
)

var (
	UnSupportedCommand     = []byte{0x05, 0x07, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	ConnectionRefused      = []byte{0x05, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0}
	UnSupportedAddressType = []byte{0x05, 0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	SucceedConnection      = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

func init() {
	log.SetPrefix("[Server]")
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

type ServerSocks struct {
}

func NewServerSocks() *ServerSocks {
	return &ServerSocks{}
}

func (s *ServerSocks) Listener(localHost string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", localHost)
	if err != nil {
		log.Fatalf("localHost Error:%s\n", err.Error())
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("listen Tcp Error:%s", err.Error())
	}
	go func() {
		for {
			accept, err := listener.Accept()
			if err != nil {
				log.Printf("Accept tcp Error:%s", err.Error())
				_ = listener.Close()
				return
			}
			s.handleTcpConn(accept)
		}
	}()
}

func (s *ServerSocks) handleTcpConn(conn net.Conn) {
	remoteConn, err := s.shakeHands(conn)
	if err != nil {
		log.Printf("shake hands Error:%s\n", err.Error())
		_ = conn.Close()
		return
	}
	log.Printf("%s-->%s \n", conn.LocalAddr(), remoteConn.RemoteAddr())
	go func() {
		n, _ := io.Copy(remoteConn, conn)
		log.Printf("%s-->%s %d bytes\n", remoteConn.RemoteAddr(), conn.RemoteAddr(), n)
	}()

	go func() {
		n, _ := io.Copy(conn, remoteConn)
		log.Printf("%s-->%s %dbytes", conn.RemoteAddr(), remoteConn.LocalAddr(), n)
	}()
}

const MaxAddrLen = 1 + 1 + 255 + 2

// 握手，握手完成后返回与目的服务器的连接
func (s *ServerSocks) shakeHands(rw io.ReadWriter) (*net.TCPConn, error) {
	buff := make([]byte, MaxAddrLen)
	/* 解析Socket5 协议 */
	// 认证阶段
	// +----+----------+----------+
	// |VER | NMETHODS | METHODS  |
	// +----+----------+----------+
	// | 1  |    1     |  1~255   |
	// +----+----------+----------+
	if _, err := io.ReadFull(rw, buff[:2]); err != nil {
		return nil, err
	}
	// Socks Version
	if buff[0] != SocksVersion {
		return nil, errors.New(fmt.Sprintf("Version is %d", buff[0]))
	}
	length := buff[1]
	if _, err := io.ReadFull(rw, buff[:length]); err != nil {
		return nil, err
	}
	// 认证响应
	// +----+--------+
	// |VER | METHOD |
	// +----+--------+
	// | 1  |   1    |
	// +----+--------+
	// TODO 暂时只支持 不需要认证，两种方式
	// 返回 []byte{0x05,0x00}
	if _, err := rw.Write([]byte{0x05, 0x00}); err != nil {
		return nil, err
	}
	// 连接阶段
	// +----+-----+-------+------+----------+----------+
	// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	// +----+-----+-------+------+----------+----------+
	// | 1  |  1  |   1   |  1   | Variable |    2     |
	// +----+-----+-------+------+----------+----------+
	if _, err := io.ReadFull(rw, buff[:4]); err != nil {
		return nil, err
	}
	if buff[0] != SocksVersion {
		return nil, errors.New(fmt.Sprintf("Version is %d", buff[0]))
	}
	switch buff[1] {
	case CmdConnect: // CONNECT 请求
		remoteHost := ""
		switch buff[3] {
		case ATypeIpv4: // IPv4
			// 读 IPv4 和 Port
			if _, err := io.ReadFull(rw, buff[:6]); err != nil {
				return nil, err
			}
			ip := net.IP(buff[:4]).String()
			port := int(uint16(buff[net.IPv4len])<<8 | uint16(buff[net.IPv4len+1]))
			remoteHost = fmt.Sprintf("%s:%d", ip, port)
		case ATypeDomain: // Domain
			// 读取域名的长度
			if _, err := io.ReadFull(rw, buff[:1]); err != nil {
				return nil, err
			}
			length := buff[0]
			if _, err := io.ReadFull(rw, buff[:length+2]); err != nil {
				return nil, err
			}
			ip := string(buff[:length])
			port := int(uint16(buff[length])<<8 | uint16(buff[length+1]))
			remoteHost = fmt.Sprintf("%s:%d", ip, port)
		case ATypeIpv6: // IPv6
			if _, err := io.ReadFull(rw, buff[:18]); err != nil {
				return nil, err
			}
			ip := net.IP(buff[:16]).String()
			port := int(uint16(buff[net.IPv6len])<<8 | uint16(buff[net.IPv6len+1]))
			remoteHost = fmt.Sprintf("%s:%d", ip, port)
		default:
			_, _ = rw.Write(UnSupportedAddressType)
			return nil, errors.New("address type not supported")
		}
		addr, _ := net.ResolveTCPAddr("tcp", remoteHost)
		remoteConn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			_, _ = rw.Write(ConnectionRefused)
			return nil, errors.New(fmt.Sprintf("connect remote error,Error:%s", err.Error()))
		}
		if _, err := rw.Write(SucceedConnection); err != nil {
			return nil, err
		}
		return remoteConn, nil
	case CmdBind: // BIND请求
		_, _ = rw.Write(UnSupportedCommand)
		return nil, errors.New("command not supported")
	case CmdUdp: // UDP转发
		_, _ = rw.Write(UnSupportedCommand)
		return nil, errors.New("command not supported")
	default:
		_, _ = rw.Write(UnSupportedCommand)
		return nil, errors.New("command not supported")
	}
}
