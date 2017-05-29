package main

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"
)

//AnalyseAPILogs 分析
func AnalyseAPILogs(lineChan chan string, outChan chan MetricDetail, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	for {
		line, ok := <-lineChan
		if !ok {
			fmt.Println("AnalyseAPILogs 管道关闭, 退出")
			waitGroup.Done()
			return
		}
		detail := parse(line)
		//fmt.Println("AnalyseAPILogs", detail)
		if detail == nil {
			continue
		}
		outChan <- *detail
	}
}

func parse(line string) *MetricDetail {
	fmt.Println(line)
	res := regexp.MustCompile(`\[(.*?)\] ".*? (.*?) .*?".*"(.*?)"\(s\)`).FindAllStringSubmatch(line, -1)
	if !(len(res) > 0 && len(res[0]) >= 3) {
		fmt.Println("line content error")
		return nil
	}
	d := res[0][1]   //时间
	url := res[0][2] //指标url
	pt := res[0][3]  //指标处理时间
	//将时间装换成分钟精度,去掉秒的数据
	stime, err := time.Parse("02/Jan/2006:15:04:05 -0700", d)                                                    //转换时间
	atime := time.Date(stime.Year(), stime.Month(), stime.Day(), stime.Hour(), stime.Minute(), 0, 0, time.Local) //修改后的时间
	st := atime.UnixNano() / 1e6
	fmt.Println("-->", url)
	baseURL, _ := ParseURL(url)
	if baseURL == "" {
		fmt.Println("ParseURL err", url)
		return nil
	}
	detail := MetricDetail{}
	detail.metric = baseURL
	detail.timestamp = st
	value, err := strconv.ParseFloat(pt, 64)
	if err != nil {
		return nil
	}
	detail.value = value * 1000
	return &detail
}
