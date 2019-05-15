package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// JSONFile は読み込み/書き込みJSONファイルパス
const JSONFile string = "out/tokyo.json"

// MaxInfoNum はInfoの最大要素数
const MaxInfoNum int = 7

// Info は天気情報
type Info struct {
	Date           [MaxInfoNum]int    `json:"date"`
	MaxTemperature [MaxInfoNum]int    `json:"max_temp"`
	MinTemperature [MaxInfoNum]int    `json:"min_temp"`
	Image          [MaxInfoNum]string `json:"image"`
	ImageRootPath  string             `json:"image_root"`
}

func main() {
	// 記録している天気情報を読み込み
	info := new(Info)
	loadErr := loadJSON(info, JSONFile)
	if loadErr != nil {
		fmt.Println(loadErr)
	}
	isLoadSuccess := (loadErr == nil)

	// 今日より前の情報は削除
	if isLoadSuccess {
		nowTime := time.Now()
		nowDay := nowTime.Day()
		deleteBeforeDate(nowDay, info)
	}

	// Webから情報を取得
	nowInfo := new(Info)
	nowInfo.ImageRootPath = "https://www.jma.go.jp/jp/week/"
	parseWeb(nowInfo)

	// 天気情報をマージ
	if isLoadSuccess {
		mergeInfo(nowInfo, info)
	}

	// 反映した情報を記録用データに出力
	if isLoadSuccess {
		exportJSON(info, JSONFile)
	} else {
		exportJSON(nowInfo, JSONFile)
	}
}

func loadJSON(outInfo *Info, filePath string) error {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	json.Unmarshal(raw, outInfo)
	return err
}

func deleteBeforeDate(date int, outInfo *Info) {
	removeCount := 0
	for i := 0; i < MaxInfoNum; i++ {
		if outInfo.Date[i] == date {
			break
		}
		removeCount++
	}

	if removeCount == 0 {
		return
	}

	// 消す分ずらして上書き
	for i := removeCount; i < MaxInfoNum; i++ {
		copyIdx := (i - removeCount)
		outInfo.Date[copyIdx] = outInfo.Date[i]
		outInfo.MaxTemperature[copyIdx] = outInfo.MaxTemperature[i]
		outInfo.MinTemperature[copyIdx] = outInfo.MinTemperature[i]
		outInfo.Image[copyIdx] = outInfo.Image[i]
	}

	// 元は削除
	lastIdx := (MaxInfoNum - 1)
	for i := lastIdx; i > (lastIdx - removeCount); i-- {
		outInfo.Date[i] = 0
		outInfo.MaxTemperature[i] = 0
		outInfo.MinTemperature[i] = 0
		outInfo.Image[i] = ""
	}
}

func parseWeb(outInfo *Info) {
	doc, err := goquery.NewDocument("https://www.jma.go.jp/jp/week/319.html")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	parseDate(doc, outInfo)

	parseTemperature(doc, outInfo)

	parseImage(doc, outInfo)
}

func parseDate(doc *goquery.Document, outInfo *Info) {
	//print("parse date")
	selection := doc.Find(".weekday")
	if selection.Length() == 0 {
		return
	}

	index := 0
	selection.Parent().Children().Each(func(_ int, s *goquery.Selection) {
		dateInt := trimDateAsInt(s.Text())
		if dateInt > 0 {
			//print(dateInt)
			//print("\n")
			outInfo.Date[index] = dateInt
			index++
		}
	})
}

func parseTemperature(doc *goquery.Document, outInfo *Info) {
	//print("parseTemperature")

	doc.Find(".cityname").Each(func(_ int, s *goquery.Selection) {
		if s.Text() == "東京" {
			parent := s.Parent()
			index := 0
			parent.Find(".maxtemp").Each(func(_ int, s *goquery.Selection) {
				intTemp := trimTemperatureAsInt(s.Text())
				outInfo.MaxTemperature[index] = intTemp
				//print(intTemp)
				//print("\n")
				index++
			})

			parent = parent.Next()
			index = 0
			parent.Find(".mintemp").Each(func(_ int, s *goquery.Selection) {
				intTemp := trimTemperatureAsInt(s.Text())
				outInfo.MinTemperature[index] = intTemp
				//print(intTemp)
				//print("\n")
				index++
			})

			return
		}
	})
}

func parseImage(doc *goquery.Document, outInfo *Info) {
	doc.Find(".normal").Each(func(_ int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "東京地方") {
			index := 0
			s.Parent().Find("img").Each(func(_ int, s *goquery.Selection) {
				imagePath, _ := s.Attr("src")
				outInfo.Image[index] = imagePath
				index++
			})

			return
		}
	})
}

var rexLeadingDigits = regexp.MustCompile(`\d+`)

func trimDateAsInt(dateStr string) int {
	rex := rexLeadingDigits.Copy()
	value, _ := strconv.Atoi(rex.FindString(dateStr))
	return value
}

func trimTemperatureAsInt(tempStr string) int {
	temp := strings.TrimSpace(tempStr)
	temps := strings.Split(temp, "\n")
	intTemp, _ := strconv.Atoi(temps[0])
	return intTemp
}

func mergeInfo(srcInfo *Info, targetInfo *Info) {
	startIdx := -1
	for i := 0; i < MaxInfoNum; i++ {
		if targetInfo.Date[i] == srcInfo.Date[0] {
			startIdx = i
			break
		}
	}

	// 基準位置が見つからないなら、infoを全上書き
	isAllOverride := (startIdx == -1)
	if isAllOverride {
		startIdx = 0
	}

	sIdx := 0
	for tIdx := startIdx; tIdx < MaxInfoNum; tIdx++ {
		targetInfo.Date[tIdx] = srcInfo.Date[sIdx]
		targetInfo.MaxTemperature[tIdx] = srcInfo.MaxTemperature[sIdx]
		minTemp := srcInfo.MinTemperature[sIdx]
		if isAllOverride || (minTemp != 99) {
			targetInfo.MinTemperature[tIdx] = srcInfo.MinTemperature[sIdx]
		}
		targetInfo.Image[tIdx] = srcInfo.Image[sIdx]

		sIdx++
	}
}

func exportJSON(info *Info, filePath string) {
	jsonBytes, _ := json.Marshal(info)
	jsonString := string(jsonBytes)

	file, err := os.Create(filePath)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	file.Write(([]byte)(jsonString))
}
