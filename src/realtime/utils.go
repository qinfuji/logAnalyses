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
	res := regexp.MustCompile(`(.*?)\?(.*)`).FindAllStringSubmatch(url, -1)
	if !(len(res) > 0 && len(res[0]) >= 2) {
		return "", ""
	}
	return res[0][1], res[0][2]
}
