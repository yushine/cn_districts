package main

import (
	"McaZones/models"
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/k0kubun/pp"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"
)

var db *sql.DB

func main() {
	_, _ = pp.Println("Start load excel ...")

	var err = errors.New("")
	db, err = sql.Open("mysql", "root:admin123my@tcp(192.168.0.112:3306)/beahu_api_development")
	if err != nil {
		_, _ = pp.Println(err.Error())
	}

	//sheets := getSheets()
	//clearData()
	//handleData(sheets)

	updateLocation()
}

func getSheets() []*xlsx.Sheet {
	file := "./data/2019-7-cn-regions.xlsx"
	xlFile, err := xlsx.OpenFile(file)
	if err != nil {
		_, _ = pp.Println(err.Error())
		return nil
	}

	return xlFile.Sheets
}

func clearData() {
	result, err := db.Exec("TRUNCATE TABLE districts")
	_, _ = pp.Println(result, err)
}

func handleData(sheets []*xlsx.Sheet) {
	for _, sheet := range sheets {
		sheetName := sheet.Name
		if sheetName != "processed" {
			continue
		}

		rowIndex := 0
		lastLevel := 1
		parentId := 0
		lastZoneId := 0
		var lastInsertId int64
		for _, row := range sheet.Rows {
			rowIndex = rowIndex + 1
			if rowIndex == 1 || rowIndex == 2 || rowIndex == 3 {
				continue
			}

			district := &models.District{}
			firstCell := strings.TrimSpace(row.Cells[0].Value)
			secondCell := strings.TrimSpace(row.Cells[1].Value)
			thirdCell := strings.TrimSpace(row.Cells[2].Value)
			fourthCell := strings.TrimSpace(row.Cells[3].Value)
			if firstCell != "" {
				district.Code = firstCell
				district.Name = secondCell
				district.Level = 1
			} else if firstCell == "" && secondCell != "" {
				district.Code = secondCell
				district.Name = thirdCell
				district.Level = 2
			} else if firstCell == "" && secondCell == "" && fourthCell != "" {
				district.Code = thirdCell
				district.Name = fourthCell
				district.Level = 3
			} else {
				continue
			}

			if district.Level == 1 {
				parentId = 0
			} else if district.Level == 2 {
				parentId = lastZoneId
			} else if lastLevel == 2 && district.Level == 3 {
				parentId = int(lastInsertId)
			}

			result, err := db.Exec(
				"INSERT INTO districts (country_id, name, level, up_id, code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				"44",
				district.Name,
				district.Level,
				parentId,
				district.Code,
				time.Now(),
				time.Now(),
			)
			lastLevel = district.Level
			lastInsertId, _ = result.LastInsertId()

			if district.Level == 1 {
				lastZoneId = int(lastInsertId)
			}

			_, _ = pp.Println(district, result, err)
		}
	}
}

func updateLocation() {
	locations := getTXLocation()
	items := locations["result"]
	_, _ = pp.Println(reflect.TypeOf(items).Kind().String())

	for index, levelOne := range items.([]interface{}) {
		_, _ = pp.Println(index)
		for index2, levelTwo := range levelOne.([]interface{}) {
			_, _ = pp.Println(index2)
			levelTwo2 := levelTwo.(map[string]interface{})
			location := levelTwo2["location"].(map[string]interface{})

			code := levelTwo2["id"]
			latitude := location["lat"]
			longitude := location["lng"]

			_, _ = pp.Println(code, latitude, longitude)
			result, err := db.Exec("UPDATE districts set longitude=?, latitude=? where code=?", longitude, latitude, code)
			if err != nil {
				_, _ = pp.Println(result, err)
				return
			}
		}
	}
}

/**
https://lbs.qq.com/webservice_v1/guide-region.html
*/
func getTXLocation() map[string]interface{} {
	file := "./data/qq-map-regions.json"
	jsonFile, err := os.Open(file)

	// if we os.Open returns an error then handle it
	if err != nil {
		_, _ = pp.Println(err)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	_ = json.Unmarshal([]byte(byteValue), &result)
	return result
}
