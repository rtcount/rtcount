package main

import (
	"./freecache"
	"./seefan/gossdb"
	"fmt"
	"strconv"
	"time"
)

const (
	/* opkey: maxmin|sum|union|count */
	COUNT  = 1 << 0
	NEW    = 1 << 1
	ACTIVE = 1 << 2
	SUM    = 1 << 3
	MAX    = 1 << 4
	MIN    = 1 << 5
)

/* setting ssdb key pre values */
const (
	PRE_KEYSET       = "set_"
	PRE_COUNT_KEYOP  = "cot_"
	PRE_NEW_KEYOP    = "new_"
	PRE_ACTIVE_KEYOP = "act_"
	PRE_SUM_KEYOP    = "sum_"
	PRE_MAX_KEYOP    = "max_"
	PRE_MIN_KEYOP    = "min_"
	PRE_INDEX_KEYOP  = "index_"
)

var OP_KEY = map[string]string{"set": "set_", "count": "cot_", "new": "new_", "active": "act_", "sum": "sum_", "max": "max_", "min": "min_", "index": "index_"}

func rtcount_gen_dates(timestmp int64, t_key *Table_Key) []string {

	date_map := make(map[string]string)
	var dates []string

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

	/*
		for _, val := range date_map {
			fmt.Print("\n", val)
		}
	*/

	t_len := len(t_key.Timeindex.Tm)
	for i := 0; i < t_len; i++ {
		//fmt.Print("\n", date_map[t_key.Timeindex.Tm[i]])
		dates = append(dates, date_map[t_key.Timeindex.Tm[i]])
	}

	return dates
}

func rtcount_gen_indexs(conn *gossdb.Client, table *Table, t_key *Table_Key, strs []string) (indexs []string) {
	indexs = append(indexs, "a") //"a" index for counting gloab data

	for _, indx := range t_key.Index {
		var index_str string
		//indx.i_columnref已经被sort过了，所以这里使用range，其顺序是固定的
		for _, val := range indx.i_columnref {
			index_str += strs[val]

			//store to table column of ssdb
			s_kvkey := PRE_KEYSET + PRE_INDEX_KEYOP + table.Name + "_" + strconv.Itoa(val)
			if freecache.Localcache_check_and_set(s_kvkey+strs[val]) == false {
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

	key := strs[t_key.ikey_columnref]
	date := strs[t_key.its_columnref]

	//key cloudn't be empty
	if len(key) == 0 {
		key = "nil"
	}

	date_int, err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		fmt.Printf("date error [%s]\n", date)
		return
	}

	opkey := t_key.keyopFlag

	//rtcount_correct_param(&strs)

	indexs := rtcount_gen_indexs(conn, table, t_key, strs)
	dates := rtcount_gen_dates(date_int, t_key)

	kv_pre := table.Name + "_" + t_key.Name

	if opkey&COUNT == COUNT {
		rtcount_core_count(conn, kv_pre, key, dates, indexs)
	}

	if opkey&NEW == NEW {
		rtcount_core_new(conn, kv_pre, key, dates, indexs)
	}

	if opkey&ACTIVE == ACTIVE {
		rtcount_core_active(conn, kv_pre, key, dates, indexs)
	}

	key_int, err := strconv.ParseInt(key, 10, 64)
	//change key to integer for calculating sum/max/min
	if err == nil {
		if opkey&SUM == SUM {
			rtcount_core_sum(conn, kv_pre, key_int, dates, indexs)
		}

		if opkey&MAX == MAX {
			rtcount_core_max(conn, kv_pre, key_int, dates, indexs)
		}

		if opkey&MIN == MIN {
			rtcount_core_min(conn, kv_pre, key_int, dates, indexs)
		}
	}

}

func rtcount_core_count(conn *gossdb.Client, kv_pre string, key string, dates []string, indexs []string) {
	var op_key_pre string = PRE_COUNT_KEYOP + kv_pre
	//kv_prt = table.Name + "_" + t_key.Name

	for _, indx_val := range indexs {
		for _, date_val := range dates {
			kvkey := op_key_pre + "_" + date_val + "_" + indx_val
			conn.Incr(kvkey, 1)
			//fmt.Println("count--- \n", kvkey)
		}
	}
}

func rtcount_core_new(conn *gossdb.Client, kv_pre string, key string, dates []string, indexs []string) {
	var key_set_pre string = PRE_KEYSET
	var op_key_pre string

	//-----handle new op-----------------------------------------------------------------------------
	op_key_pre = PRE_NEW_KEYOP + kv_pre
	for _, indx_val := range indexs {
		s_kvkey := key_set_pre + op_key_pre + "_a" + "_" + indx_val
		//check localcace first
		if freecache.Localcache_check_and_set(s_kvkey+key) == true {
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
		for _, date_val := range dates {
			kvkey := op_key_pre + "_" + date_val + "_" + indx_val
			conn.Incr(kvkey, 1)
			//fmt.Println("new--- \n", kvkey)
		}
	}
}

func rtcount_core_active(conn *gossdb.Client, kv_pre string, key string, dates []string, indexs []string) {
	var key_set_pre string = PRE_KEYSET
	var op_key_pre string

	//-----handle active op-----------------------------------------------------------------------------
	op_key_pre = PRE_ACTIVE_KEYOP + kv_pre
	for _, indx_val := range indexs {
		t_len := len(dates)
		for t := 0; t < t_len; t++ {
			if dates[t] == "a" {
				continue //"ALL" don't need to cale active...
			}
			kvkey := op_key_pre + "_" + dates[t] + "_" + indx_val
			s_kvkey := key_set_pre + kvkey

			//check localcace first
			if freecache.Localcache_check_and_set(s_kvkey+key) == true {
				//old key
				continue
			}

			if exists, err := conn.Zexists(s_kvkey, key); err != nil {
				//ssdb op error
				return
			} else if exists == false {
				//new key, insert it to set, and incr the counter.
				//we check from big date, so remaining date must be new key.
				for last := t; last < t_len; last++ {
					if dates[last] == "a" {
						continue //"ALL" don't need to cale active...
					}
					kvkey := op_key_pre + "_" + dates[last] + "_" + indx_val
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

func rtcount_core_sum(conn *gossdb.Client, kv_pre string, key_int int64, dates []string, indexs []string) {
	var op_key_pre string = PRE_SUM_KEYOP + kv_pre

	for _, indx_val := range indexs {
		for _, date_val := range dates {
			kvkey := op_key_pre + "_" + date_val + "_" + indx_val
			conn.Incr(kvkey, key_int)
			//fmt.Println("sum--- \n", kvkey)
		}
	}
}

func rtcount_core_max(conn *gossdb.Client, kv_pre string, key_int int64, dates []string, indexs []string) {
	var op_key_pre string = PRE_MAX_KEYOP + kv_pre

	for _, indx_val := range indexs {
		for _, date_val := range dates {
			kvkey := op_key_pre + "_" + date_val + "_" + indx_val

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
}

func rtcount_core_min(conn *gossdb.Client, kv_pre string, key_int int64, dates []string, indexs []string) {
	var op_key_pre string = PRE_MIN_KEYOP + kv_pre

	for _, indx_val := range indexs {
		for _, date_val := range dates {
			kvkey := op_key_pre + "_" + date_val + "_" + indx_val

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
}
