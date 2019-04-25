package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Info jsonに出力する週間天気情報
type Info struct {
	City           string   `json:"city"`
	MaxTemperature []string `json:"maxTemp"`
	MinTemperature []string `json:"minTemp"`
	Image          []string `json:"image"`
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
			s.Parent().Find(".maxtemp").Each(func(_ int, s *goquery.Selection) {
				temp := strings.TrimSpace(s.Text())
				temps := strings.Split(temp, "\n")
				weatherInfo.MaxTemperature = append(weatherInfo.MaxTemperature, temps[0])
				isSetMax = true
			})
		}

		if !isSetMin && strings.Contains(s.Text(), "最低") {
			s.Parent().Find(".mintemp").Each(func(_ int, s *goquery.Selection) {
				temp := strings.TrimSpace(s.Text())
				temps := strings.Split(temp, "\n")
				weatherInfo.MinTemperature = append(weatherInfo.MinTemperature, temps[0])
				isSetMin = true
			})
		}
	})

	isSetImg := false
	doc.Find(".normal").Each(func(_ int, s *goquery.Selection) {
		if !isSetImg && strings.Contains(s.Text(), "東京地方") {
			s.Parent().Find("img").Each(func(_ int, s *goquery.Selection) {
				imagePath, _ := s.Attr("src")
				imagePath = "https://www.jma.go.jp/jp/week/" + imagePath
				weatherInfo.Image = append(weatherInfo.Image, imagePath)
			})
			isSetImg = true
		}
	})

	jsonBytes, _ := json.Marshal(weatherInfo)
	jsonString := string(jsonBytes)
	//print(jsonString)

	file, err := os.Create("out/weather.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.Write(([]byte)(jsonString))
}
