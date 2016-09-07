#RTCount - 实时数据统计系统


##Features
* 无需代码，使用XML文件配置
* 五种计算方式 (计数，去重，求和，最大，最小)
* 自定义索引查询
* 多时间维度支持
* 自带数据接收和查询的http接口

##Build
./build.sh

##Example Usage
1. write your.xml 
2. ./count your.xml  

##SSDB
start ssdb
./ssdb/ssdb-server -d ./ssdb/ssdb.conf


##API
* 上传单条数据:
curl -d "your log data" http://ip:port/table/your_tablename

 *按文件上传:
curl -F test=@your_file http://ip:port/table/your_tablename

* 查询概要统计信息
curl http://ip:port/info

##Notice

##How it is done

##TODO
* Support SQL query
* Optimize localcache to RTCount
* Integrate SSDB to RTCount

##License
The MIT License
