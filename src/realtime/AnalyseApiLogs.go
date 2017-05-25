package main

import (
	"fmt"
	"regexp"
	"time"
)

type Detail struct {
	metric    string //指标名称
	value     int    //指标值
	timestamp int    //时间， 分钟
}

var details map[int]*Detail

//AnalyseAPILogs 分析
func AnalyseAPILogs(lineChan chan string) {

	for line := range lineChan {
		parse(line)
	}
}

//2017-05-23T00:00:09+08:00 wemedia_service_21_syq service_sys 10.90.6.59 - - [23/May/2017:00:00:00 +0800] "GET /content/getContentDetail?id=0158b012-b817-4f03-b107-f2d93dc9b571 HTTP/1.1" "local.fhhapi.ifeng.com" - 200 3562 "-" "-" "10.90.2.36" "0.002"(s)

func parse(line string) Detail {
	s := `2017-05-23T00:00:09+08:00 wemedia_service_21_syq service_sys 10.90.6.59 - - [23/May/2017:00:00:00 +0800] "GET /content/getContentDetail?id=0158b012-b817-4f03-b107-f2d93dc9b571 HTTP/1.1" "local.fhhapi.ifeng.com" - 200 3562 "-" "-" "10.90.2.36" "0.002"(s)`
	d := regexp.MustCompile(`\[(.*?)\]`).FindAllStringSubmatch(s, -1)
	fmt.Println(d[0][1])
	stime, err := time.Parse("02/Jan/2006:15:04:05 -0700", d[0][1])
	//the_time, err := time.Parse("02/Jan/2006:15:04",  "23/May/2017:00:20") 格式取整
	if err == nil {
		unixtime := stime.Unix()
		fmt.Println(unixtime)
	} else {
		fmt.Println(err)
	}

	url := regexp.MustCompile(`GET (.*?) HTTP`).FindAllStringSubmatch(s, -1)
	fmt.Println(url[0][1])

	v := regexp.MustCompile(`"(.*?)"`).FindAllStringSubmatch(s, -1)
	fmt.Println(v[len(v)-1][1])
	return Detail{}
}
