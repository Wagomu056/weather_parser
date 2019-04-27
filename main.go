package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Info jsonに出力する週間天気情報
type Info struct {
	City           string    `json:"city"`
	MaxTemperature [7]int    `json:"maxTemp"`
	MinTemperature [7]int    `json:"minTemp"`
	Image          [7]string `json:"image"`
}

func main() {
	doc, err := goquery.NewDocument("https://www.jma.go.jp/jp/week/319.html")
	if err != nil {
		panic(err)
	}

	var weatherInfo Info
	// 都市名
	weatherInfo.City = doc.Find(".cityname").First().Text()

	// 最高、最低気温
	isSetMax := false
	isSetMin := false
	doc.Find(".citypro").Each(func(_ int, s *goquery.Selection) {
		if !isSetMax && strings.Contains(s.Text(), "最高") {
			index := 0
			s.Parent().Find(".maxtemp").Each(func(_ int, s *goquery.Selection) {
				temp := strings.TrimSpace(s.Text())
				temps := strings.Split(temp, "\n")
				intTemp, _ := strconv.Atoi(temps[0])
				weatherInfo.MaxTemperature[index] = intTemp
				index++
			})

			isSetMax = true
		}

		if !isSetMin && strings.Contains(s.Text(), "最低") {
			index := 0
			s.Parent().Find(".mintemp").Each(func(_ int, s *goquery.Selection) {
				temp := strings.TrimSpace(s.Text())
				temps := strings.Split(temp, "\n")
				intTemp, _ := strconv.Atoi(temps[0])
				weatherInfo.MinTemperature[index] = intTemp
				index++
			})

			// 6日分しか記録されていない場合というのは、今日の最低気温がない場合(気象庁のサイトの仕様)。
			// 先頭に不正な値を入力する
			if index == 6 {
				for i := 6; i > 0; i-- {
					weatherInfo.MinTemperature[i] = weatherInfo.MinTemperature[i-1]
				}
				weatherInfo.MinTemperature[0] = 99
			}

			isSetMin = true
		}
	})

	isSetImg := false
	doc.Find(".normal").Each(func(_ int, s *goquery.Selection) {
		if !isSetImg && strings.Contains(s.Text(), "東京地方") {
			index := 0
			s.Parent().Find("img").Each(func(_ int, s *goquery.Selection) {
				imagePath, _ := s.Attr("src")
				imagePath = "https://www.jma.go.jp/jp/week/" + imagePath
				weatherInfo.Image[index] = imagePath
				index++
			})

			isSetImg = true
		}
	})

	jsonBytes, _ := json.Marshal(weatherInfo)
	jsonString := string(jsonBytes)

	//file, err := os.Create("out/weather.json")
	file, err := os.Create("/home/tsubakuro/Projects/Go/Web/weather/data/tokyo.json")

	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.Write(([]byte)(jsonString))
}
