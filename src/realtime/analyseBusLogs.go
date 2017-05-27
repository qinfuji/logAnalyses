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
		detail := parse(line)
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
	reqTimeStr := res1[0][2]
	fmt.Println("reqTime", reqTimeStr)
	reqTime, _ := strconv.ParseInt(reqTimeStr, 10, 64)

	stime, _ := time.Parse("02/Jan/2006:15:04:05 -0700", d)
	atime := time.Date(stime.Year(), stime.Month(), stime.Day(), stime.Hour(), stime.Minute(), 0, 0, time.Local)
	resTime := atime.Unix()

	baseURL, _ := ParseURL(url)
	offsettime := (resTime - reqTime)
	detail := MetricDetail{}
	detail.metric = baseURL
	detail.timestamp = reqTime
	detail.value = float64(offsettime)
	return &detail
}
