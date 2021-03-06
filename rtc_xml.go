package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	//"sort"
	"strings"
)

type RTC_Conf struct {
	XMLName xml.Name `xml:"all"`
	Port    string   `xml:"httpport"`
	Table   []Table  `xml:"table"`
	Kvdb    KVDB     `xml:"kvdb"`
}

type KVDB struct {
	Host        string `xml:"host"`
	Port        int    `xml:"port"`
	MinPoolSize int    `xml:"minpoolsize"`
	MaxPoolSize int    `xml:"maxpoolsize"`
}
type Table struct {
	Name        string      `xml:"name,attr"`
	Column      []string    `xml:"column"`
	Keys        []Table_Key `xml:"key"`
	CTerminated string      `xml:"column_terminated"`
	LTerminated string      `xml:"line_terminated"`
	TcpPort     int         `xml:"tcpport"`
}

type Table_Key struct {
	Name           string `xml:"name,attr"`
	Key_columnref  string `xml:"key_columnref"`
	ikey_columnref int    //* index number of Key_columnref in table.Column[*]
	Ts_columnref   string `xml:"timestamp_columnref"`
	its_columnref  int    //* index number of Ts_columnref in table.Column[*]
	KeyOP          KeyOp  `xml:"keyop"`
	keyopFlag      int
	Timeindex      Timeindex `xml:"timeindex"`
	Index          []Index   `xml:"index"`
	max_index      int       //store the minimun length of log
}

type KeyOp struct {
	Op []string `xml:"op"`
}

type Timeindex struct {
	Tm []string `xml:"tmindex"`
}

type Index struct {
	Name        string   `xml:"name,attr"`
	Columnref   []string `xml:"columnref"`
	i_columnref []int    //* index number of Columnref in table.Column[*]
}

var KEYOPS = []string{"count", "new", "active", "sum", "max", "min"}
var KEYOPS_default = []string{"new"}
var keyopFlag_default int = NEW

var TIMEINDEXS = []string{"min", "min5", "min10", "min30", "hour", "day", "week", "week2", "mon", "year"}
var TIMEINDEXS_defualt = []string{"hour", "day"}

func RTC_conf_GetTableByName(tablename string, rtc_conf *RTC_Conf) *Table {
	t_len := len(rtc_conf.Table)
	for i := 0; i < t_len; i++ {
		if rtc_conf.Table[i].Name == tablename {
			return &rtc_conf.Table[i]
		}
	}

	return nil
}

func GetIndexInArrayByString(item string, arr []string) int {

	for index, val := range arr {
		if val == item {
			return index
		}
	}

	return -1
}

func SortTimeIndexArray(arr []string) (ret []string) {

	//{"min", "min5", "min10", "min30", "hour", "day", "week", "week2", "mon", "year"}
	tm_size := len(TIMEINDEXS)
	for i := tm_size - 1; i >= 0; i-- {
		for _, val := range arr {
			if val == TIMEINDEXS[i] {
				ret = append(ret, strings.ToLower(val))
			}
		} //end for range arr
	} //end for TIMEINDEXS

	return ret
}

func ArrayFilterAndFormat(arr []string) (ret []string) {

	l_ary := len(arr)
	for i := 0; i < l_ary; i++ {
		if len(arr[i]) == 0 || GetIndexInArrayByString(strings.ToLower(arr[i]), ret) != -1 {
			continue
		}

		ret = append(ret, strings.ToLower(arr[i]))
	}
	/*
		if needsort == true {
			sort.Sort(sort.StringSlice(ret))
		}
	*/
	return ret
}

