//定时触发日志读写，并且记录日志读取位置
package main

import (
	"bufio"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
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
	var stateFileDir string
	var analyseType string
	flag.StringVar(&second, "second", "2", "crontab seconds")
	flag.StringVar(&minite, "minite", "0", "crontab minite")
	flag.StringVar(&hour, "hour", "0", "crontab hour")
	flag.StringVar(&logPath, "logPath", "./../api.log", "log file path")
	flag.StringVar(&stateFileDir, "stateFileDir", "./", "state persist dir")
	flag.StringVar(&analyseType, "analyseType", "Api", "分析文件的类型")
	flag.Parse()

	c := cron.New()
	crontabS := strings.Join([]string{second, minite, hour, "*", "*", "*"}, " ")
	fmt.Println("crontab is", crontabS)

	var state = FileReadState{logPath: logPath, lines: analyzChan, maxReadSize: 2 * 1024 * 1024, stateFileDir: stateFileDir}
	state.LoadState()

	c.AddFunc("*/5 * * * *", func() {
		process(&state)
	})

	c.Start()

	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		}
	}()

	anylizline(analyzChan, analyseType)
}

//检查文件状态，
func checkState(state *FileReadState) bool {
	return true
}

//开始处理文件
func process(state *FileReadState) {
	state.LoadState()
	state.Read()
}

func anylizline(lineChan chan string, analyseType string) {
	for {
		line := <-lineChan
		if analyseType == "Api" {
			AnalyseAPILogs(line)
		}
	}
}

type TimeStatisticResult struct {
	timeStamp  int
	metricName string
	value      int
}

//FileReadState 文件读状态
type FileReadState struct {
	offset        int64       //当前文件偏移
	stateFileDir  string      //状态文件目录
	maxReadSize   int         //一次最大读取量
	logPath       string      //日志路径
	lines         chan string //读取的管道
	handlingByte  []byte      //在处理中的字节，当没有出现完整的行
	stateFullName string      //状态文件的路,防止文件来回打开
}

//LoadState 加载当前文件读取的状态
func (state *FileReadState) LoadState() bool {

	state.offset = int64(0)
	state.handlingByte = make([]byte, 0)
	hashPath := pathToHashCode(state.logPath)
	stateFilePath := path.Join(state.stateFileDir, hashPath)
	var f *os.File
	var err1 error
	if checkFileIsExist(stateFilePath) { //如果文件存在
		f, err1 = os.Open(stateFilePath) //打开文件
		check(err1)
		if err1 == nil {
			scanner := bufio.NewScanner(f)
			line := 0
			for scanner.Scan() {
				o := scanner.Text()
				if line == 0 {
					offset, _ := strconv.ParseInt(o, 10, 64)
					state.offset = offset
				} else {
					state.handlingByte = scanner.Bytes()
				}
				line++
			}
		}
	} else {
		f, err1 = os.Create(stateFilePath) //创建文件
		state.offset = int64(0)
		state.handlingByte = []byte{}
	}
	check(err1)
	if err1 != nil {
		defer f.Close()
	}
	return true
}

//Save 保存状态
func (state *FileReadState) Save() {
	hashPath := pathToHashCode(state.logPath)
	stateFilePath := path.Join(state.stateFileDir, hashPath)
	so := strconv.FormatInt(state.offset, 10)
	d := append([]byte(so), []byte{'\n'}...)
	d = append(d, state.handlingByte...)
	ioutil.WriteFile(stateFilePath, d, 0666) //写入文件(字节数组)
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
		if readCount >= state.maxReadSize {
			break
		}
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
			state.handlingByte = append(state.handlingByte, value)
			readCount++
			state.offset++
			if value == '\n' {
				line := string(state.handlingByte)
				//fmt.Print(line)
				state.lines <- line
				state.handlingByte = make([]byte, 0)
			}
		}
	}
	state.Save()
}

//PathExists 判断文件是否存在
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//PathToHashCode 转换文件路径到hash
func pathToHashCode(filePath string) string {
	t := sha1.New()
	io.WriteString(t, filePath)
	return fmt.Sprintf("%x", t.Sum(nil))

}
