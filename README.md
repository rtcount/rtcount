#RTCount - 高性能,支持多维度查询的实时统计系统


##Features
* 无需编写代码，通过配置即可完成功能
* 支持五种计算方式 (计数，去重，求和，最大，最小)
* 支持自定义索引和查询
* 支持多维度时间
* 自带HTTP数据查询和接收接口
* 支持分布式部署

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

* 按文件上传:
curl -F data=@your_file http://ip:port/table/your_tablename

* 查询概要统计信息
curl http://ip:port/info

##Notice

##How it is done

##TODO
* Support SQL query
* Support system monitor and status
* Optimize localcache to RTCount
* Integrate SSDB to RTCount
* 分布式部署和数据备份管理

##License
The MIT License
