package main

import (
	"fmt"
	"log"
	"sync"

	client "github.com/influxdata/influxdb/client/v2"
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

	c, err := client.NewHTTPClient(client.HTTPConfig{Addr: "http://10.50.8.91:8086"})
	if err != nil {
		fmt.Println(err)
	}

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "fhh",
		Precision: "ms",
	})
	if err != nil {
		log.Fatal(err)
	}

	waitGroup.Add(1)
	var pcount = 0
	var batchSize = 10000 //一次处理10000条记录
	var processedDetail = make(map[int64]StatisticStruct, 0)
	for {
		detail, ok := <-details

		if !ok {
			send(processedDetail, c, bp, waitGroup) //发送还未处理的数据
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
		if pcount >= batchSize {
			//执行指标发送
			go send(processedDetail, c, bp, waitGroup)
			processedDetail = make(map[int64]StatisticStruct, 0)
		}
	}
}

func send(processedDetail map[int64]StatisticStruct, influxClient client.Client, bp client.BatchPoints, waitGroup *sync.WaitGroup) {

	waitGroup.Add(1)
	batchSize := 200
	count := 0
	for _, detail := range processedDetail {
		count++
		metric := detail.metricName
		timestamp := detail.timestamp
		value := detail.total / float64(detail.count) //计算平均时间
		fmt.Println(metric, value, timestamp)
		tags := map[string]string{"metric": metric}
		fields := map[string]interface{}{"value": value, "time": timestamp}
		pt, _ := client.NewPoint("fhhKpi", tags, fields)
		bp.AddPoint(pt)
		if count >= batchSize {
			if err := influxClient.Write(bp); err != nil {
				fmt.Println(err)
			}
		}
	}

	waitGroup.Done()
}
