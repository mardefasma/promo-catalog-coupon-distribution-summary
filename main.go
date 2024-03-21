package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// Adjust query to export the csv
// 
// select c.base_code, r.claimed, r.redeem_date
// from "catalog" c
// join (select catalog_id,count(id) as claimed,max(create_time) as redeem_date
// from redeem r where
// redeem_create_time >= 20240325
// and r.create_time >= '2024-03-25 00:00:00'
// group by catalog_id) r on c.id=r.catalog_id
// where c.base_code in ('KK25MARAA','KK25MARAB','KK25MARAC','KK25MARAD','KK25MARAE','KK25MARAF','BLJD11HGN','BLJD10HGN','DUGCSRE35')
// order by c.slug

const (
	InputFileNameCSV = "query_result_2024-03-21T14_15_37.264296+07_00.csv"
)

var (
	PromoDetailMapByBaseCode = map[string]PromoDetail{
		// Dummy
		"BBI03BCC": {
			Benefit:    "Diskon 12% up to 100k",
			LimitPerTW: 95000,
		},

		// 25 Mar
		"KK25MARAA": {
			Benefit:    "CB 100% up tp 50k",
			LimitPerTW: 1000,
		},
		"KK25MARAB": {
			Benefit:    "CB 50% up tp 100k",
			LimitPerTW: 1000,
		},
		"KK25MARAC": {
			Benefit:    "CB 20% up tp 200k",
			LimitPerTW: 1000,
		},
		"KK25MARAD": {
			Benefit:    "CB 10% up tp 500k",
			LimitPerTW: 1000,
		},
		"KK25MARAE": {
			Benefit:    "CB 5% up tp 1jt",
			LimitPerTW: 1000,
		},
		"KK25MARAF": {
			Benefit:    "CB 3% up tp 5jt",
			LimitPerTW: 1000,
		},
		"BLJD11HGN": {
			Benefit:    "Diskon 100% up to 50K",
			LimitPerTW: 1000,
		},
		"BLJD10HGN": {
			Benefit:    "Diskon 100% up to 40K",
			LimitPerTW: 1000,
		},
		"DUGCSRE35": {
			Benefit:    "Diskon 100% up to 35K",
			LimitPerTW: 1000,
		},
	}
)

type PromoDetail struct {
	Benefit    string
	LimitPerTW int64
}

type QueryExport struct {
	BaseCode       string
	TotalClaimed   int64
	LastRedeemTime time.Time
}

func createQueryExport(data [][]string) []QueryExport {
	var newRelicLog []QueryExport
	for i, line := range data {
		if i > 0 { // omit header line
			var rec QueryExport
			for j, field := range line {
				// 0 "base_code"
				// 1 "total_claimed"
				// 2 "last_redeem"
				switch j {
				case 0:
					rec.BaseCode = field

				case 1:
					totalClaimed, _ := strconv.ParseInt(field, 10, 64)
					rec.TotalClaimed = totalClaimed

				case 2:
					lastRedeemTime, _ := time.Parse(time.RFC3339, field)
					rec.LastRedeemTime = lastRedeemTime
				}
			}
			newRelicLog = append(newRelicLog, rec)
		}
	}
	return newRelicLog
}

func main() {
	// open file
	f, err := os.Open(InputFileNameCSV)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// convert records to array of structs
	queryExportList := createQueryExport(data)

	var ongoingClaimSummary, fullyClaimSummary []string
	for _, queryExport := range queryExportList {
		// get information name and max limit per tw
		promoDetail, ok := PromoDetailMapByBaseCode[queryExport.BaseCode]
		if !ok {
			continue
		}

		// get percentage from redeem
		percentageClaimProgress := 100 * (float64(queryExport.TotalClaimed) / float64(promoDetail.LimitPerTW))

		// filter group by fully redeem
		if percentageClaimProgress < 100.0 {
			// construct for ongoing claim
			tempStr := fmt.Sprintf("%s (quota: %d) || %.2f%%", promoDetail.Benefit, promoDetail.LimitPerTW, percentageClaimProgress)

			ongoingClaimSummary = append(ongoingClaimSummary, tempStr)
		} else {
			// construct for fully claim
			tempStr := fmt.Sprintf("%s (quota: %d) || Last redeem: %s", promoDetail.Benefit, promoDetail.LimitPerTW, queryExport.LastRedeemTime.Format(time.Kitchen))

			fullyClaimSummary = append(fullyClaimSummary, tempStr)
		}
	}

	fmt.Println("Ongoing Coupon Claim")
	for idx, tempStr := range ongoingClaimSummary {
		fmt.Printf("%d. %s", idx+1, tempStr)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("Fully Coupon Claim")
	for idx, tempStr := range fullyClaimSummary {
		fmt.Printf("%d. %s\n", idx+1, tempStr)
	}
}
