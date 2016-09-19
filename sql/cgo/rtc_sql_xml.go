package main

import (
	"encoding/xml"
)

type RTC_Sql struct {
	XMLName xml.Name    `xml:"all"`
	Op      string      `xml:"op"`
	Table   string      `xml:"table"`
	Key     string      `xml:"key"`
	With    []string    `xml:"with"`
	Condis  []condition `xml:"condition"`
}

type condition struct {
	LhsAttr  string `xml:"lhsAttr"`
	Op       string `xml:"op"`
	Value    string `xml:"value"`
	Val_type string `xml:"val_type"`
}

func RTC_sql_check(xmls string) (sql RTC_Sql, msg string, ret bool) {

	bxml := []byte(xmls)

	err := xml.Unmarshal(bxml, &sql)
	if err != nil {
		return sql, "\n ----------xml parse<------\n" + xmls + "\n--------->fail!---------\n", false
	}

	//fmt.Println(sql)

	return sql, "OK", true
}
