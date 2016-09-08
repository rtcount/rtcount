package main

import (
	"./freecache"
	"./seefan/gossdb"
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var dbpoll *gossdb.Connectors
var g_rtc_conf RTC_Conf
var m int = 1

var sem = make(chan int, 990)

const (
	/* opkey: maxmin|sum|union|count */
	COUNT = 1 << 0
	UNION = 1 << 1
	SUM   = 1 << 2
	MAX   = 1 << 3
	MIN   = 1 << 4
)

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
	go rtcount_handle_talbe(tablename, strs)
	<-sem

	//rtcount_core(UNION|SUM|MIN|MAX, strs[3], strs[1], strs[0],strs[2], "")
}

func rtcount_handle_date(timestmp int64) map[string]string {

	date_map := make(map[string]string)

	tm := time.Unix(timestmp, 0)
	f := tm.Format("200601020304")

	year, week := tm.ISOWeek()
	//fmt.Println("\nweek---", year, week)

	var min5last string
	if f[11:12] > "5" {
		min5last = "5"
	} else {
		min5last = "0"
	}

	var min30last string
	if f[10:12] >= "30" {
		min30last = "30"
	} else {
		min30last = "00"
	}

	var week2 int
	if ((week)%2) == 1 && week > 0 {
		week2 = week - 1
	} else {
		week2 = week
	}

	date_map["ALL"] = "a"
	date_map["MIN"] = "m" + f[0:12]
	date_map["MIN5"] = "5m" + f[0:11] + min5last
	date_map["MIN10"] = "1m" + f[0:11] + "0"
	date_map["MIN30"] = "3m" + f[0:10] + min30last
	date_map["HOUR"] = "h" + f[0:10]
	date_map["DAY"] = "d" + f[0:8]
	date_map["WEEK"] = "w" + fmt.Sprintf("%d%02d", year, week)
	date_map["WEEK2"] = "2w" + fmt.Sprintf("%d%02d", year, week2)
	date_map["MON"] = "mo" + f[0:6]
	date_map["YEAR"] = "y" + f[0:4]

	/*
		for _, val := range date_map {
			fmt.Print("\n", val)
		}
	*/

	return date_map
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

func rtcount_gen_index(conn *gossdb.Client, table *Table, t_key *Table_Key, strs []string) (indexs []string) {
	indexs = append(indexs, "a") //"a" index for counting gloab data

	for _, indx := range t_key.Index {
		var index_str string
		for _, val := range indx.i_columnref {
			index_str += strs[val]

			//store to table column of ssdb
			s_kvkey := "set_index_" + table.Name + "_" + strconv.Itoa(val)
			if freecache.Localcache_check_and_set(s_kvkey) == false {
				conn.Zset(s_kvkey, strs[val], 1)
			}

		}
		indexs = append(indexs, index_str)
	}
	/*
		for _, val := range indexs {
			fmt.Printf("index:[%s]\n", val)
		}
	*/
	return indexs
}

func rtcount_core_table_key(conn *gossdb.Client, table *Table, t_key *Table_Key, strs []string) {

	if t_key.max_index > len(strs) {
		fmt.Printf("data error \n")
		return
	}
	//fmt.Printf("data error [%d],[%d]\n", len(strs), t_key.max_index)
	//fmt.Println(strs)

	key := strs[t_key.ikey_columnref]
	date := strs[t_key.its_columnref]

	//key cloudn't be empty
	if len(key) == 0 {
		key = "n"
	}

	date_int, err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		fmt.Printf("date error [%s]\n", date)
		return
	}

	date_map := rtcount_handle_date(date_int)

	opkey := t_key.keyopFlag

	//rtcount_correct_param(&strs)

	indexs := rtcount_gen_index(conn, table, t_key, strs)

	if opkey&COUNT == COUNT {
		rtcount_core_count(conn, table, t_key, key, date_map, indexs)
	}

	if opkey&UNION == UNION {
		rtcount_core_union(conn, table, t_key, key, date_map, indexs)
	}

	key_int, err := strconv.ParseInt(key, 10, 64)
	//change key to integer for calculating sum/max/min
	if err == nil {
		if opkey&SUM == SUM {
			rtcount_core_sum(conn, table, t_key, key_int, date_map, indexs)
		}

		if opkey&MAX == MAX {
			rtcount_core_max(conn, table, t_key, key_int, date_map, indexs)
		}

		if opkey&MIN == MIN {
			rtcount_core_min(conn, table, t_key, key_int, date_map, indexs)
		}
	}

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

func rtcount_core_count(conn *gossdb.Client, table *Table, t_key *Table_Key, key string, date_map map[string]string, indexs []string) {
	var op_key_pre string = "c_"

	for _, indx_val := range indexs {
		for _, tm_val := range t_key.Timeindex.Tm {
			kvkey := op_key_pre + table.Name + "_" + t_key.Name + "_" +
				date_map[tm_val] + "_" + indx_val
			conn.Incr(kvkey, 1)
			//fmt.Println("count--- \n", kvkey)
		}
	}

	//fmt.Printf("count \n")
}

func rtcount_core_union(conn *gossdb.Client, table *Table, t_key *Table_Key, key string, date_map map[string]string, indexs []string) {
	var i_len int = len(indexs)
	var key_set_pre string = "set_"
	var op_key_pre string

	//-----handle new op-----------------------------------------------------------------------------
	op_key_pre = "n_"
	for _, indx_val := range indexs {
		s_kvkey := key_set_pre + op_key_pre + table.Name + "_" + t_key.Name + "_" +
			date_map["ALL"] + "_" + indx_val
		//check localcace first
		if freecache.Localcache_check_and_set(s_kvkey) == true {
			//old key
			continue
		}

		if exists, err := conn.Zexists(s_kvkey, key); err != nil {
			//ssdb op error
			return
		} else if exists == true {
			//old key
			continue
		}

		//new key, insert it to set.
		conn.Zset(s_kvkey, key, 1)
		for _, tm_val := range t_key.Timeindex.Tm {
			kvkey := op_key_pre + table.Name + "_" + t_key.Name + "_" +
				date_map[tm_val] + "_" + indx_val
			conn.Incr(kvkey, 1)
			//fmt.Println("new--- \n", kvkey)
		}
	}

	//-----handle active op-----------------------------------------------------------------------------
	op_key_pre = "a_"
	for i := 0; i < i_len; i++ {
		t_len := len(t_key.Timeindex.Tm)
		for t := 0; t < t_len; t++ {
			if t_key.Timeindex.Tm[t] == "ALL" {
				continue //"ALL" don't need to cale active...
			}
			kvkey := op_key_pre + table.Name + "_" + t_key.Name + "_" +
				date_map[t_key.Timeindex.Tm[t]] + "_" + indexs[i]
			s_kvkey := key_set_pre + kvkey

			//check localcace first
			if freecache.Localcache_check_and_set(s_kvkey) == true {
				//old key
				continue
			}

			if exists, err := conn.Zexists(s_kvkey, key); err != nil {
				//ssdb op error
				return
			} else if exists == false {
				//new key, insert it to set, and incr the counter.
				//we check from big date, so remaining date must be new key.
				for l := t; l < t_len; l++ {
					kvkey := op_key_pre + table.Name + "_" + t_key.Name + "_" +
						date_map[t_key.Timeindex.Tm[t]] + "_" + indexs[i]
					s_kvkey := key_set_pre + kvkey

					conn.Zset(s_kvkey, key, 1)
					conn.Incr(kvkey, 1)
					//	fmt.Println("actvie--- \n", kvkey)
				}
				break //end time loop
			}
		} //end timeindex loop
	} //end indexs loop
}

func rtcount_core_sum(conn *gossdb.Client, table *Table, t_key *Table_Key, key_int int64, date_map map[string]string, indexs []string) {
	var op_key_pre string = "s_"

	for _, indx_val := range indexs {
		for _, tm_val := range t_key.Timeindex.Tm {
			kvkey := op_key_pre + table.Name + "_" + t_key.Name + "_" +
				date_map[tm_val] + "_" + indx_val
			conn.Incr(kvkey, key_int)
			//fmt.Println("sum--- \n", kvkey)
		}
	}

	//fmt.Printf("sum \n")
}

func rtcount_core_max(conn *gossdb.Client, table *Table, t_key *Table_Key, key_int int64, date_map map[string]string, indexs []string) {
	var op_key_pre string = "max_"

	for _, indx_val := range indexs {
		for _, tm_val := range t_key.Timeindex.Tm {
			kvkey := op_key_pre + table.Name + "_" + t_key.Name + "_" +
				date_map[tm_val] + "_" + indx_val

			if freecache.Localcache_cache_is_big(kvkey, key_int) == true {
				continue
			}

			if res, err := conn.Get(kvkey); err == nil {
				if len(res.String()) == 0 || key_int > res.Int64() {
					conn.Set(kvkey, key_int)
				}
			}
		}
	}

	//fmt.Printf("max \n")
}

func rtcount_core_min(conn *gossdb.Client, table *Table, t_key *Table_Key, key_int int64, date_map map[string]string, indexs []string) {
	var op_key_pre string = "min_"

	for _, indx_val := range indexs {
		for _, tm_val := range t_key.Timeindex.Tm {
			kvkey := op_key_pre + table.Name + "_" + t_key.Name + "_" +
				date_map[tm_val] + "_" + indx_val

			if freecache.Localcache_cache_is_small(kvkey, key_int) == true {
				continue
			}

			if res, err := conn.Get(kvkey); err == nil {
				if len(res.String()) == 0 || key_int < res.Int64() {
					conn.Set(kvkey, key_int)
				}
			}
		}
	}

	//fmt.Printf("min \n")
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

	go Rtc_StartWebServer(&g_rtc_conf)
	Rtc_StartTcpServer(&g_rtc_conf)
}
