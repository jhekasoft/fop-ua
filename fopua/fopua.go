package fopua

import (
	"fmt"
	"io/ioutil"
	"sort"
	"time"

	"github.com/ghodss/yaml"
)

// GetFopSingleCalendar returns calendar data for UA FOP with single tax
func GetFopSingleCalendar(dataPath string, group int, withPdv bool) (data []Month, err error) {
	if group > 3 || group < 1 {
		group = 1
	}

	monthTitles, err := parseMonthes(dataPath + "/calendar/monthes.yml")
	if err != nil {
		return
	}
	// fmt.Printf("data: %v\n", monthTitles)

	// Read ESV calendar
	esvItems, err := parseCalendarData(dataPath + "/calendar/2019/fop_single/esv.yml")
	if err != nil {
		return
	}

	taxItems := []InputFopCalendarSingleItem{}
	if group == 3 {
		// Read Group 3 calendar
		taxItems, err = parseCalendarData(dataPath + "/calendar/2019/fop_single/tax_group3.yml")
		if err != nil {
			return nil, err
		}
	} else {
		// Read Group 1, 2 calendar
		taxItems, err = parseCalendarData(dataPath + "/calendar/2019/fop_single/tax_group1_2.yml")
		if err != nil {
			return nil, err
		}
	}

	dataItems := append(esvItems, taxItems...)
	data, err = combineCalendarData(monthTitles, dataItems, group, withPdv)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	return
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

	// Items to month
	for monthNum, monthTitle := range monthTitles {
		monthItems := []MonthItem{}

		for _, item := range allMonthItems {
			if int(item.EndDate.Month()) == monthNum {
				monthItems = append(monthItems, item)
			}
		}

		// Sort month items
		sort.Slice(monthItems, func(i, j int) bool {
			return monthItems[i].EndDate.Before(monthItems[j].EndDate)
		})

		month := Month{
			Number: monthNum,
			Title:  monthTitle,
			Items:  monthItems,
		}
		result = append(result, month)
	}

	// Sort monthes
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

// InputFopCalendarSingleItem represents calendar data in YML
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

// Month represents result calendar data
type Month struct {
	Number int         `json:"number,omitempty"`
	Title  string      `json:"title,omitempty"`
	Items  []MonthItem `json:"items,omitempty"`
}

// MonthItem represents result calendar month data
type MonthItem struct {
	EndDate        time.Time `json:"end_date,omitempty"`
	Title          string    `json:"title,omitempty"`
	Amount         float32   `json:"amount,omitempty"`
	AmountMax      float32   `json:"amount_max,omitempty"`
	AmountPercents int       `json:"amount_percents,omitempty"`
}
