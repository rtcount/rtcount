package main

import (
	"bufio"
	"fmt"
	"net"
	//	"time"
)

func Rtc_StartTcpServer(rtc_conf *RTC_Conf) {

	for _, table := range rtc_conf.Table {
		if table.TcpPort > 0 {
			//fmt.Println("Rtc_StartTcpServer : ", table.TcpPort, table.Name)
			go Rtc_StartTcpPort(table.TcpPort, table.Name)
		}
	}
}

func Rtc_StartTcpPort(port int, table_name string) {
	//	var tcpAddr *net.TCPAddr

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+fmt.Sprintf("%d", port))
	if err != nil {
		fmt.Println("Rtc_StartTcpPort for table_name err:", table_name, port, err)
		return
	}

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Rtc_StartTcpPort for table_name err:", table_name, port, err)
		return
	}

	defer tcpListener.Close()

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}

		fmt.Println("A client connected : " + tcpConn.RemoteAddr().String())
		go tcpPipe(tcpConn, table_name)
	}

}

func tcpPipe(conn *net.TCPConn, table_name string) {
	ipStr := conn.RemoteAddr().String()
	defer func() {
		fmt.Println("disconnected :" + ipStr)
		conn.Close()
	}()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		WebDataHandle(table_name, message)
	}
}
