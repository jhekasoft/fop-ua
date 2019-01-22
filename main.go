package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
)

// our main function
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/calendar", GetCalendar).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func GetCalendar(w http.ResponseWriter, r *http.Request) {
	monthTitles, err := parseMonthes("data/calendar/monthes.yml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	// fmt.Printf("data: %v\n", monthTitles)

	// Read ESV calendar
	esv, err := parseCalendarData("data/calendar/2019/fop_single/esv.yml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	// fmt.Printf("data: %v\n", esv)

	// Read Group 1, 2 calendar
	// taxGroup12, err := parseCalendarData("data/calendar/2019/fop_single/tax_group1_2.yml")
	// if err != nil {
	// 	fmt.Printf("err: %v\n", err)
	// 	return
	// }
	// fmt.Printf("data: %v\n", taxGroup12)

	// Read Group 3 calendar
	taxGroup3, err := parseCalendarData("data/calendar/2019/fop_single/tax_group3.yml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	// fmt.Printf("data: %v\n", taxGroup3)

	data, err := combineCalendarData(monthTitles, append(esv, taxGroup3...), 3, false)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	json.NewEncoder(w).Encode(data)
}

func combineCalendarData(monthTitles map[int]string, data []InputFopCalendarSingleItem, group int, isWithPdv bool) (result []Month, err error) {
	allMonthItems := []MonthItem{}

	for _, inputCalendarItem := range data {
		endData, err := time.Parse("02.01.2006", inputCalendarItem.EndDate)
		if err != nil {
			return nil, err
		}

		var amountMax float32
		if group == 1 {
			amountMax = inputCalendarItem.AmountMaxGroup1
		} else if group == 2 {
			amountMax = inputCalendarItem.AmountMaxGroup2
		}

		var amountPercents int
		if group == 3 && isWithPdv {
			amountPercents = inputCalendarItem.AmountPercentsPdv
		} else if group == 3 && !isWithPdv {
			amountPercents = inputCalendarItem.AmountPercentsNoPdv
		}

		allMonthItems = append(allMonthItems, MonthItem{
			EndDate:        endData,
			Title:          inputCalendarItem.Title,
			Amount:         inputCalendarItem.Amount,
			AmountMax:      amountMax,
			AmountPercents: amountPercents,
		})
	}

	for monthNum, monthTitle := range monthTitles {
		monthItems := []MonthItem{}

		for _, item := range allMonthItems {
			if int(item.EndDate.Month()) == monthNum {
				monthItems = append(monthItems, item)
			}
		}

		month := Month{
			Number: monthNum,
			Title:  monthTitle,
			Items:  monthItems,
		}
		result = append(result, month)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Number < result[j].Number
	})

	return
}

func parseCalendarData(path string) (parsedData []InputFopCalendarSingleItem, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &parsedData)
	if err != nil {
		return
	}

	return parsedData, nil
}

func parseMonthes(path string) (parsedData map[int]string, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &parsedData)
	if err != nil {
		return
	}

	return parsedData, nil
}

// Input data
// TODO: move to the other file
type InputFopCalendarSingleItem struct {
	EndDate             string  `json:"end_date"`
	Type                string  `json:"type"`
	Title               string  `json:"title"`
	Amount              float32 `json:"amount"`
	AmountMaxGroup1     float32 `json:"amount_max_group1"`
	AmountMaxGroup2     float32 `json:"amount_max_group2"`
	AmountPercentsPdv   int     `json:"amount_percents_pdv"`
	AmountPercentsNoPdv int     `json:"amount_percents_no_pdv"`
}

// Output data
// TODO: move to the other file
type Month struct {
	Number int         `json:"number,omitempty"`
	Title  string      `json:"title,omitempty"`
	Items  []MonthItem `json:"items,omitempty"`
}
type MonthItem struct {
	EndDate        time.Time `json:"end_date,omitempty"`
	Title          string    `json:"title,omitempty"`
	Amount         float32   `json:"amount,omitempty"`
	AmountMax      float32   `json:"amount_max,omitempty"`
	AmountPercents int       `json:"amount_percents,omitempty"`
}
