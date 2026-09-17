# CDC BMI-for-age reference

`cdc_bmi_age_data.go` is generated from the official CDC 2000 BMI-for-age
LMS table for ages 2 through 19 years:

https://www.cdc.gov/growthcharts/data/zscore/bmiagerev.csv

The table contains monthly, sex-specific LMS parameters and the 5th, 85th and
95th percentiles. Regenerate it from the project backend directory with:

```sh
go run ./tools/generate_cdc_bmi
```
