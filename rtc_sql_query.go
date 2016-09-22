package main

import (
	// #cgo LDFLAGS: -L${SRCDIR}/ -L./lib -lparser
	/*
		#include <stdlib.h>
		#include "sql/api.h"
	*/
	"C"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"unsafe"
)

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

func checkTimeIndex(with []string, TIME *string, INDEX *string, i_Index **Index, t_key *Table_Key) (bool, string) {

	if GetIndexInArrayByString(with[0], TIMEINDEXS) == -1 {
		if GetIndexInArrayByString(with[1], TIMEINDEXS) == -1 {
			return false, "Don't set TIME in your SQL"
		} else {
			*TIME = with[1]
			*INDEX = with[0]
		}
	} else {
		*TIME = with[0]
		*INDEX = with[1]
	}

	//检测TIME是否在当前KEY里设置了
	if GetIndexInArrayByString(*TIME, t_key.Timeindex.Tm) == -1 {
		return false, "The TIME[" + *TIME + "] did not define in key[" + t_key.Name + "]"
	}

	//检测索引是否在当前KEY里设置了
	match := false
	for _, indx := range t_key.Index {

		//	fmt.Println("\n", indx.Name, "\n")
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

func checkTableKey(TABLE string, KEY string, t_key **Table_Key, table **Table) (bool, string) {

	match := false
	//查找配置文件里是否有指定的表名
	for _, t_val := range g_rtc_conf.Table {
		if t_val.Name == TABLE {
			match = true
			*table = &t_val
			break
		}
	}

	if match == false {
		return false, "The TABLE[" + TABLE + "] did not define in your xml config"
	}

	match = false
	//查找表中是否有这个KEY
	for _, key_v := range (*table).Keys {
		if key_v.Name == KEY {
			match = true
			*t_key = &key_v
			break
		}
	}
	if match == false {
		return false, "The KEY[" + KEY + "] did not define in table[" + TABLE + "]"
	}

	return true, "OK"
}

func rtcount_gen_time(timestmp int64, TimeIndex string) (string, string) {

	date_map := make(map[string]string)
	time_map := make(map[string]string)
	//var dates []string

	tm := time.Unix(timestmp, 0)
	f := tm.Format("200601021504")

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

	date_map["all"] = "a"
	date_map["min"] = "m" + f[0:12]
	date_map["min5"] = "5m" + f[0:11] + min5last
	date_map["min10"] = "1m" + f[0:11] + "0"
	date_map["min30"] = "3m" + f[0:10] + min30last
	date_map["hour"] = "h" + f[0:10]
	date_map["day"] = "d" + f[0:8]
	date_map["week"] = "w" + fmt.Sprintf("%d%02d", year, week)
	date_map["week2"] = "2w" + fmt.Sprintf("%d%02d", year, week2)
	date_map["mon"] = "mo" + f[0:6]
	date_map["year"] = "y" + f[0:4]

	time_map["all"] = "a"
	time_map["min"] = f[0:12]
	time_map["min5"] = f[0:11] + min5last
	time_map["min10"] = f[0:11] + "0"
	time_map["min30"] = f[0:10] + min30last
	time_map["hour"] = f[0:10]
	time_map["day"] = f[0:8]
	time_map["week"] = fmt.Sprintf("%d%02d", year, week)
	time_map["week2"] = fmt.Sprintf("%d%02d", year, week2)
	time_map["mon"] = f[0:6]
	time_map["year"] = f[0:4]

	/*
			for _, val := range date_map {
				fmt.Print("\n", val)
			}

		t_len := len(t_key.Timeindex.Tm)
		for i := 0; i < t_len; i++ {
			//fmt.Print("\n", date_map[t_key.Timeindex.Tm[i]])
			dates = append(dates, date_map[t_key.Timeindex.Tm[i]])
		}
	*/

	//fmt.Print("------------", date_map[TimeIndex], time_map[TimeIndex])
	return date_map[TimeIndex], time_map[TimeIndex]
}

var time2Op = []string{"<", "<=", ">", ">="}
var TIMEINDEXS_Step = map[string]int64{
	"min":   60,
	"min5":  5 * 60,
	"min10": 10 * 60,
	"min30": 30 * 60,
	"hour":  60 * 60,
	"day":   24 * 60 * 60,
	"week":  7 * 24 * 60 * 60,
	"week2": 14 * 24 * 60 * 60,
	"mon":   30 * 24 * 60 * 60,
	"year":  365 * 24 * 60 * 60}

func rtcount_gen_arraytime(L_OP string, L_int int64, G_OP string, G_int int64, TimeOP string) map[string]string {
	//fmt.Println("run 0--- rtcount_gen_arraytime")
	maptime := make(map[string]string)
	var step = TIMEINDEXS_Step[TimeOP]
	var start_time, end_time int64
	start_time = G_int
	end_time = L_int

	if L_OP == "<" {
		end_time -= 1
	}

	if G_OP == ">" {
		start_time += 1
	}

	//fmt.Println(L_OP, start_time, G_OP, end_time, step)

	for i := start_time; i <= end_time; i = i + step {
		time_index, time_1 := rtcount_gen_time(i, TimeOP)
		//fmt.Println("rtcount_gen_arraytime:", time_index, time_1)
		maptime[time_1] = time_index
	}

	//fmt.Println(maptime)
	return maptime
}

func checkTimeWhere(whereTimeList []condition, TimeOP string, timeIndex *map[string]string) (bool, string) {

	//fmt.Println("run 0--- whereTimeList")

	num_tmlist := len(whereTimeList)
	if num_tmlist > 2 {
		return false, "Too mant time index column near where "
	}

	if num_tmlist == 0 {
		//*timeIndex = append(*timeIndex, "a")
		(*timeIndex)["a"] = "a"
		return true, "OK"
	}

	if num_tmlist == 1 {
		if whereTimeList[0].Op != "=" {
			return false, "The time index column just support '=' OP near where "
		}

		date_int, err := strconv.ParseInt(whereTimeList[0].Value, 10, 64)
		if err != nil {
			return false, "time is value error near where"
		}

		time_index, time_1 := rtcount_gen_time(date_int, TimeOP)

		(*timeIndex)[time_1] = time_index
		//*timeIndex = append(*timeIndex, time_index)
	}

	if num_tmlist == 2 {
		var Le, Ge int
		var Le_int, Ge_int int64
		var LeOp, GeOp string
		i0 := GetIndexInArrayByString(whereTimeList[0].Op, time2Op)
		i1 := GetIndexInArrayByString(whereTimeList[1].Op, time2Op)

		tml0_int, err0 := strconv.ParseInt(whereTimeList[0].Value, 10, 64)
		if err0 != nil {
			return false, "time is value error near where"
		}

		tml1_int, err1 := strconv.ParseInt(whereTimeList[1].Value, 10, 64)
		if err1 != nil {
			return false, "time is value error near where"
		}
		if i0 == -1 || i1 == -1 || i0 == i1 {
			return false, "time index error near where"
		}
		if i0 > i1 {
			Ge = i0
			Ge_int = tml0_int
			Le = i1
			Le_int = tml1_int
		} else {
			Ge = i1
			Ge_int = tml1_int
			Le = i0
			Le_int = tml0_int
		}

		if Ge-Le == 1 {
			if (Le == 0 && Ge == 1) || (Le == 2 && Ge == 3) {
				return false, "time index error near where"
			}
		}

		GeOp = time2Op[Ge]
		LeOp = time2Op[Le]

		*timeIndex = rtcount_gen_arraytime(LeOp, Le_int, GeOp, Ge_int, TimeOP)

	}

	return true, "OK"
}

func checkWhere(where []condition, TimeIndex string, i_Index *Index, table *Table, query_Index *map[string]string) (bool, string) {

	var whereList []condition
	var whereTimeList []condition

	for _, item := range where {
		if item.LhsAttr == "time" {
			whereTimeList = append(whereTimeList, item)
		} else {
			whereList = append(whereList, item)
		}
	}

	if len(whereTimeList) == 0 && len(whereList) == 0 {
		(*query_Index)["a"] = "a_a"
		return true, "OK"
	}

	//fmt.Println(whereList)
	//fmt.Println(whereTimeList)
	//var time_index string
	var index_str string
	var timeIndex map[string]string

	checkTimeWhere(whereTimeList, TimeIndex, &timeIndex)
	//fmt.Println(timeIndex)

	l_len := len(whereList)
	if l_len == 0 {
		index_str = "a"
	} else {
		//过滤索引列和操作值
		for i := 0; i < l_len; i++ {
			index_column := GetIndexInArrayByString(whereList[i].LhsAttr, i_Index.Columnref)
			if index_column == -1 {
				return false, "There is a error with [" + whereList[i].LhsAttr + "] near where "
			}
			whereList[i].i_columnref = GetIndexInArrayByString(whereList[i].LhsAttr, table.Column)

			//if whereList[i].Op != "=" {
			if whereList[i].Op != "=" {
				return false, "The index column [" + whereList[i].LhsAttr + "] just support '=' OP near where "
			}
		}

		//检测索引列是否重复

		var whereNameList []string
		for _, val := range whereList {
			if GetIndexInArrayByString(val.LhsAttr, whereNameList) != -1 {
				return false, "The index column [" + val.LhsAttr + "] repeat near where "
			}
			whereNameList = append(whereNameList, val.LhsAttr)
		}

		//检测索引列数目缺少的
		for _, column := range i_Index.Columnref {
			if GetIndexInArrayByString(column, whereNameList) == -1 {
				return false, "Miss index column [" + column + "] in where "
			}
		}

		//产生index索引key
		//indx.i_columnref已经被sort过了，所以这里使用range，其顺序是固定的
		for _, val := range i_Index.i_columnref {
			//fmt.Println("\n", val, "\n")
			for _, item := range whereList {
				if val == item.i_columnref {
					index_str += item.Value
					//		fmt.Println("\n", item.Value, "\n")
				}
			}
		}
	}

	//fmt.Println(whereList)
	//fmt.Println(time_index + "_" + index_str)

	for idx, val := range timeIndex {
		(*query_Index)[idx] = val + "_" + index_str
	}

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

func checkSQL(sql RTC_Sql, query_index *map[string]string) (bool, string) {
	var TIME, INDEX string
	var table *Table
	var i_Index *Index
	var t_key *Table_Key

	//fmt.Println(sql)

	ret, err_msg := checkTableKey(sql.Table, sql.Key, &t_key, &table)
	if ret == false {
		return ret, err_msg
	}

	ret, err_msg = checkOP(sql.Op, t_key)
	if ret == false {
		return ret, err_msg
	}

	ret, err_msg = checkTimeIndex(sql.With, &TIME, &INDEX, &i_Index, t_key)
	if ret == false {
		return ret, err_msg
	}

	ret, err_msg = checkWhere(sql.Condis, TIME, i_Index, table, query_index)
	if ret == false {
		return ret, err_msg
	}

	//fmt.Println(sql.Op, sql.Table, sql.Key, query_index)
	sql_gen_key(sql.Op, sql.Table, sql.Key, query_index)
	//fmt.Println(kv)
	/*
		for _, val := range query_index {
			fmt.Println(val)
			return true, val
		}
	*/

	return true, "OK"

}

func sql_gen_key(OP string, Table string, Key string, query_index *map[string]string) map[string]string {
	var op_key_pre string = OP_KEY[OP]

	for idx, val := range *query_index {
		(*query_index)[idx] = op_key_pre + "_" + Table + "_" + Key + "_" + val
	}

	return (*query_index)
}

func sql_query(query string) string {
	csql := C.CString(query)
	defer C.free(unsafe.Pointer(csql))
	cc := C.GoString(C.ddd(csql))

	if cc == "" {
		return "{\"code\":-1, \"msg\":\"SQL Syntax Error\"}"
	}

	//fmt.Println(cc)

	query_index := make(map[string]string)

	sql, msg, ret := RTC_sql_check(cc)

	if ret == false {
		fmt.Println(sql, msg, ret)
		return "{\"code\":-1, \"msg\":\"" + msg + "\"}"
	}

	ret, msg = checkSQL(sql, &query_index)

	if ret == false {
		return "{\"code\":-1, \"msg\":\"" + msg + "\"}"
	}

	conn, err := dbpoll.NewClient()
	if err != nil {
		time.Sleep(1)
		conn, err = dbpoll.NewClient()
		if err != nil {
			fmt.Println("Failed to create new client:", err)
			return "{\"code\":-1, \"msg\":\"INEL ERROR\"}"
		}
	}
	defer conn.Close()

	for idx, val := range query_index {
		//fmt.Println("ssdb:", val)
		if res, err := conn.Get(val); err == nil {
			if res.String() == "" {
				query_index[idx] = "0"
			} else {
				query_index[idx] = res.String()
			}
		}
	}
	/*
		if res, err := conn.Get(kv); err == nil {
			return "kv-----[" + res.String() + "]"
		}
	*/

	b, _ := json.Marshal(query_index)
	//fmt.Println(string(b))
	//fmt.Println(cret, kv, query_index)
	return string(b)
}

/*
func test_pars() {

	str1 := "select SUm from DEMO.DEMO_mini with mIN and INDEx_1 where time = 1466252795 and clm1='xxx' and clm3='iyy' and clm2='zz';"
	cstr := C.CString(str1)

	fmt.Println(str1)
	//C.hello()
	cc := C.GoString(C.ddd(cstr))
	fmt.Println(cc)

	sql, msg, ret := RTC_sql_check(cc)

	fmt.Println(sql, msg, ret)

	cret, emsg := checkSQL(sql)
	fmt.Println(cret, emsg)

	C.free(unsafe.Pointer(cstr))
}
*/
