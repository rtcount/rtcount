package main

import (
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
	//func Rtc_StartWebServer() {
	fmt.Println("HttpServer start!")

	//第一个参数为客户端发起http请求时的接口名，第二个参数是一个func，负责处理这个请求。
	http.HandleFunc("/query", query)
	for _, table := range rtc_conf.Table {
		http.HandleFunc("/table/"+table.Name, httpdata)
	}

	//			err := http.ListenAndServe(":"+rtc_conf, nil)

	http.HandleFunc("/table/test", httpdata)
	http.HandleFunc("/table/ddd", httpdata)

	err := http.ListenAndServe(":"+rtc_conf.Port, nil)
	//err := http.ListenAndServe(":9999", nil)

	if err != nil {
		fmt.Println("ListenAndServe error: ", err.Error())
	}
}

func query(w http.ResponseWriter, req *http.Request) {
	//fmt.Println(req)

	fmt.Println(req.URL)

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
