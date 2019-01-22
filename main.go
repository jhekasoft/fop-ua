package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	// Read ESV calendar
	esv, err := parseCalendarData("data/calendar/2019/fop_single/esv.yml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Printf("data: %v\n", esv)

	// Read Group 1, 2 calendar
	taxGroup12, err := parseCalendarData("data/calendar/2019/fop_single/tax_group1_2.yml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Printf("data: %v\n", taxGroup12)

	// Read Group 3 calendar
	taxGroup3, err := parseCalendarData("data/calendar/2019/fop_single/tax_group3.yml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Printf("data: %v\n", taxGroup3)

	data, err := combineCalendarData(append(esv, taxGroup3...), 3, false)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	// t, err := time.Parse(layout, str)

	// var monthes []Month
	// monthes = append(monthes, Month{
	// 	Number: 1,
	// 	Title:  "Січень",
	// 	Items:  []MonthItem{MonthItem{EndDate: "01.01.2019", Title: "Test", Amount: "100"}},
	// })

	json.NewEncoder(w).Encode(data)
}

func combineCalendarData(data []InputFopSingleMonth, group int, isWithPdv bool) (result []Month, err error) {
	monthTitles := map[int]string{
		1:  "Січень",
		2:  "Лютий",
		3:  "Березень",
		4:  "Квітень",
		5:  "Травень",
		6:  "Червень",
		7:  "Липень",
		8:  "Серпень",
		9:  "Вересень",
		10: "Жовтень",
		11: "Листопад",
		12: "Грудень",
	}

	for i := 1; i <= 12; i++ {
		monthItems := []MonthItem{}

		for _, inputMonth := range data {
			if inputMonth.Num == i {
				for _, inputMonthItem := range inputMonth.Items {
					endData, err := time.Parse("02.01.2006", inputMonthItem.EndDate)
					if err != nil {
						return nil, err
					}

					var amountMax float32
					if group == 1 {
						amountMax = inputMonthItem.AmountMaxGroup1
					} else if group == 2 {
						amountMax = inputMonthItem.AmountMaxGroup2
					}

					var amountPercents int
					if group == 3 && isWithPdv {
						amountPercents = inputMonthItem.AmountPercentsPdv
					} else if group == 3 && !isWithPdv {
						amountPercents = inputMonthItem.AmountPercentsNoPdv
					}

					monthItems = append(monthItems, MonthItem{
						EndDate:        endData,
						Title:          inputMonthItem.Title,
						Amount:         inputMonthItem.Amount,
						AmountMax:      amountMax,
						AmountPercents: amountPercents,
					})
				}
			}
		}

		month := Month{
			Number: i,
			Title:  monthTitles[i],
			Items:  monthItems,
		}
		result = append(result, month)
	}

	return
}

func parseCalendarData(path string) (parsedData []InputFopSingleMonth, err error) {
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
type InputFopSingleMonth struct {
	Num   int                       `json:"num"`
	Items []InputFopSingleMonthItem `json:"items"`
}
type InputFopSingleMonthItem struct {
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
