package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
    nowInfo.ImageRootPath = "https://s.yimg.jp/images/weather/general/next/size150/"
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
    doc, err := goquery.NewDocument("https://weather.yahoo.co.jp/weather/jp/13/4410.html")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	parseForecast(doc, outInfo)
    parseWeek(doc, outInfo)
}

func parseForecast(doc *goquery.Document, outInfo *Info) {
    forecast := doc.Find("div.forecastCity > table > tbody > tr > td > div")
	if forecast.Length() == 0 {
		return
	}

    today := forecast.First()
    parseForecastPair(today, outInfo, 0)

    tomorrow := forecast.Last()
    parseForecastPair(tomorrow, outInfo, 1)
}

func parseForecastPair(selection *goquery.Selection, outInfo *Info, listIdx int) {
    // 日付
    date := selection.Find(".date")
    outInfo.Date[listIdx] = trimDateAsInt(date.Text())

    // 天気マーク
    img := selection.Find(".pict > img")
    path, _ := img.Attr("src")
    outInfo.Image[listIdx] = trimImageFile(path)

    // 最高気温
    max_temp := selection.Find(".temp > .high > em")
    outInfo.MaxTemperature[listIdx], _ = strconv.Atoi(max_temp.Text())

    // 最低気温
    min_temp := selection.Find(".temp > .low > em")
    outInfo.MinTemperature[listIdx], _ = strconv.Atoi(min_temp.Text())
}

func parseWeek(doc *goquery.Document, outInfo *Info) {
    yjw_week := doc.Find("div#yjw_week > table.yjw_table > tbody > tr")
    yjw_week.Each(func(rowIdx int, row *goquery.Selection) {
        row.Find("td").Each(func(i int, s *goquery.Selection) {
            if i >= 1 && i <= 5{
                listIdx := i + 1;

                switch(rowIdx) {
                    // 日付
                    case 0: {
                        small := s.Find("small").First()
                        outInfo.Date[listIdx] = trimDateAsInt(small.Text())
                    }
                    // 天気マーク
                    case 1: {
                        path, _ := s.Find("img").First().Attr("src")
                        outInfo.Image[listIdx] = trimImageFile(path)
                    }
                    // 気温
                    case 2: {
                        small := s.Find("small > font")
                        outInfo.MinTemperature[listIdx], _ = strconv.Atoi(small.First().Text())
                        outInfo.MaxTemperature[listIdx], _ = strconv.Atoi(small.Last().Text())
                    }
                }
            }
        })
    })
}

func trimDateAsInt(dateStr string) int {
    dateStr = strings.Split(dateStr, "日")[0]
    dateStr = strings.Split(dateStr, "月")[1]
	value, _ := strconv.Atoi(dateStr)
	return value
}

func trimImageFile(pathStr string) string {
    splited := strings.Split(pathStr, "/")
    return splited[len(splited) - 1]
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
