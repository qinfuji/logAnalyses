package main

import (
	"fmt"
)

func main() {
	var analyzers []int
	//var l = len(analyzers)
	analyzers = append(analyzers, 1)
	fmt.Println(len(analyzers))

	var analyzer = BusLogAnalyzer{}
	var pipe = AnalyzerPipeline{[]Analyzer{}}
	pipe.addAnalyzer(analyzer)

}

// FileHandle  完整的文件描述
type FileHandle struct {
	offset    int    //文件当前读取的偏移量
	filePath  string //文件路径
	lastTime  int    //最后读取时间
	readCount int    //读取多少行后才输出
}

//Store 将文件信息持久化
func (fs *FileHandle) Store() bool {
	return false
}

//ReadLine 读取行
func (fs *FileHandle) ReadLine() []string {
	return nil
}

//Kpi 指标结构
type Kpi struct {
	value int    //处理时长
	id    string //某条数据的唯一标示
}

//Analyzer 分析器接口
type Analyzer interface {
	Analyze(line string) (Kpi, error)
	Match(line string) bool
}

//AnalyzerPipeline 行分析管道
type AnalyzerPipeline struct {
	analyzers []Analyzer //分析map
}

//Analyze 责任链的分析工具
func (pipeline *AnalyzerPipeline) Analyze(line string) (Kpi, error) {
	for _, analyze := range pipeline.analyzers {
		if analyze.Match(line) {
			return analyze.Analyze(line)
		}
	}
	return Kpi{}, nil
}

//BusLogAnalyzer 总线日志分析
type BusLogAnalyzer struct{}

//Analyze 总线接口数据分析
func (analyzer *BusLogAnalyzer) Analyze(line string) (Kpi, error) {
	fmt.Println(line)
	return Kpi{}, nil
}

//Match 是否匹配
func (analyzer *BusLogAnalyzer) Match(ling string) bool {
	return true
}
