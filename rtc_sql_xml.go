package main

import (
	"encoding/xml"
	"strings"
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
	LhsAttr     string `xml:"lhsAttr"`
	Op          string `xml:"op"`
	Value       string `xml:"value"`
	Val_type    string `xml:"val_type"`
	i_columnref int
}

func RTC_sql_check(xmls string) (sql RTC_Sql, msg string, ret bool) {

	bxml := []byte(xmls)

	err := xml.Unmarshal(bxml, &sql)
	if err != nil {
		return sql, "xml parse<" + xmls + ">fail!", false
	}
	//fmt.Println(sql)

	sql.Op = strings.ToLower(sql.Op)
	sql.Table = strings.ToLower(sql.Table)
	sql.Key = strings.ToLower(sql.Key)
	sql.With[0] = strings.ToLower(sql.With[0])
	sql.With[1] = strings.ToLower(sql.With[1])

	for _, item := range sql.Condis {

		if item.Val_type == "Attr" {
			return sql, item.LhsAttr + item.Op + item.Value + " format error", false
		}
		item.LhsAttr = strings.ToLower(item.LhsAttr)
	}

	return sql, "OK", true
}
