package calc

import (
	"fmt"
	"strconv"
	//"fmt"
	//"io/ioutil"
	//"log"
	//"os"
	//"strings"
)

func HandleCommand(words []string) (string, error) {
	message := ""

	if words[1] == "brixtoog" {
		if len(words) < 3 {
			return "", nil
		}
		brix, err := strconv.ParseFloat(words[2], 64)
		if err != nil {
			return "", nil
		}
		og, berr := BrixToOg(brix)
		if berr != nil {
			return "", nil
		}
		message = fmt.Sprintf("Original Gravity = %.5f", og)

	} else if words[1] == "ogtobrix" {
		if len(words) < 3 {
			return "", nil
		}
		og, err := strconv.ParseFloat(words[2], 64)
		if err != nil {
			return "", nil
		}
		brix, berr := OgToBrix(og)
		if berr != nil {
			return "", nil
		}
		message = fmt.Sprintf("Brix = %.2f", brix)

	} else if words[1] == "abv" {
		if len(words) < 4 {
			return "", nil
		}
		og, err := strconv.ParseFloat(words[2], 64)
		fg, err2 := strconv.ParseFloat(words[3], 64)

		if err != nil || err2 != nil {
			return "", nil
		}
		abv, berr := Abv(og, fg)
		if berr != nil {
			return "", nil
		}
		message = fmt.Sprintf("ABV = %.2f%%", abv)

	}
	return message, nil
}

func BrixToOg(brix float64) (float64, error) {
	og := (brix / (258.6 - ((brix / 258.2) * 227.1))) + 1

	return og, nil
}

func OgToBrix(og float64) (float64, error) {
	brix := (((182.4601*og-775.6821)*og+1262.7794)*og - 669.5622)

	return brix, nil
}

func Abv(og, fg float64) (float64, error) {
	abv := (og - fg) * 131.25

	return abv, nil
}
