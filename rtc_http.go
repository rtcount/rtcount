package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func Rtc_StartWebServer(rtc_conf *RTC_Conf) {
	fmt.Println("HttpServer start!")

	http.HandleFunc("/query", query)
	for _, table := range rtc_conf.Table {
		http.HandleFunc("/table/"+table.Name, httpdata)
	}

	http.HandleFunc("/info", info)

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
		   table: table_name key: key_name
		    INDEX: index_name
			ALL OP num: count()
		    ALL OP num: count()
	*/
	var rtc_server_info string

	for _, table := range g_rtc_conf.Table {
		rtc_server_info += "Table: " + table.Name
		for _, t_key := range table.Keys {

			rtc_server_info += " KEY: " + t_key.Name + "<br>"
			opkey := t_key.keyopFlag

			for _, indx := range t_key.Index {
				rtc_server_info += "INDEX: " + indx.Name + "<br>"
				if opkey&COUNT == COUNT {
					//select COUNT from table.Name.t_key.Name with TIME and INDEX
					Sql := "select COUNT from " + table.Name + "." + t_key.Name + " with " + indx.Name + " and " + t_key.Timeindex.Tm[0] + ";"
					ret := sql_query(Sql)
					rtc_server_info += " COUNT[ " + ret + " ]"
				}

				if opkey&NEW == NEW {
					//select NEW from table.Name.t_key.Name with TIME and INDEX
					Sql := "select NEW from " + table.Name + "." + t_key.Name + " with " + indx.Name + " and " + t_key.Timeindex.Tm[0] + ";"
					ret := sql_query(Sql)
					rtc_server_info += " NEW[ " + ret + " ]"
				}

				if opkey&SUM == SUM {
					//select SUM from table.Name.t_key.Name with TIME and INDEX
					Sql := "select SUM from " + table.Name + "." + t_key.Name + " with " + indx.Name + " and " + t_key.Timeindex.Tm[0] + ";"
					ret := sql_query(Sql)
					rtc_server_info += " SUM[ " + ret + " ]"
				}

				if opkey&MAX == MAX {
					//select MAX from table.Name.t_key.Name with TIME and INDEX
					Sql := "select MAX from " + table.Name + "." + t_key.Name + " with " + indx.Name + " and " + t_key.Timeindex.Tm[0] + ";"
					ret := sql_query(Sql)
					rtc_server_info += " MAX[ " + ret + " ]"
				}

				if opkey&MIN == MIN {
					//select MIN from table.Name.t_key.Name with TIME and INDEX
					Sql := "select MIN from " + table.Name + "." + t_key.Name + " with " + indx.Name + " and " + t_key.Timeindex.Tm[0] + ";"
					ret := sql_query(Sql)
					rtc_server_info += " MIN[ " + ret + " ]"
				}
				rtc_server_info += " <br>"
			}
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

	query, _ := ioutil.ReadAll(req.Body)
	ret := sql_query(p_byteString(query))

	fmt.Fprint(w, ret)
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
	cou++

	if tablename == "chukong_game" {
		CK_handle_log(tablename, line)
		return
	}

	xx := bytes.Split(line, []byte("\x02"))
	strs := s_byteString(xx)

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

	tablename := url[7:len(url)]

	//start := time.Now().Unix()

	scanner := bufio.NewScanner(req.Body)
	var linenum int = 0
	for scanner.Scan() {
		WebDataHandle(tablename, scanner.Bytes())
		linenum++
	}

	err := scanner.Err()
	if err != nil {
		fmt.Println("scanner err:\n", err)
	}

	req.Body.Close()
	//end := time.Now().Unix()

	//fmt.Printf("handle url[%s],lines[%d], using[%d]s\n", url, linenum, (end - start))
	fmt.Fprint(w, "ok")
}
