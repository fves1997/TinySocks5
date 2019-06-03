package Socks5

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"syscall"
)

type RedSocks struct {
	LocalAddress      *net.TCPAddr
	LocalSocksAddress *net.TCPAddr
}

func (r *RedSocks) Listener() error {
	listener, err := net.ListenTCP("tcp", r.LocalAddress)
	if err != nil {
		return err
	}
	for {
		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Accept Error:", err)
			continue
		}
		dstAddress, tcpConn, err := getOriginalAddress(tcpConn)
		if err != nil {
			log.Println("Get Remote Address Error:", err)
		}
		go func() {
			defer tcpConn.Close()
			/* 与本地socks建立连接 */
			dialTCP, err := net.DialTCP("tcp", nil, r.LocalSocksAddress)
			if err != nil {
				log.Println("Dial LocalSocks Error:", err)
				return
			}
			err = handShack(dstAddress, dialTCP)
			if err != nil {
				return
			}
		}()
	}
}

func handShack(dst *net.TCPAddr, conn *net.TCPConn) error {
	/* socks客户端  握手 */
	/* TODO */
	return nil
}

const SO_ORIGINAL_DST = 80


// 获取原目的地址和端口
func getOriginalAddress(tcpConn *net.TCPConn) (*net.TCPAddr, *net.TCPConn, error) {
	tcpConnFile, err := tcpConn.File()
	if err != nil {
		return nil, nil, err
	}
	_ = tcpConn.Close()

	ipMreq, err := syscall.GetsockoptIPv6Mreq(int(tcpConnFile.Fd()), syscall.SOL_IP, SO_ORIGINAL_DST)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println(ipMreq.Multiaddr)
	port := binary.BigEndian.Uint16(ipMreq.Multiaddr[2:4])
	ip := ipMreq.Multiaddr[4:8]

	fileConn, err := net.FileConn(tcpConnFile)
	if err != nil {
		return nil, nil, err
	}
	return &net.TCPAddr{
		IP:   ip,
		Port: int(port),
	}, fileConn.(*net.TCPConn), nil
}
