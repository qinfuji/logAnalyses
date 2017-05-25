//定时触发日志读写，并且记录日志读取位置
package main

import (
	"crypto/sha1"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"io/ioutil"
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
	flag.StringVar(&logPath, "logPath", "../../api.log", "log file path")
	flag.Parse()

	c := cron.New()
	crontabS := strings.Join([]string{second, minite, hour, "*", "*", "*"}, " ")
	fmt.Println("crontab is", crontabS)

	var state = FileReadState{logPath: logPath, lines: analyzChan, maxReadSize: 64, stateFileDir: "./"}
	state.LoadState()

	c.AddFunc("*/1 * * * *", func() {
		process(&state)
	})

	c.Start()
	anylizline(analyzChan)
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

	//state.offset = int64(0)
	//state.handlingByte = make([]byte, 0)
	hashPath := PathToHashCode(state.logPath)
	stateFilePath := path.Join(state.stateFileDir, hashPath)
	content , err := ioutil.ReadFile(stateFilePath)

	if  err == nil{
		fmt.Println("nil")
		return false;
	}
	if(err == io.EOF){
		fmt.Println("EOF")
		return false;
	}
	ob := make([]byte, 0)
	handlingByte := make([]byte, 0)
	line := 1
	for _, value := range content {
	   if value == '\n' {
		 line++
		 continue
	   }
	   if line == 1 {
			ob = append(ob, value)
			continue
	   }
	   handlingByte = append(handlingByte, value)
	}
	offset, _ := strconv.ParseInt(string(ob), 10, 64)
	fmt.Println("LoadState", offset)
	fmt.Println("LoadState", len(handlingByte))
	state.offset = offset
	state.handlingByte = handlingByte
	return true
}

//Save 保存状态
func (state *FileReadState) Save() {

	hashPath := PathToHashCode(state.logPath)
	stateFilePath := path.Join(state.stateFileDir, hashPath)
	so := strconv.FormatInt(state.offset, 10)
	b := append([]byte(so), []byte{'\n'}...)
	b = append(b, state.handlingByte...)
	fmt.Println("Save", len(b))
	ioutil.WriteFile(stateFilePath, b, os.ModeAppend)
	fmt.Println("Save", state.offset)
	fmt.Println("Save", len(state.handlingByte))
}

//读取文件内容
func (state *FileReadState) Read() {
	file, err := os.Open(state.logPath)
	defer file.Close()
	if err != nil {
		fmt.Println("Failed to open log file", err)
		return
	}
	fmt.Println("Read", state.offset)
	fmt.Println("Read", len(state.handlingByte))
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
		//state.offset = state.offset + int64(n)
		b := lineBuffer[:n]
		for _, value := range b {
			state.handlingByte = append(state.handlingByte, value)
			readCount++
			state.offset++

			if value == '\n' {
				line := string(state.handlingByte)
				fmt.Print(line)
				//state.lines <- line
				state.handlingByte = make([]byte, 0)
				
			}
		}
	}
	fmt.Println("----->" , len(state.handlingByte))
	state.Save()
}

//PathExists 判断文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//PathToHashCode 转换文件路径到hash
func PathToHashCode(filePath string) string {
	t := sha1.New()
	io.WriteString(t, filePath)
	return fmt.Sprintf("%x", t.Sum(nil))

}

//Int64ToBytes int64 to []byte
func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

//BytesToInt64 []byte to int64
func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
