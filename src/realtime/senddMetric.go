package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

//StatisticStruct 传送指标
type StatisticStruct struct {
	timestamp  int64
	total      float64 //总时长
	count      int     //总次数
	metricName string  //指标名称
}

//SendMetricDetails 发送指标到远端
func SendMetricDetails(details chan MetricDetail, waitGroup *sync.WaitGroup) {

	waitGroup.Add(1)
	var pcount = 0
	var max = 10000 //一次处理10000条记录
	var processedDetail = make(map[int64]StatisticStruct, 0)
	for {
		detail, ok := <-details
		if !ok {
			send(&processedDetail, waitGroup) //发送还未处理的数据
			fmt.Println("SendMetricDetails 管道关闭, 退出")
			waitGroup.Done()
			return
		}
		t := detail.timestamp
		m := detail.metric
		v := detail.value

		ss, ok := processedDetail[t]
		if !ok {
			processedDetail[t] = StatisticStruct{timestamp: t, metricName: m, count: 1, total: v}
		} else {
			ss.count++
			ss.total += v
		}
		pcount++ //计数器
		if pcount >= max {
			//执行指标发送
			go send(&processedDetail, waitGroup)
			processedDetail = make(map[int64]StatisticStruct, 0)
		}
	}
}

func send(processedDetail *map[int64]StatisticStruct, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)

	// for detail := range processedDetail {
	// 	fmt.Println(detail)
	// }
	//发送
	resp, err := http.Post("http://10.50.8.91:8806/write?db=fhh&&precision=ms", "application/x-www-form-urlencoded", nil)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	fmt.Println(string(body))
	waitGroup.Done()
}
