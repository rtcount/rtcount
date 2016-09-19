package main

import (
	// #cgo LDFLAGS: -L${SRCDIR}/ -L./ -lparser
	/*
		#include <stdlib.h>
		#include "api.h"
	*/
	"C"
	"fmt"
	"unsafe"
)

func main() {
	//fmt.Println(C.Num)
	fmt.Println(C.random())

	str1 := "select ddd from ddd.ddd with ddd and ddd where ddd >= 1466252795 and ddd_id = 'ddd' and ddd = \"ddd\" and ddd =\"ddd\" and ddd >'ddd';"
	cstr := C.CString(str1)

	//fmt.Println(str1)
	//C.hello()
	cc := C.GoString(C.ddd(cstr))
	//fmt.Println(cc)

	sql, msg, ret := RTC_sql_check(cc)

	fmt.Println(sql, msg, ret)

	str2 := "select asdddd from T_devices.ddd with zxc and xxx where created_at >= 1466252795 and product_id = 'T_devices.product_id' and product_id = \"T_devices.product_id\" and time =\"ad\" and time >'asd';"
	cstr = C.CString(str2)

	//fmt.Println(str2)
	dd := C.GoString(C.ddd(cstr))
	//fmt.Println(dd)

	sql, msg, ret = RTC_sql_check(dd)
	fmt.Println(sql, msg, ret)

	C.free(unsafe.Pointer(cstr))
}
