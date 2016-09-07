package main

import (
	//"./seefan/gossdb"
	"./freecache"
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	//"time"
)

/*var AppList = []string{"437982673", "433400453"}*/

type BIPaySucc struct {
	OrderId interface{} `json:"orderId"`
}
type BIPayRequest struct {
	OrderId interface{} `json:"orderId"`
	CAmount int         `json:"cAmount"`
}

func ck_check_log(strs []string) bool {
	if len(strs) < 48 {
		return false
	}

	appKey := strs[7]
	eventKey := strs[35]
	os := strs[11]
	channel := strs[9]
	uuid := strs[28]
	dvid := strs[47]

	// Check field
	if len(appKey) <= 3 {
		return false
	}
	/*
		match := false
		for _, app := range AppList {
			if appKey == app {
				match = true
				break
			}
		}
			if !match {
				return false
			}
	*/
	if (len(eventKey) <= 4) || (eventKey[0:3] != "cc_") {
		return false
	}
	if len(dvid) == 0 {
		strs[47] = "d" // default dev
	} else if len(dvid) >= 32 {
		strs[47] = dvid[0:32]
	}
	if len(uuid) == 0 {
		strs[28] = "u" // default user
	} else if len(uuid) >= 64 {
		strs[28] = uuid[0:64]
	}

	if len(os) == 0 {
		strs[11] = "u" // unkown
	} else if strings.ToLower(os[0:1]) == "a" {
		strs[11] = "a" // android
	} else if strings.ToLower(os[0:1]) == "i" {
		strs[11] = "i" // ios
	} else {
		strs[11] = "u" // unkown
	}

	if len(channel) != 6 {
		strs[9] = "999998" // unkown channel
	}

	return true
}

func ck_gen_pay_log(strs []string, pay_mount string) {
	/*
		<column>dvid</column>
		<column>uuid</column>
		<column>time</column>
		<column>channel</column>
		<column>appKey</column>
		<column>os</column>
		<column>pay_mount</column>
	*/
	dvid := strs[47]
	uuid := strs[28]
	time := strs[2]
	channel := strs[9]
	appKey := strs[7]
	os := strs[11]
	var pay_str []string
	pay_str = append(pay_str, dvid)
	pay_str = append(pay_str, uuid)
	pay_str = append(pay_str, time)
	pay_str = append(pay_str, channel)
	pay_str = append(pay_str, appKey)
	pay_str = append(pay_str, os)
	pay_str = append(pay_str, pay_mount)

	rtcount_before("chukong_pay", pay_str)
}

func CK_handle_log(tablename string, line []byte) {
	var logOrderId string

	xx := bytes.Split(line, []byte("\x02"))
	strs := s_byteString(xx)

	if ck_check_log(strs) == false {
		return
	}

	appKey := strs[7]
	eventKey := strs[35]
	params := strs[38]

	// pay
	if eventKey == "cc_payRequest" {
		var payReq BIPayRequest
		if len(params) >= 20 {
			json.Unmarshal([]byte(params), &payReq)
			switch v := payReq.OrderId.(type) {
			case int32, int64:
				logOrderId = fmt.Sprintf("%d", payReq.OrderId)
			case float64:
				logOrderId = fmt.Sprintf("%0.f", payReq.OrderId)
			case string:
				logOrderId = fmt.Sprintf("%s", payReq.OrderId)
			default:
				logOrderId = fmt.Sprintf("%s", payReq.OrderId)
				fmt.Println(v)
			}
		}
		orderKey := appKey + ":oid:" + logOrderId
		sucKey := "suc:" + orderKey
		//check in sucKey for ordes

		affected := freecache.Localcache_del(sucKey)
		if affected == true {
			//pay table...
			ck_gen_pay_log(strs, strconv.Itoa(payReq.CAmount))
		} else {
			freecache.Localcache_set(orderKey, strconv.Itoa(payReq.CAmount), 3600)
			return
		}

		//check in Succord
		//conn.Set(orderKey, payReq.CAmount, 3600)

	} else if eventKey == "cc_paySucc" {
		var paySucc BIPaySucc
		if len(params) >= 12 {
			json.Unmarshal([]byte(params), &paySucc)
			switch v := paySucc.OrderId.(type) {
			case int32, int64:
				logOrderId = fmt.Sprintf("%d", paySucc.OrderId)
			case float64:
				logOrderId = fmt.Sprintf("%0.f", paySucc.OrderId)
			case string:
				logOrderId = fmt.Sprintf("%s", paySucc.OrderId)
			default:
				logOrderId = fmt.Sprintf("%s", paySucc.OrderId)
				fmt.Println(v)
			}
			orderKey := appKey + ":oid:" + logOrderId
			pay, e := freecache.Localcache_get(orderKey)
			if e == freecache.ErrNotFound {
				sucKey := "suc:" + orderKey
				freecache.Localcache_set(sucKey, "1", 3600) // 1 hour
				return
			} else {
				//pay table...
				ck_gen_pay_log(strs, p_byteString(pay))
				freecache.Localcache_del(orderKey)
			}
		} else {
			return
		}

	}

	rtcount_before(tablename, strs)

}
