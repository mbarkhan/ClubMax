package main

import (
	"encoding/csv"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const sourceURL = "https://www.cdc.gov/growthcharts/data/zscore/bmiagerev.csv"

func main() {
	response, err := http.Get(sourceURL)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		panic(response.Status)
	}
	rows, err := csv.NewReader(io.LimitReader(response.Body, 2<<20)).ReadAll()
	if err != nil {
		panic(err)
	}
	var output strings.Builder
	output.WriteString("// Code generated from the CDC 2000 BMI-for-age LMS data; DO NOT EDIT.\n")
	output.WriteString("// Source: " + sourceURL + "\npackage reference\n\n")
	output.WriteString("type BMIAgeRow struct { Month, L, M, S, P5, P85, P95 float64 }\n\n")
	for sex := 1; sex <= 2; sex++ {
		fmt.Fprintf(&output, "var cdcBMIAgeSex%d = []BMIAgeRow{\n", sex)
		for _, row := range rows[1:] {
			value, _ := strconv.Atoi(row[0])
			month, _ := strconv.ParseFloat(row[1], 64)
			if value != sex || month < 24.5 || month > 239.5 {
				continue
			}
			fmt.Fprintf(&output, "{%s,%s,%s,%s,%s,%s,%s},\n", row[1], row[2], row[3], row[4], row[6], row[11], row[13])
		}
		output.WriteString("}\n\n")
	}
	output.WriteString(`func CDCBMIAgeRow(sex, completedMonths int) (BMIAgeRow, bool) {
	if completedMonths < 24 || completedMonths > 239 { return BMIAgeRow{}, false }
	rows := cdcBMIAgeSex1
	if sex == 2 { rows = cdcBMIAgeSex2 } else if sex != 1 { return BMIAgeRow{}, false }
	index := completedMonths - 24
	if index < 0 || index >= len(rows) { return BMIAgeRow{}, false }
	return rows[index], true
}
`)
	formatted, err := format.Source([]byte(output.String()))
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile("internal/reference/cdc_bmi_age_data.go", formatted, 0644); err != nil {
		panic(err)
	}
}
