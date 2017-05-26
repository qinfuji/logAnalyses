package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"regexp"
)

//PathToHashCode 转换文件路径到hash
func pathToHashCode(filePath string) string {
	t := sha1.New()
	io.WriteString(t, filePath)
	return fmt.Sprintf("%x", t.Sum(nil))

}

//ParseURL 获取url 以及 queryString
func ParseURL(url string) (baseURL string, queryString string) {
	//line := `/content/getContentDetail?id=0158b012-b817-4f03-b107-f2d93dc9b571&aqsdasd`
	res := regexp.MustCompile(`(.*?)\?(.*)`).FindAllStringSubmatch(queryString, -1)
	fmt.Println(res)
	return res[0][1], res[0][2]
}
