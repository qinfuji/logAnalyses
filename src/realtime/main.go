//定时触发日志读写，并且记录日志读取位置
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/robfig/cron"
)

//MetricDetail 指标结构
type MetricDetail struct {
	metric    string  //指标名称
	value     float64 //指标值
	timestamp int64   //时间， 分钟
}

var stop = false
var stopLock sync.Mutex
var signalChan = make(chan os.Signal, 1)
var mcrom = cron.New()
var lineChan = make(chan string, 10)               //行分析队列
var metricDetailChan = make(chan MetricDetail, 10) //处理后的结果队列

var readCloseChan = make(chan string) //文件读停止管道

var waitGroup sync.WaitGroup //定义一个同步等待的组

func main() {

	var second string
	var minite string
	var hour string
	var logPath string
	var stateFileDir string
	var analyseType string
	flag.StringVar(&second, "second", "0", "crontab seconds")
	flag.StringVar(&minite, "minite", "2", "crontab minite")
	flag.StringVar(&hour, "hour", "0", "crontab hour")
	//flag.StringVar(&logPath, "logPath", "/home/qinfuji/service_sys-20170526.log", "log file path")
	flag.StringVar(&logPath, "logPath", "../../fhh_entry_sys.log", "log file path")
	flag.StringVar(&stateFileDir, "stateFileDir", "../..", "state persist dir")
	//flag.StringVar(&analyseType, "analyseType", "QueryApi", "分析文件的类型") //QueryApi |  BusApi
	flag.StringVar(&analyseType, "analyseType", "BusApi", "分析文件的类型") //QueryApi |  BusApi
	flag.Parse()

	//c := cron.New()
	crontabS := strings.Join([]string{second, minite, hour, "*", "*", "*"}, " ")
	fmt.Println("crontab is", crontabS)

	waitGroup.Add(1)
	var state = FileReadState{logPath: logPath, lines: lineChan, maxReadSize: 10 * 1024 * 1024, stateFileDir: stateFileDir, waitGroup: &waitGroup}
	state.LoadState()

	mcrom.AddFunc("*/10 * * * *", func() {
		process(&state)
	})

	mcrom.Start()

	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		}
	}()

	if analyseType == "QueryApi" { //客户端查询接口调用
		//go AnalyseAPILogs(lineChan, metricDetailChan, &waitGroup)
	} else if analyseType == "BusApi" { //系统间调用接口
		go AnalyseBusLogs(lineChan, metricDetailChan, &waitGroup)
	}
	go SendMetricDetails(metricDetailChan, &waitGroup)

	// signal.Notify(signalChan,
	// 	os.Kill,
	// 	os.Interrupt,
	// 	syscall.SIGHUP,
	// 	syscall.SIGINT,
	// 	syscall.SIGTERM,
	// 	syscall.SIGQUIT)

	// go waitExitNotify(&waitGroup)

	select {}
}

//检查文件状态，
func checkState(state *FileReadState) bool {
	return true
}

//开始处理文件
func process(state *FileReadState) {
	mcrom.Stop()
	state.LoadState()
	state.Read()
	mcrom.Start()
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
	waitGroup     *sync.WaitGroup
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
	if stop {
		//如果已经停止则需要发送文件读完成消息
		readCloseChan <- "close"
	}
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

func waitExitNotify(waitGroup *sync.WaitGroup) {
	<-signalChan
	stopLock.Lock()
	stop = true
	stopLock.Unlock()
	mcrom.Stop() //停止
	fmt.Println("任务调度停止")
	fmt.Println("等待文件读写停止")
	<-readCloseChan
	fmt.Println("文件读写停止")

	fmt.Println("关闭行分析管道")
	close(lineChan)
	fmt.Println("关闭指标处理管道")
	close(metricDetailChan)
	waitGroup.Wait()
	fmt.Println("系统退出")
	os.Exit(0)
}
