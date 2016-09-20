
export LD_LIBRARY_PATH=./lib
rm count;
cd sql
./b.sh
cd ../
go build -o count rtc_xml.go rtc_main.go rtc_http.go handle_ck_log.go rtc_socket.go rtc_core.go rtc_sql_query.go rtc_sql_xml.go
