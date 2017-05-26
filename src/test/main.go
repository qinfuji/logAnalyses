package main

import "regexp"
import "fmt"

import "strings"

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

	// parseURL(url)
	// paseDetailQuery("")

	testMap()
}

func parseURL(url string) (baseURL string, queryString string) {
	line := `/content/getContentDetail?id=0158b012-b817-4f03-b107-f2d93dc9b571&aqsdasd`
	res := regexp.MustCompile(`(.*?)\?(.*)`).FindAllStringSubmatch(line, -1)
	fmt.Println(res)
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

type Sss struct {
	a int
}

func testMap() {
	var m = make(map[int]Sss, 0)
	v, ok := m[1]
	fmt.Println(v)
	fmt.Println(ok)
}
