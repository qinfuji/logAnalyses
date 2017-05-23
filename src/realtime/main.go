//定时触发日志读写，并且记录日志读取位置
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/robfig/cron"
)

func main() {
	analyzChan := make(chan string) //行分析队列
	//resultChan := make(chan TimeStatisticResult) //处理后的结果队列

	var second string
	var minite string
	var hour string
	var logPath string
	flag.StringVar(&second, "second", "2", "crontab seconds")
	flag.StringVar(&minite, "minite", "0", "crontab minite")
	flag.StringVar(&hour, "hour", "0", "crontab hour")
	flag.StringVar(&logPath, "logPath", "./main.go", "log file path")
	flag.Parse()

	c := cron.New()
	crontabS := strings.Join([]string{second, minite, hour, "*", "*", "*"}, " ")
	fmt.Println("crontab is", crontabS)

	var state = FileReadState{logPath: logPath, lines: analyzChan, maxReadSize: 1024}
	state.LoadState()

	c.AddFunc("*/2 * * * *", func() {
		//fmt.Println(1)
		// if !checkState(&state) {
		// 	state = FileReadState{logPath: logPath, lines: analyzChan, maxReadSize: 1024*1024}
		// 	state.LoadState()
		// }
		process(&state)
	})

	c.Start()
	go anylizline(analyzChan)

	select {}
}

//检查文件状态，
func checkState(state *FileReadState) bool {
	return true
}

//开始处理文件
func process(state *FileReadState) {
	state.Read()
}

func anylizline(lineChan chan string) {
	for {
		line := <-lineChan
		fmt.Print(line)
	}
}

type TimeStatisticResult struct {
	timeStamp  int
	metricName string
	value      int
}

//FileReadState 文件读状态
type FileReadState struct {
	offset       int64       //当前文件偏移
	stateFile    string      //状态文件路径
	maxReadSize  int         //一次最大读取量
	logPath      string      //日志路径
	lines        chan string //读取的管道
	handlingByte []byte      //在处理中的字节，当没有出现完整的行
}

//LoadState 加载当前文件读取的状态
func (state *FileReadState) LoadState() {
	//读取文件状态，如果不存在则创建
	state.offset = int64(0)
	state.handlingByte = make([]byte, 0)
}

//Save 保存状态
func (state *FileReadState) Save() {

}

//读取文件内容
func (state *FileReadState) Read() {

	file, err := os.Open(state.logPath)
	defer file.Close()
	if err != nil {
		fmt.Println("Failed to open log file", err)
		return
	}
	file.Seek(state.offset, 0)
	lineBuffer := make([]byte, state.maxReadSize)
	var readCount = 0
	for {
		n, err := file.Read(lineBuffer) //读取文件
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("read error:", err)
			return
		}
		if n == 0 {
			break
		}
		b := lineBuffer[:n]
		for _, value := range b {
			readCount++
			state.offset++
			state.handlingByte = append(state.handlingByte, value)
			if value == '\n' {
				line := string(state.handlingByte)
				fmt.Print(line)
				//state.lines <- line
				state.handlingByte = make([]byte, 0)
			}
			if readCount >= state.maxReadSize {
				break
			}
		}
	}
}
