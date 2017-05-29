package main

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"
)

//AnalyseBusLogs 分析总线日志
func AnalyseBusLogs(lineChan chan string, outChan chan MetricDetail, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	for {
		line, ok := <-lineChan
		if !ok {
			fmt.Println("AnalyseBusLogs 管道关闭, 退出")
			waitGroup.Done()
			return
		}
		detail := parseBusLine(line)
		if detail != nil {
			continue
		}
		outChan <- *detail
	}
}

func parseBusLine(line string) *MetricDetail {
	fmt.Println(line)
	res := regexp.MustCompile(`\[(.*?)\] ".*? (.*?) .*?".*"(.*?)"\(s\)`).FindAllStringSubmatch(line, -1)
	d := res[0][1]   //时间
	url := res[0][2] //指标url
	res1 := regexp.MustCompile(`(.*?)\?tag=(.*?)_(.*)`).FindAllStringSubmatch(url, -1)

	reqTimeStr := res1[0][2] //请求时间
	reqTimeInt, _ := strconv.ParseInt(reqTimeStr, 10, 64)
	reqTime := time.Unix(reqTimeInt, 0)
	reqTimeVal := reqTime.UnixNano() / 1e6
	fmt.Println("reqTime", reqTimeVal)

	stime, _ := time.Parse("02/Jan/2006:15:04:05 -0700", d) //响应时间
	resTime := stime.UnixNano() / 1e6
	fmt.Println("resTime", resTime)

	baseURL, _ := ParseURL(url)
	detail := MetricDetail{}
	detail.metric = baseURL
	//使用请求时间作为key
	reqMinuteTime := time.Date(reqTime.Year(), reqTime.Month(), reqTime.Day(), reqTime.Hour(), reqTime.Minute(), 0, 0, time.Local)
	detail.timestamp = reqMinuteTime.UnixNano() / 1e6
	detail.value = float64(resTime - reqTimeVal) //处理时长
	fmt.Println("--->parseBusLine", detail)
	return &detail
}
