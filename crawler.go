package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

func HttpGet(url string) (result string, err error) {

	//由于部分网站有反爬程序，所以要加头信息
	//1.创建客户端
	client := &http.Client{}
	//2.创建请求request
	req, err3 := http.NewRequest("GET", url, nil)
	if err3 != nil {
		err = err3
		return
	}
	//3.给req添加头信息
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36")
	resp, err1 := client.Do(req)

	//对于没有反爬程序的网站，可以直接通过http.Get(url string) 方法调用
	//resp, err1 := http.Get(url)
	if err1 != nil {
		err = err1
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 4096)
	//遍历读取
	for {
		n, err2 := resp.Body.Read(buf)
		if n == 0 {
			break
		}
		if err2 != nil && err2 != io.EOF {
			err = err2
			return
		}
		result += string(buf[:n])
	}
	return
}

//保存到本地函数
func SaveAsFile(names, scores, nums [][]string, ids int) (err error) {
	//首先创建文件
	file, err1 := os.Create("第" + strconv.Itoa(ids) + "页数据.txt")
	if err1 != nil {
		err = err1
		return
	}
	file.WriteString("电影名称\t\t电影评分\t\t评价人数\t" + "\r\n")
	for i := 0; i < len(names); i++ {
		file.WriteString(names[i][1] + "\t\t" + scores[i][1] + "\t\t" + nums[i][1] + "\r\n")
	}
	file.Close()
	return
}

func SpiderPageDb(ids int, pageChan chan int) {
	num := (ids - 1) * 25
	url := "https://movie.douban.com/top250?start=" + strconv.Itoa(num) + "&filter="
	result, err := HttpGet(url)
	if err != nil {
		fmt.Println("HttpGet err:", err)
		return
	}
	//fmt.Println(result)
	/*
		fileName   regexp  <img width="100" alt="fileName"
		fileScore  regexp  <span class="rating_num" property="v:average">fileScore<span>
		evalNum    regexp  <span>evalNum人评价</span>
	*/

	//获取电影名称
	//(?s:(.*?)) 默认.*不匹配换行符等
	//(?s)即Singleline(单行模式)。表示更改.的含义，使它与每一个字符匹配（包括换行 符\n）。
	fileNameExp := `<img width="100" alt="(?s:(.*?))"`
	fileNameReg := regexp.MustCompile(fileNameExp)
	fileNames := fileNameReg.FindAllStringSubmatch(result, -1)

	//获取电影评分
	//(.*?)表示只匹配0个或1个
	fileScoreExp := `<span class="rating_num" property="v:average">(.*?)</span>`
	fileScoreReg := regexp.MustCompile(fileScoreExp)
	fileScores := fileScoreReg.FindAllStringSubmatch(result, -1)
	//获取电影评价人数
	//(.*?)表示只匹配0个或1个
	evalNumExp := ` <span>(.*?)人评价</span>`
	evalNumReg := regexp.MustCompile(evalNumExp)
	evalNums := evalNumReg.FindAllStringSubmatch(result, -1)

	//将爬取的数据保存到本地
	err = SaveAsFile(fileNames, fileScores, evalNums, ids)
	if err != nil {
		fmt.Println("SaveAsFile err:", err)
		return
	}
	//表示当前go程执行完毕
	pageChan <- ids
}

func ToWork(start, end int) {
	pageChan := make(chan int)
	for i := start; i <= end; i++ {
		go SpiderPageDb(i, pageChan)
	}
	for i := start; i <= end; i++ {
		fmt.Printf("第%v页数据爬取完成\n", <-pageChan)
	}
}
func main() {
	/*
		https://movie.douban.com/top250?start=0&filter=      1
		https://movie.douban.com/top250?start=25&filter=     2
		https://movie.douban.com/top250?start=50&filter=     3
		https://movie.douban.com/top250?start=75&filter=     4
	*/
	beginTime := time.Now().Unix()
	var start, end int
	fmt.Print("请输入爬取起始页：")
	fmt.Scan(&start)
	fmt.Print("请输入爬取末尾页：")
	fmt.Scan(&end)
	ToWork(start, end)
	endTime := time.Now().Unix()
	fmt.Printf("爬取数据共耗时:%v秒", endTime-beginTime)
}
