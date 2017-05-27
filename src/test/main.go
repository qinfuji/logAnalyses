package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

func main() {
	// line := `2017-05-23T00:00:09+08:00 wemedia_service_21_syq service_sys 10.90.6.59 - - [23/May/2017:03:10:00 +0800] "GET /content/getContentDetail?id=0158b012-b817-4f03-b107-f2d93dc9b571 HTTP/1.1" "local.fhhapi.ifeng.com" - 200 3562 "-" "-" "10.90.2.36" "0.002"(s)`
	// res := regexp.MustCompile(`\[(.*?)\] ".*? (.*?) .*?".*"(.*?)"\(s\)`).FindAllStringSubmatch(line, -1)
	// d := res[0][1]
	// url := res[0][2]
	// pt := res[0][3]
	// fmt.Println(d, url, pt)

	// stime, _ := time.Parse("02/Jan/2006:15:04:05 -0700", d)                                                      //转换时间
	// atime := time.Date(stime.Year(), stime.Month(), stime.Day(), stime.Hour(), stime.Minute(), 0, 0, time.Local) //修改后的时间
	// ts := atime.Unix()
	// fmt.Println(ts)

	//parseURL("")
	// paseDetailQuery("")

	//parseBusLog()

	//testChan()

	//testMap()

	sentInfulx()
	select {}
}

func parseURL(url string) (baseURL string, queryString string) {
	line := `/content/getContentDetail?id=16378201`
	res := regexp.MustCompile(`(.*?)\?(.*)`).FindAllStringSubmatch(line, -1)
	fmt.Println(res[0][1], res[0][2])
	return res[0][1], res[0][2]
}

//分析文章详情接口，返回内容类型  vedio article
func paseDetailQuery(queryString string) (contentType string) {
	line := `id=0158b012-b817-4f03-b107-f2d93dc9b571&aqsdasd`
	res := regexp.MustCompile(`id=(.*)&.*`).FindAllStringSubmatch(line, -1)
	id := res[0][1]
	if strings.Index(id, "-") > -1 {
		return "Video"
	}
	return "Article"
}

func parseBusLog() {
	line := `2017-05-23T00:01:02+08:00 wemedia_api_web140v13_syq fhh_entry_sys 10.90.6.58 - - [23/May/2017:00:01:02 +0800] "POST /stream/video/transcode?tag=1495468773792_fa120ba6-6576-42fe-982b-d32bcc90e241 HTTP/1.1" 200 1 "-" "Java/1.8.0_102" "10.50.6.53" "0.003"(s)
2`
	res := regexp.MustCompile(`\[(.*?)\] ".*? (.*?) .*?".*"(.*?)"\(s\)`).FindAllStringSubmatch(line, -1)

	//fmt.Println(res)
	d := res[0][1]   //时间
	url := res[0][2] //指标url

	fmt.Println(d, url)

	res1 := regexp.MustCompile(`(.*?)\?tag=(.*?)_(.*)`).FindAllStringSubmatch(url, -1)
	reqTimeStr := res1[0][2]
	fmt.Println("reqTime", reqTimeStr)
	reqTime, _ := strconv.ParseInt(reqTimeStr, 10, 64)

	stime, _ := time.Parse("02/Jan/2006:15:04:05 -0700", d)
	atime := time.Date(stime.Year(), stime.Month(), stime.Day(), stime.Hour(), stime.Minute(), 0, 0, time.Local)
	resTime := atime.Unix() * 1000

	fmt.Println(resTime, reqTime)
	offsettime := (resTime - reqTime)
	fmt.Println(offsettime / 1000)
}

//Tchan aa
type Tchan struct {
	i int
}

func testChan() {
	ch := make(chan Tchan, 100)
	chSend := make(chan Tchan)

	go go1(ch)
	go go2(chSend)

	go func() {
		for i := 0; i < 10; i++ {
			ch <- Tchan{i}
		}

		for i := 0; i < 10; i++ {
			chSend <- Tchan{i}
		}
	}()

	close(ch)
	close(chSend)

	select {}
}

func go1(tchan chan Tchan) {
	for {
		time.Sleep(1000 * time.Millisecond)
		i, ok := <-tchan
		if !ok {
			fmt.Println("go1 chan close")
		} else {
			fmt.Println("go1", i)
		}

	}
}

func go2(tchan chan Tchan) {
	for {
		time.Sleep(1000 * time.Millisecond)
		i, ok := <-tchan
		if !ok {
			fmt.Println("go2 chan close")
		} else {
			fmt.Println("go2", i)
		}

	}
}

//StatisticStruct aa
type StatisticStruct struct {
	timestamp  int64
	total      float64 //总时长
	count      int     //总次数
	metricName string  //指标名称
}

func testMap() {
	m := make(map[int64]StatisticStruct, 0)
	m[1] = StatisticStruct{}
	m[2] = StatisticStruct{}
	m[3] = StatisticStruct{}
	m[4] = StatisticStruct{}
	m[5] = StatisticStruct{}

	for _, v := range m {
		fmt.Println(v)
	}
}

func sentInfulx() {
	// body := make([]string, 0)
	// //body = append(body, "test,host=server02 value=0.67 1434055562000000000\n")1434055562000000000
	// body = append(body, "abc,host=server02 value=0.674 1495471620000")
	// //body = append(body, "test224,host=server02 value=0.64 1495471620000\n")
	// //body = append(body, "test224,host=server02 value=0.64 1495471620000")

	// //sbody := strings.Join(body, "")
	// resp, err := http.Post("http://10.50.8.91:8086/write?db=fhh&&precision=ms", "", strings.NewReader("abc1,host=server01,region=us-west value=0.64 1495471620000"))
	// if err != nil {
	// 	fmt.Println("save to influx err:", err)
	// 	return
	// }
	// defer resp.Body.Close()
	// data, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(data))
	c, err := client.NewHTTPClient(client.HTTPConfig{Addr: "http://10.50.8.91:8086"})
	if err != nil {
		log.Fatal(err)
	}

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "fhh",
		Precision: "ms",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a point and add to batch
	tags := map[string]string{"cpu": "cpu-total"}
	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
		"time":   1495875117}
	fmt.Println(time.Now().Unix())
	pt, err := client.NewPoint("myPoint4", tags, fields)
	if err != nil {
		fmt.Println(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		fmt.Println(err)
	}
}
