package Socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"
)

type ServerSocks struct {
	LocalAddress *net.TCPAddr
}

func init() {
	log.SetPrefix("[Server]")
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

var byteBuffer = make([]byte, 4096)

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
	go func() {
		for {
			runtime.Gosched()
			time.Sleep(1 * time.Second)
		}
	}()
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
	defer localConn.Close()
	log.Println("本地接收到请求", localConn.RemoteAddr(), "=====>>", localConn.LocalAddr())

	/* 解析Socks5协议 */
	dstConn, err := serverSocks.resolveSocks5Protocol(localConn)
	if err != nil {
		log.Println("代理结束")
		return
	}
	defer dstConn.Close()
	log.Println("开始转发数据")

	isOver := make(chan bool)

	go func() {
		/* 并将目标服务器的结果返回给代理服务器 */
		log.Println("代理服务器<<==========目的服务器")
		for {
			if err = serverSocks.transprotData(dstConn, localConn); err != nil {
				dstConn.Close()
				localConn.Close()
				break
			}
		}
		log.Println("代理服务器==========>>目的服务器，传输完成")
		isOver <- true
	}()

	log.Println("代理服务器==========>>目的服务器")
	for {
		/* 将代理服务器接收到的请求转发给目标服务器 */
		if err := serverSocks.transprotData(localConn, dstConn); err != nil {
			dstConn.Close()
			localConn.Close()
			break
		}
	}
	log.Println("代理服务器==========>>目的服务器，传输完成")
	<-isOver
	log.Println("代理结束")
}

func (serverSocks *ServerSocks) resolveSocks5Protocol(localConn *net.TCPConn) (*net.TCPConn, error) {
	/* 解析Socket5 协议 */
	n, err := localConn.Read(byteBuffer)
	if err != nil {
		log.Println("==>从请求中读取数据失败", err)
		return nil, err
	}
	log.Println("==>接收到:", byteBuffer[:n])
	// ver nmethods methods
	if byteBuffer[0] != 0x05 {
		log.Println("==>非Socks5协议")
		return nil, errors.New("非Socks5协议")
	}

	// 默认采用 不认证的方式 TODO
	// 响应客户端 []byte{0x05,0x00}
	_, err = localConn.Write([]byte{0x05, 0x00})
	if err != nil {
		log.Println("<==返回数据发送错误", err)
		return nil, err
	}
	log.Println("<==返回数据:", []byte{0x05, 0x02})

	n, err = localConn.Read(byteBuffer)
	if err != nil {
		log.Println("==>从请求中读取数据失败", err)
		return nil, err
	}
	log.Println("==>接收到:", byteBuffer[:n])
	atyp := byteBuffer[3]
	var desIp []byte // 目的IP
	switch atyp {
	case 0x01: // IPv4
		desIp = byteBuffer[4 : 4+net.IPv4len]
	case 0x03: // Domainname
		// 第4个字节是域名长度
		domainname := string(byteBuffer[5 : 5+byteBuffer[4]])
		log.Println("解析IP:", domainname)
		addr, err := net.ResolveIPAddr("ip", domainname)
		if err != nil {
			log.Println("解析IP出错:", err)
			return nil, err
		}
		desIp = addr.IP.To4()
	case 0x04: // IPv6
		desIp = byteBuffer[4 : 4+net.IPv6len]
	default: // 未定义
		return nil, errors.New("未定义的协议")
	}
	log.Println("目的IP:", desIp)
	desPort := byteBuffer[n-2:] // 目的端口
	log.Println("目的Port:", binary.BigEndian.Uint16(desPort))

	desAddr := &net.TCPAddr{
		IP:   desIp,
		Port: int(binary.BigEndian.Uint16(desPort)),
	}

	/* 连接 目的服务器  */
	dstConn, err := net.DialTCP("tcp", nil, desAddr)
	if err != nil {
		log.Println("连接目的服务器失败,", err)
		// TODO
		return nil, err
	}

	log.Println("连接目的服务器成功")
	n, err = localConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		log.Println("<==返回数据发生错误", err)
		return nil, err
	}
	log.Println("<==返回数据:", []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	return dstConn, nil
}

func (serverSocks *ServerSocks) transprotData(from *net.TCPConn, to *net.TCPConn) error {
	for {
		n, err := from.Read(byteBuffer)
		if n > 0 {
			_, err := to.Write(byteBuffer[:n])
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
}
