package main

import (
	//"./seefan/gossdb"
	"bufio"
	"bytes"
	"fmt"
	"time"
	//"io"
	//"io/ioutil"
	"net/http"
	//	"time"
)

func Rtc_StartWebServer(rtc_conf *RTC_Conf) {
	fmt.Println("HttpServer start!")

	http.HandleFunc("/query", query)
	for _, table := range rtc_conf.Table {
		http.HandleFunc("/table/"+table.Name, httpdata)
	}

	http.HandleFunc("/info", info)

	http.HandleFunc("/table/test", httpdata)
	http.HandleFunc("/table/ddd", httpdata)

	err := http.ListenAndServe(":"+rtc_conf.Port, nil)

	if err != nil {
		fmt.Println("ListenAndServe error: ", err.Error())
	}
}

func info(w http.ResponseWriter, req *http.Request) {

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

	/*
	   table: table_name
	   key: key_name  ALL OP num: count()
	   key: key_name  ALL OP num: count()
	*/
	var rtc_server_info string
	for _, table := range g_rtc_conf.Table {
		rtc_server_info += "Table: " + table.Name + "<br>"
		for _, t_key := range table.Keys {

			rtc_server_info += "KEY: " + t_key.Name + "<br>"
			opkey := t_key.keyopFlag
			if opkey&COUNT == COUNT {
				kvkey := "c_" + table.Name + "_" + t_key.Name + "_a_a"
				if res, err := conn.Get(kvkey); err == nil {
					rtc_server_info += " COUNT[ " + res.String() + " ]"
				} else {
					rtc_server_info += " COUNT[ 0 ]"
				}
			}

			if opkey&UNION == UNION {
				kvkey := "n_" + table.Name + "_" + t_key.Name + "_a_a"
				if res, err := conn.Get(kvkey); err == nil {
					rtc_server_info += " UNION[ " + res.String() + " ]"
				} else {
					rtc_server_info += " UNION[ 0 ]"
				}
			}

			if opkey&SUM == SUM {
				kvkey := "s_" + table.Name + "_" + t_key.Name + "_a_a"
				if res, err := conn.Get(kvkey); err == nil {
					rtc_server_info += " SUM[ " + res.String() + " ]"
				} else {
					rtc_server_info += " SUM[ 0 ]"
				}
			}

			if opkey&MAX == MAX {
				kvkey := "max_" + table.Name + "_" + t_key.Name + "_a_a"
				if res, err := conn.Get(kvkey); err == nil {
					rtc_server_info += " MAX[ " + res.String() + " ]"
				} else {
					rtc_server_info += " MAX[ 0 ]"
				}
			}

			if opkey&MIN == MIN {
				kvkey := "min_" + table.Name + "_" + t_key.Name + "_a_a"
				if res, err := conn.Get(kvkey); err == nil {
					rtc_server_info += " MIN[ " + res.String() + " ]"
				} else {
					rtc_server_info += " MIN[ 0 ]"
				}
			}
			rtc_server_info += " <br>"
		}
	}

	fmt.Fprint(w, rtc_server_info)
}

func query(w http.ResponseWriter, req *http.Request) {
	//fmt.Println(req)

	if req.ContentLength == 0 {
		fmt.Fprint(w, "no query data")
		return
	}

	fmt.Fprint(w, "ok")
}

func s_byteString(p [][]byte) (strs []string) {
	for i := 0; i < len(p); i++ {
		strs = append(strs, p_byteString(p[i]))
	}
	return strs
}

func p_byteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

var cou int = 0

func WebDataHandle(tablename string, line []byte) {

	if tablename == "chukong_game" {
		CK_handle_log(tablename, line)
		return
	}

	cou++
	//fmt.Printf("----[%d]------------\n", cou)
	//fmt.Printf("old:%s\n", p_byteString(line))

	//"\x02"
	xx := bytes.Split(line, []byte("\x02"))
	//xx := bytes.Split(line, []byte(","))
	strs := s_byteString(xx)

	//fmt.Println(strs)

	//fmt.Printf("----[%d]----len[%d]--------\n", cou, len(strs))

	//rtcount_handle_talbe(tablename, strs)

	rtcount_before(tablename, strs)
}

func httpdata(w http.ResponseWriter, req *http.Request) {
	//fmt.Println(req)

	if req.ContentLength == 0 {
		fmt.Fprint(w, "no data upload")
		return
	}

	/*
		if req.Method != "POST" {
			fmt.Fprint(w, "Only support POST")
			return
		}
	*/
	url := req.RequestURI

	//7 == len("/table/")
	tablename := url[7:len(url)]
	fmt.Println(tablename)
	fmt.Println(req.RequestURI)

	start := time.Now().Unix()

	scanner := bufio.NewScanner(req.Body)
	var linenum int = 0
	for scanner.Scan() {
		WebDataHandle(tablename, scanner.Bytes())
		linenum++
		//scanner.Text()
	}

	err := scanner.Err()
	if err != nil {
		fmt.Println("scanner err:\n", err)
	}
	//result, _ := ioutil.ReadAll(req.Body)

	req.Body.Close()
	//fmt.Printf("%s\n", result)

	end := time.Now().Unix()

	fmt.Printf("handle url[%s],lines[%d], using[%d]s\n", url, linenum, (end - start))

	fmt.Fprint(w, "ok")
}

/*
func main() {

	Rtc_StartWebServer()
	fmt.Println("loginTask is running...")
}
*/
