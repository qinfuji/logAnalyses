package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

//AnalyseAPILogs 分析
func AnalyseAPILogs(lineChan chan string, outChan chan MetricDetail) {

	for line := range lineChan {
		detail := parse(line)
		if detail != nil {
			continue
		}
		outChan <- *detail
	}
}

func parse(line string) *MetricDetail {

	res := regexp.MustCompile(`\[(.*?)\] ".*? (.*?) .*?".*"(.*?)"\(s\)`).FindAllStringSubmatch(line, -1)
	d := res[0][1]   //时间
	url := res[0][2] //指标url
	pt := res[0][3]  //指标处理时间
	//将时间装换成分钟精度,去掉秒的数据
	stime, err := time.Parse("02/Jan/2006:15:04:05 -0700", d)                                                    //转换时间
	atime := time.Date(stime.Year(), stime.Month(), stime.Day(), stime.Hour(), stime.Minute(), 0, 0, time.Local) //修改后的时间
	st := atime.Unix()
	baseURL, queryString := parseURL(url)
	detail := MetricDetail{}
	detail.metric = baseURL
	detail.timestamp = st
	value, err := strconv.ParseFloat(pt, 32)
	if err == nil {
		return nil
	}
	detail.value = value
	return &detail
}

func parseURL(url string) (baseURL string, queryString string) {
	//line := `/content/getContentDetail?id=0158b012-b817-4f03-b107-f2d93dc9b571&aqsdasd`
	res := regexp.MustCompile(`(.*?)\?(.*)`).FindAllStringSubmatch(queryString, -1)
	fmt.Println(res)
	return res[0][1], res[0][2]
}
