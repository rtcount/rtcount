package main

import (
	"./freecache"
	"./seefan/gossdb"
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

var dbpoll *gossdb.Connectors
var g_rtc_conf RTC_Conf
var m int = 1

var sem = make(chan int, 990)

func byteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

func rtcount_before(tablename string, strs []string) {
	//fmt.Println("-----------------------\n")
	//fmt.Println(strs)
	sem <- 1
	go rtcount_handle_talbe(strings.ToLower(tablename), strs)
	<-sem

	//rtcount_core(UNION|SUM|MIN|MAX, strs[3], strs[1], strs[0],strs[2], "")
}

func rtcount_correct_param(line *[]string) {

	l := len(*line)
	for i := 0; i < l; i++ {
		fmt.Print("\nxxlin:", i, (*line)[i])
		st := (*line)[i]
		if st == " " {
			(*line)[i] = "nil"
			fmt.Print("\nnil:", (*line)[i])
		}
		fmt.Print("\nlen:", len(st))
	}
}

func rtcount_handle_talbe(tablename string, strs []string) {
	sem <- 1 // 等待队列缓冲区非满
	//rtcount_core(tablename, UNION, strs[3], strs[1], strs[0], strs[2], strs[3])
	rtcount_core(tablename, strs)
	<-sem // 请求处理完成，准备处理下一个请求
}

func rtcount_core(tablename string, strs []string) {

	table := RTC_conf_GetTableByName(tablename, &g_rtc_conf)
	if table == nil {
		fmt.Printf("don't find table[%s]\n", tablename)
		return
	}

	//fmt.Printf("table column_terminated[%s]\n", table.CTerminated)

	conn, err := dbpoll.NewClient()
	if err != nil {
		time.Sleep(1)
		conn, err = dbpoll.NewClient()
		if err != nil {
			fmt.Println("Failed to create new client:", err)
			return
		}
	}
	defer conn.Close()

	for _, key := range table.Keys {
		rtcount_core_table_key(conn, table, &key, strs)
	}

}

func rtcount_core_init_kv(rtc_conf *RTC_Conf, dbpoll **gossdb.Connectors) (err error) {
	*dbpoll, err = gossdb.NewPool(&gossdb.Config{
		Host:             rtc_conf.Kvdb.Host,
		Port:             rtc_conf.Kvdb.Port,
		MinPoolSize:      rtc_conf.Kvdb.MinPoolSize,
		MaxPoolSize:      rtc_conf.Kvdb.MaxPoolSize,
		AcquireIncrement: 1,
	})
	if err != nil {
		fmt.Println("rtcount_core_init_kv init error, %s", err)
		return err
	}
	// test connection
	conn, err := (*dbpoll).NewClient()
	if err != nil {
		fmt.Println("rtcount_core_init_kv Failed to create new client, %s", err)
		return err
	}
	defer conn.Close()
	return

}

func processLine(line []byte) {

	//os.Stdout.Write(line)
	s1 := byteString(line)
	strs := strings.Split(s1, "\x02")
	if len(strs) < 3 {
		fmt.Printf("err --[%s] \n", s1)
		return
	}
	rtcount_before("test_game", strs)
}

func ReadLine(filePth string, hookfn func([]byte)) error {
	file, err := os.Open(filePth)
	if err != nil {
		return err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hookfn(scanner.Bytes())
		m = m + 1
		if m >= 10 {
			return nil
		}
		//scanner.Text() = "hello, world!"
	}
	return scanner.Err()
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	arg_num := len(os.Args)
	if arg_num != 2 {
		fmt.Printf("usage: ./count your.xml\n")
		return
	}

	xml_conf := os.Args[1]
	rtc_conf, msg, ret := RTC_conf_init(xml_conf)
	if ret != true {
		fmt.Println(msg, ret)
		return
	}
	g_rtc_conf = rtc_conf
	//fmt.Println(rtc_conf)

	var err = rtcount_core_init_kv(&g_rtc_conf, &dbpoll)
	if err != nil {
		fmt.Println("rtcount_core_init_kv init error")
		return
	}

	/*
		//local test file
			start := time.Now().Unix()
			ReadLine("20160901.data", processLine)

			end := time.Now().Unix()
			fmt.Println("run:", (end - start))
	*/

	freecache.Localcache_cache_init(1024 * 10240)

	go Rtc_StartTcpServer(&g_rtc_conf)
	Rtc_StartWebServer(&g_rtc_conf)
}
