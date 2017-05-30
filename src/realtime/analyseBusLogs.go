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

	res := regexp.MustCompile(`\[(.*?)\] ".*? (.*?) .*?".*"(.*?)"\(s\)`).FindAllStringSubmatch(line, -1)
	d := res[0][1]   //时间
	url := res[0][2] //指标url
	res1 := regexp.MustCompile(`(.*?)\?tag=(.*?)_(.*)`).FindAllStringSubmatch(url, -1)

	reqTimeStr := res1[0][2] //请求时间,精度是毫秒
	reqTime, _ := strconv.ParseInt(reqTimeStr, 10, 64)

	stime, _ := time.Parse("02/Jan/2006:15:04:05 -0700", d) //响应时间
	resTime := stime.UnixNano() / 1e6

	baseURL, _ := ParseURL(url)
	detail := MetricDetail{}
	detail.metric = baseURL
	//使用请求时间作为key
	reqTimeTime := time.Unix(reqTime/int64(1000), 0) //go 时间戳 比java少了3位 精确到秒
	//reqMinuteTime := time.Date(reqTimeTime.Year(), reqTimeTime.Month(), reqTimeTime.Day(), reqTimeTime.Hour(), reqTimeTime.Minute(), 0, 0, time.Local)
	detail.timestamp = reqTimeTime.UnixNano() / 1e6 //紧缺到毫秒

	offsetValue := float64(resTime - reqTime)
	if offsetValue < 0 {
		detail.value = 0
	} else {
		detail.value = float64(resTime - reqTime) //处理时长
	}
	return &detail
}
