package main

import (
	//"./seefan/gossdb"
	//"bufio"
	//"bytes"
	"fmt"
	"strings"
	//"io"
	//"io/ioutil"
	"regexp"
	//	"time"
)

func test_pars() {
	sql := "select SUM from demo.demo_mini with MIN and index_1 where INDEX_Colu =xxx and time <=123 and time >= zxc and time > zxc;;  ;"
	//过滤 ';'
	sql = strings.Replace(sql, ";", "", -1)
	//过滤前后的空格
	sql = strings.TrimSpace(sql)
	//全部转换为大写
	sql = strings.ToUpper(sql)

	/*
			var myExp = myRegexp{regexp.MustCompile(`SELECT (?P<op>.*) FROM (?P<table_key>.*) WITH (?P<time_index>.*) WHERE (?P<where>.*)`)}
			mmap := myExp.FindStringSubmatchMap(sql)

			ww := mmap["op"]
			wm := mmap["table_key"]
			w1 := mmap["time_index"]
			w2 := mmap["where"]
			fmt.Println(mmap)

		fmt.Println(ww)
		fmt.Println(wm)
		fmt.Println(w1)
		fmt.Println(w2)
	*/

	//fmt.Println(sql)
	ret, OP := getSubStringBw(sql, "SELECT", "FROM")
	if ret == false {
		fmt.Println("OP error", OP)
		return
	}
	//fmt.Println(OP)

	ret, TABLE_KEY := getSubStringBw(sql, "FROM", "WITH")
	if ret == false {
		fmt.Println("TABLE error", TABLE_KEY)
		return
	}

	ret, TIME_INDEX := getSubStringBw(sql, "WITH", "WHERE")
	if ret == false {
		fmt.Println("WITH error", TIME_INDEX)
		return
	}

	where := strings.Split(sql, " WHERE ")[1]
	//	fmt.Println(where)

	ret, err_msg := checkParam(OP, TABLE_KEY, TIME_INDEX, where)
	if ret == false {

		fmt.Println("errmse", err_msg)
	}

}

func checkOP(OP string, t_key *Table_Key) (bool, string) {

	i := GetIndexInArrayByString(OP, KEYOPS)
	if i == -1 {
		return false, "The OP[" + OP + "] not support "
	}
	i = GetIndexInArrayByString(OP, t_key.KeyOP.Op)
	if i == -1 {
		return false, "The OP[" + OP + "] not support in KEY[" + t_key.Name + "]"
	}

	return true, "OK"
}

func checkTimeIndex(TIME_INDEX string, TIME *string, INDEX *string, i_Index **Index, t_key *Table_Key) (bool, string) {
	strs := strings.Split(TIME_INDEX, " AND ")

	if GetIndexInArrayByString(strs[0], TIMEINDEXS) == -1 {
		if GetIndexInArrayByString(strs[1], TIMEINDEXS) == -1 {
			return false, "Don't set TIME in your SQL"
		} else {
			*TIME = strs[1]
			*INDEX = strs[0]
		}
	} else {
		*TIME = strs[0]
		*INDEX = strs[1]
	}

	//检测TIME是否在当前KEY里设置了
	if GetIndexInArrayByString(*TIME, t_key.Timeindex.Tm) == -1 {
		return false, "The TIME[" + *TIME + "] did not define in key[" + t_key.Name + "]"
	}

	//检测索引是否在当前KEY里设置了
	match := false
	for _, indx := range t_key.Index {
		if strings.ToUpper(indx.Name) == *INDEX {
			match = true
			*i_Index = &indx
			break
		}
	}
	if match == false {
		return false, "The INDEX[" + *INDEX + "] did not define in key[" + t_key.Name + "]"
	}
	return true, "OK"
}

func checkTableKey(TABLE_KEY string, TABLE *string, KEY *string, t_key **Table_Key) (bool, string) {
	strs := strings.Split(TABLE_KEY, ".")
	if len(strs) != 2 {
		return false, "The TABLE_KET[" + TABLE_KEY + "] not error"
	}
	*TABLE = strs[0]
	*KEY = strs[1]

	var table *Table
	match := false
	//查找配置文件里是否有指定的表名
	for _, t_val := range g_rtc_conf.Table {
		if strings.ToUpper(t_val.Name) == *TABLE {
			match = true
			table = &t_val
			break
		}
	}

	if match == false {
		return false, "The TABLE[" + *TABLE + "] did not define in your xml config"
	}

	match = false
	//查找表中是否有这个KEY
	for _, key_v := range table.Keys {
		if strings.ToUpper(key_v.Name) == *KEY {
			match = true
			*t_key = &key_v
			break
		}
	}
	if match == false {
		return false, "The KEY[" + *KEY + "] did not define in table[" + *TABLE + "]"
	}

	return true, "OK"
}

//embed regexp.Regexp in a new type so we can extend it
type myRegexp struct {
	*regexp.Regexp
}

//add a new method to our new regular expression type
func (r *myRegexp) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.SubexpNames() {
		//Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		captures[name] = match[i]

	}
	return captures
}

func checkWhere(where string, i_Index *Index) (bool, string) {

	and := strings.Split(where, " AND ")
	fmt.Println(and)

	for _, val := range and {
		var myExp = myRegexp{regexp.MustCompile(`(?P<k>.*)(>=|<=|[^<>]=|[^=]>|[^=]<)(?P<v>.*)`)}
		mmap := myExp.FindStringSubmatchMap(val)
		ww := mmap["k"]
		wm := mmap["v"]

		//fmt.Println(mmap)
		fmt.Println(ww)
		fmt.Println(wm)
	}

	/*
		var TIME1, TIME2 string
		for _, val := range where {
			if val == item {
				return index
			}
		}
	*/

	return true, "OK"
}

func checkParam(OP string, TABLE_KEY string, TIME_INDEX string, where string) (bool, string) {

	var TABLE, KEY, TIME, INDEX string
	var i_Index *Index
	var t_key *Table_Key

	ret, err_msg := checkTableKey(TABLE_KEY, &TABLE, &KEY, &t_key)
	if ret == false {
		return ret, err_msg
	}

	ret, err_msg = checkOP(OP, t_key)
	if ret == false {
		return ret, err_msg
	}

	ret, err_msg = checkTimeIndex(TIME_INDEX, &TIME, &INDEX, &i_Index, t_key)
	if ret == false {
		return ret, err_msg
	}

	ret, err_msg = checkWhere(where, i_Index)
	if ret == false {
		return ret, err_msg
	}

	fmt.Println(OP, TABLE, KEY, TIME, INDEX)
	fmt.Println(where)

	return true, "OK"
}

func getSubStringBw(ori string, beigin string, end string) (bool, string) {

	sqlRegexp := regexp.MustCompile(`(?i:` + beigin + ` ).*`)
	tmp := sqlRegexp.FindString(ori)
	if tmp == "" {
		return false, beigin + "error"
	}

	sqlRegexp = regexp.MustCompile(`(?i:` + beigin + ` ).*(?i: ` + end + ` )`)
	tmp = sqlRegexp.FindString(ori)
	if tmp == "" {
		return false, end + "error"
	}

	ret := tmp
	ret = strings.Replace(ret, beigin, "", -1)
	ret = strings.Replace(ret, end, "", -1)
	ret = strings.TrimSpace(ret)

	return true, ret
}
