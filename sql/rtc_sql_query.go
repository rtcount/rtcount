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
	//"sort"
	//	"time"
)

func test_pars() {
	ori_sql := "select SUM from demo.demo_mini with MIN and index_1 WHERE clm1 =clm1 and 1clm3 =c1lm3 and time <=123 and time >= zxc and  clm2 =clm2;;  ;"
	//过滤 ';'
	ori_sql = strings.Replace(ori_sql, ";", "", -1)

	//过滤前后的空格
	ori_sql = strings.TrimSpace(ori_sql)
	//全部转换为小写
	sql := strings.ToLower(ori_sql)

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
	ret, OP := getSubStringBw(sql, "select", "from")
	if ret == false {
		fmt.Println("OP error", OP)
		return
	}
	//fmt.Println(OP)

	ret, TABLE_KEY := getSubStringBw(sql, "from", "with")
	if ret == false {
		fmt.Println("TABLE error", TABLE_KEY)
		return
	}

	ret, TIME_INDEX := getSubStringBw(sql, "with", "where")
	if ret == false {
		fmt.Println("WITH error", TIME_INDEX)
		return
	}

	where := strings.Split(sql, " where ")[1]
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
	strs := strings.Split(TIME_INDEX, " and ")

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
		if indx.Name == *INDEX {
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

func checkTableKey(TABLE_KEY string, TABLE *string, KEY *string, t_key **Table_Key, table **Table) (bool, string) {
	strs := strings.Split(TABLE_KEY, ".")
	if len(strs) != 2 {
		return false, "The TABLE_KET[" + TABLE_KEY + "] not error"
	}
	*TABLE = strs[0]
	*KEY = strs[1]

	match := false
	//查找配置文件里是否有指定的表名
	for _, t_val := range g_rtc_conf.Table {
		if t_val.Name == *TABLE {
			match = true
			*table = &t_val
			break
		}
	}

	if match == false {
		return false, "The TABLE[" + *TABLE + "] did not define in your xml config"
	}

	match = false
	//查找表中是否有这个KEY
	for _, key_v := range (*table).Keys {
		if key_v.Name == *KEY {
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

type whereItem struct {
	name        string
	op          string
	value       string
	i_columnref int
}

func checkWhere(where string, i_Index *Index, table *Table) (bool, string) {

	and := strings.Split(where, " and ")
	fmt.Println(and)

	//key_value := make(map[string]string)
	var whereList []whereItem
	var whereTimeList []whereItem

	for _, val := range and {
		var myExp = myRegexp{regexp.MustCompile(`(?P<k>.*)(?P<op>>=|<=|[^<>]=|[^=]>|[^=]<)(?P<v>.*)`)}
		mmap := myExp.FindStringSubmatchMap(val)

		var item whereItem
		item.name = strings.TrimSpace(mmap["k"])
		item.op = strings.TrimSpace(mmap["op"])
		item.value = strings.TrimSpace(mmap["v"])
		item.i_columnref = -1
		if item.name == "time" {
			whereTimeList = append(whereTimeList, item)
		} else {
			whereList = append(whereList, item)
		}
	}

	fmt.Println(whereList)
	fmt.Println(whereTimeList)

	if len(whereTimeList) > 2 {
		return false, "Two many TIME near where"
	}

	l_len := len(whereList)
	//过滤索引列和操作值
	for i := 0; i < l_len; i++ {
		index_column := GetIndexInArrayByString(whereList[i].name, i_Index.Columnref)
		if index_column == -1 {
			return false, "There is a error with [" + whereList[i].name + "] near where "
		}
		whereList[i].i_columnref = GetIndexInArrayByString(whereList[i].name, table.Column)

		if whereList[i].op != "=" {
			return false, "The index column [" + whereList[i].name + "] just support '=' OP near where "
		}
	}

	//检测索引列是否重复

	var whereNameList []string
	for _, val := range whereList {
		if GetIndexInArrayByString(val.name, whereNameList) != -1 {
			return false, "The index column [" + val.name + "] repeat near where "
		}
		whereNameList = append(whereNameList, val.name)
	}

	//检测索引列数目缺少的
	for _, column := range i_Index.Columnref {
		if GetIndexInArrayByString(column, whereNameList) == -1 {
			return false, "Miss index column [" + column + "] in where "
		}
	}

	//产生index索引key
	var index_str string
	//indx.i_columnref已经被sort过了，所以这里使用range，其顺序是固定的
	for _, val := range i_Index.i_columnref {
		fmt.Println("\n", val, "\n")
		for _, item := range whereList {
			if val == item.i_columnref {
				index_str += item.value
				fmt.Println("\n", item.value, "\n")
			}
		}
	}
	fmt.Println(whereList)
	fmt.Println(index_str)

	//fmt.Println(key_value)

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
	var table *Table
	var i_Index *Index
	var t_key *Table_Key

	ret, err_msg := checkTableKey(TABLE_KEY, &TABLE, &KEY, &t_key, &table)
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

	ret, err_msg = checkWhere(where, i_Index, table)
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
