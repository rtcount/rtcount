package main

import (
	"bufio"
	"fmt"
	//"io"
	//"io/ioutil"
	"net/http"
	//	"time"
)

//func Rtc_StartWebServer(rtc_conf *RTC_Conf) {
func Rtc_StartWebServer() {
	fmt.Println("HttpServer start!")

	//第一个参数为客户端发起http请求时的接口名，第二个参数是一个func，负责处理这个请求。
	/*
		http.HandleFunc("query", query)
					for _, table := range rtc_conf {
						http.HandleFunc("/table/"+table.Name, httpdata)
					}

			err := http.ListenAndServe(":"+rtc_conf, nil)
	*/

	http.HandleFunc("/table/test", httpdata)
	http.HandleFunc("/table/ddd", httpdata)

	//服务器要监听的主机地址和端口号
	err := http.ListenAndServe(":9999", nil)

	if err != nil {
		fmt.Println("ListenAndServe error: ", err.Error())
	}
}

func query(w http.ResponseWriter, req *http.Request) {
	//fmt.Println(req)

	fmt.Println(req.URL)

	fmt.Fprint(w, "ok")
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
	//rtcount_before(line)

	cou++
	fmt.Printf("----[%d]------------\n", cou)
	fmt.Printf("%s\n", p_byteString(line))
	rtcount_before(tablename, p)
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

	scanner := bufio.NewScanner(req.Body)
	for scanner.Scan() {
		WebDataHandle(tablename, scanner.Bytes())
		//scanner.Text()
	}

	err := scanner.Err()
	if err != nil {
		fmt.Println("scanner err:\n", err)
	}
	//result, _ := ioutil.ReadAll(req.Body)

	req.Body.Close()
	//fmt.Printf("%s\n", result)

	fmt.Fprint(w, "ok")
}

func main() {

	arg_num := len(os.Args)
	fmt.Printf("the num of input is %d\n", arg_num)

	fmt.Printf("they are :\n")
	for i := 0; i < arg_num; i++ {
		fmt.Println(os.Args[i])
	}

	Rtc_StartWebServer()
	fmt.Println("loginTask is running...")
}
