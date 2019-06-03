package Socks5

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

type ServerSocks struct {
	LocalAddress *net.TCPAddr
}

func init() {
	log.SetPrefix("[Server]")
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func NewServerSocks(localHost string) *ServerSocks {
	tcpAddr, err := net.ResolveTCPAddr("tcp", localHost)
	if err != nil {
		log.Fatal("本地Host配置错误", err)
	}
	return &ServerSocks{
		LocalAddress: tcpAddr,
	}
}

func (serverSocks *ServerSocks) Listener() {
	listener, err := net.ListenTCP("tcp", serverSocks.LocalAddress)
	if err != nil {
		log.Fatal("本地监听开启失败", err)
	}
	log.Println("开始接收连接...")
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("Accept TCP error", err)
			continue
		}
		go serverSocks.handleTcpConn(conn)
	}
}

func (serverSocks *ServerSocks) handleTcpConn(localConn *net.TCPConn) {
	/* 解析Socks5协议 */
	dstConn, err := serverSocks.handShack(localConn)
	if err != nil {
		log.Println(err)
		localConn.Close()
		log.Println("代理结束")
		return
	}
	_ = dstConn.SetKeepAlive(true)
	go func() {
		defer localConn.Close()
		defer dstConn.Close()
		log.Printf("%s===>%s===>%s\n", localConn.RemoteAddr(), localConn.LocalAddr(), dstConn.RemoteAddr())
		_, _ = io.Copy(dstConn, localConn)
	}()

	go func() {
		defer localConn.Close()
		defer dstConn.Close()
		log.Printf("%s<===%s<===%s\n", localConn.RemoteAddr(), localConn.LocalAddr(), dstConn.RemoteAddr())
		_, _ = io.Copy(localConn, dstConn)
	}()
}

// https://zh.wikipedia.org/wiki/SOCKS
// MaxAddrLen is the maximum size of SOCKS address in bytes.
const MaxAddrLen = 1 + 1 + 255 + 2

// 握手，获取与目的服务器的连接
func (serverSocks *ServerSocks) handShack(rw io.ReadWriter) (*net.TCPConn, error) {
	buff := make([]byte, MaxAddrLen)
	/* 解析Socket5 协议 */
	// 读 ver 和 nmethods
	if _, err := io.ReadFull(rw, buff[:2]); err != nil {
		return nil, err
	}
	// Socks Version
	if buff[0] != SOCKS_VERSION {
		return nil, errors.New("非Socks5协议")
	}
	nmethods := buff[1]
	// 读 methods
	if _, err := io.ReadFull(rw, buff[:nmethods]); err != nil {
		return nil, err
	}
	// TODO 暂时只支持 不需要认证和用户名密码，两种方式
	// 返回 []byte{0x05,0x00}
	if _, err := rw.Write([]byte{0x05, 0x00}); err != nil {
		return nil, err
	}

	// 读 ver，cmd，rsv
	if _, err := io.ReadFull(rw, buff[:3]); err != nil {
		return nil, err
	}

	if buff[0] != SOCKS_VERSION {
		return nil, errors.New("非Socks5协议")
	}

	cmd := buff[1]
	switch cmd {
	case CMD_CONNECT: // CONNECT 请求
		host, err := readIPAndPortToHost(rw, buff)
		if err != nil {
			return nil, err
		}
		conn, err := net.Dial("tcp", host)
		if err != nil {
			// 如果连接目标服务器失败，则server直接断开连接
			return nil, err
		}
		if _, err := rw.Write([]byte{0x5, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}); err != nil {
			_ = conn.Close()
			return nil, err
		}
		return conn.(*net.TCPConn), nil
	case CMD_BIND: // BIND请求
		// TODO 暂不支持
		return nil, errors.New("not support")
	case CMD_UDP: // UDP转发
		// TODO 暂不支持
		return nil, errors.New("not support")
	default:
		return nil, errors.New("not support")
	}
}

// 读取IP和Port 并转换成Host IP:Port
func readIPAndPortToHost(r io.Reader, buff []byte) (string, error) {
	if len(buff) < MaxAddrLen {
		return "", errors.New("buff太小")
	}
	// read atype
	if _, err := io.ReadFull(r, buff[:1]); err != nil {
		return "", err
	}
	atyp := buff[0] // addr 类型
	switch atyp {
	case ATYPE_IPV4: // IPv4
		// 读 IPv4 和 Port
		if _, err := io.ReadFull(r, buff[:6]); err != nil {
			return "", err
		}
		ip := net.IP(buff[:4]).String()
		port := int(uint16(buff[net.IPv4len])<<8 | uint16(buff[net.IPv4len+1]))
		return net.JoinHostPort(ip, strconv.Itoa(port)), nil
	case ATYPE_DOMAIN: // Domain
		// 读取域名的长度
		if _, err := io.ReadFull(r, buff[:1]); err != nil {
			return "", err
		}
		length := buff[0]
		if _, err := io.ReadFull(r, buff[:length+2]); err != nil {
			return "", err
		}
		ip := string(buff[:length])
		port := int(uint16(buff[length])<<8 | uint16(buff[length+1]))
		return net.JoinHostPort(ip, strconv.Itoa(port)), nil
	case ATYPE_IPV6: // IPv6
		if _, err := io.ReadFull(r, buff[:18]); err != nil {
			return "", err
		}
		ip := net.IP(buff[:16]).String()
		port := int(uint16(buff[net.IPv6len])<<8 | uint16(buff[net.IPv6len+1]))
		return net.JoinHostPort(ip, strconv.Itoa(port)), nil
	default:
		return "", errors.New("不支持的地址类型")
	}
}
