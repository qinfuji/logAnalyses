package main

import (
	"fmt"
	"sync"
	"time"
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
	var batchSize = 5 //每5分钟同步一次
	var processedDetail = make(map[int64]*StatisticStruct, 0)
	for {
		detail, ok := <-details

		if !ok {
			send(processedDetail, waitGroup) //发送还未处理的数据
			fmt.Println("SendMetricDetails 管道关闭, 退出")
			waitGroup.Done()
			return
		}
		st := time.Unix(detail.timestamp/1000, 0)
		atime := time.Date(st.Year(), st.Month(), st.Day(), st.Hour(), st.Minute(), 0, 0, time.Local)
		t := atime.UnixNano() / 1e6 //以分钟为key
		t1 := detail.timestamp
		m := detail.metric
		v := detail.value

		ss, ok := processedDetail[t]
		if !ok {
			processedDetail[t] = &StatisticStruct{timestamp: t1, metricName: m, count: 1, total: v}
		} else {
			ss.count++
			ss.total += v
		}

		if len(processedDetail) >= batchSize {
			go send(processedDetail, waitGroup)
			processedDetail = make(map[int64]*StatisticStruct, 0)
		}

	}
}

func send(processedDetail map[int64]*StatisticStruct, waitGroup *sync.WaitGroup) {

	waitGroup.Add(1)
	for _, detail := range processedDetail {
		metric := detail.metricName
		timestamp := detail.timestamp
		value := detail.total / float64(detail.count) //计算平均时间
		fmt.Println("----->SendMetricDetails", metric, value, timestamp)

	}

	waitGroup.Done()
}