func CheckAndFix_table_key(table *Table, t_key *Table_Key) (string, bool) {

	var message string

	if t_key.Name == "" {
		message := " need set a name"
		return message, false
	}

	t_key.Name = strings.ToLower(t_key.Name)
	//check Key and Timestamps
	if len(t_key.Key_columnref) == 0 {
		message = "key_columnref don't set"
		goto Err
	}

	t_key.Key_columnref = strings.ToLower(t_key.Key_columnref)
	t_key.ikey_columnref = GetIndexInArrayByString(t_key.Key_columnref, table.Column)
	if t_key.ikey_columnref == -1 {
		message = "key_columnref set error"
		goto Err
	}

	if len(t_key.Ts_columnref) == 0 {
		message = "timestamp_columnref don't set"
		goto Err
	}

	t_key.Ts_columnref = strings.ToLower(t_key.Ts_columnref)
	t_key.its_columnref = GetIndexInArrayByString(t_key.Ts_columnref, table.Column)
	if t_key.its_columnref == -1 {
		message = "timestamp_columnref set error"
		goto Err
	}

	//check keyop
	t_key.keyopFlag = 0
	t_key.KeyOP.Op = ArrayFilterAndFormat(t_key.KeyOP.Op)
	if t_len := len(t_key.KeyOP.Op); t_len == 0 {
		t_key.KeyOP.Op = KEYOPS_default
		t_key.keyopFlag = keyopFlag_default
	} else {
		for _, op_val := range t_key.KeyOP.Op {
			if GetIndexInArrayByString(op_val, KEYOPS) == -1 {
				message = "keyop <" + op_val + "> don't be supported"
				goto Err
			}
			switch op_val {
			case "count":
				t_key.keyopFlag = t_key.keyopFlag | COUNT
			case "new":
				t_key.keyopFlag = t_key.keyopFlag | NEW
			case "active":
				t_key.keyopFlag = t_key.keyopFlag | ACTIVE
			case "sum":
				t_key.keyopFlag = t_key.keyopFlag | SUM
			case "max":
				t_key.keyopFlag = t_key.keyopFlag | MAX
			case "min":
				t_key.keyopFlag = t_key.keyopFlag | MIN
			}
		}
	}

	//check Timeindex
	t_key.Timeindex.Tm = ArrayFilterAndFormat(t_key.Timeindex.Tm)
	if t_len := len(t_key.Timeindex.Tm); t_len == 0 {
		t_key.Timeindex.Tm = TIMEINDEXS_defualt
	} else {

		for _, tm_val := range t_key.Timeindex.Tm {
			if GetIndexInArrayByString(tm_val, TIMEINDEXS) == -1 {
				message = "timestamps <" + tm_val + "> don't support"
				goto Err
			}
		}
	}
	t_key.Timeindex.Tm = SortTimeIndexArray(t_key.Timeindex.Tm)

	/* "ALL" is an inside timeindex for all data */
	t_key.Timeindex.Tm = append(t_key.Timeindex.Tm, "all")

	//check index
	if t_len := len(t_key.Index); t_len != 0 {
		for i := 0; i < t_len; i++ {
			t_key.Index[i].Name = strings.ToLower(t_key.Index[i].Name)
			t_key.Index[i].Columnref = ArrayFilterAndFormat(t_key.Index[i].Columnref)
			c_len := len(t_key.Index[i].Columnref)
			for c := 0; c < c_len; c++ {
				c_in := GetIndexInArrayByString(t_key.Index[i].Columnref[c], table.Column)

				if t_key.max_index < c_in {
					// store the maximum index value for this key
					t_key.max_index = c_in
				}

				t_key.Index[i].i_columnref = append(t_key.Index[i].i_columnref, c_in)
				if t_key.Index[i].i_columnref[c] == -1 {
					message = "index <" + t_key.Index[i].Name + ">'s columnref <" + t_key.Index[i].Columnref[c] + "> don't exists in table"
					goto Err
				}
			}
			/*
					//对key索引列值进行排序，这样key索引位置可以变动，而不影响其生成
					fmt.Println("\n*************************************\n")
					fmt.Println(t_key.Index[i].i_columnref)
					fmt.Println(t_key.Index[i].Columnref)

					fmt.Println("\n-------------------------------------\n")
					sort.Sort(sort.IntSlice(t_key.Index[i].i_columnref))
					fmt.Println(t_key.Index[i].i_columnref)
					fmt.Println("\n%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%\n")
				for _, val := range t_key.Index[i].i_columnref {
					fmt.Println("\n", val, "\n")
				}
			*/
		}

	}

	//fmt.Println(table)
	return "success", true
Err:
	return "[table:" + table.Name + " key:" + t_key.Name + "] " + message, false
}
func CheckAndFix_table(table *Table) (string, bool) {
	//fmt.Println(table)

	//Lower table column
	if table.Name == "" {
		message := "table need set a name"
		return message, false
	}

	table.Name = strings.ToLower(table.Name)
	table.Column = ArrayFilterAndFormat(table.Column)
	l_len := len(table.Column)
	if l_len < 2 {
		message := "[table:" + table.Name + "] is column must bigger than two"
		return message, false
	}
	/*
		for i := 0; i < l_len; i++ {
			fmt.Println("column:", table.Column[i])
		}
	*/

	if k_len := len(table.Keys); k_len == 0 {
		message := "[table:" + table.Name + "] key don't set"
		return message, false
	} else {
		for i := 0; i < k_len; i++ {
			if msg, ret := CheckAndFix_table_key(table, &table.Keys[i]); ret == false {
				return msg, ret
			}
		}

		return "success", true
	}

	fmt.Println("haha ddd")
	return "hahaha", false
}

func CheckAndFix_xml(rtc_conf *RTC_Conf) (string, bool) {
	n := len(rtc_conf.Table)
	if n == 0 {
		return "no table info!", false
	}

	for i := 0; i < n; i++ {
		if msg, res := CheckAndFix_table(&rtc_conf.Table[i]); res != true {
			return msg, res
		}
	}

	return "success", true
}

func RTC_conf_init(xmlfile string) (rtc_conf RTC_Conf, msg string, ret bool) {
	content, err := ioutil.ReadFile(xmlfile)
	if err != nil {
		return rtc_conf, "<" + xmlfile + ">read fail!", false
	}
	err = xml.Unmarshal(content, &rtc_conf)
	if err != nil {
		return rtc_conf, "xml parse<" + xmlfile + ">fail!", false
	}
	//fmt.Println(result.Table)

	msg, ret = CheckAndFix_xml(&rtc_conf)
	//fmt.Println(rtc_conf)
	return rtc_conf, msg, ret
}

/*
func main() {
	rtc_conf, msg, ret := RTC_conf_init("test.xml")
	fmt.Println(rtc_conf)
	fmt.Println(msg, ret)
}
*/
