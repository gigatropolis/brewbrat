package calc

import (
	"fmt"
	"math"
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
		message = fmt.Sprintf("Original Gravity = %.4f", og)

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

	} else if words[1] == "refractotofg" {
		if len(words) < 4 {
			return "", nil
		}
		OriginalBrix, err := strconv.ParseFloat(words[2], 64)
		FinalBrix, err2 := strconv.ParseFloat(words[3], 64)

		if err != nil || err2 != nil {
			return "", nil
		}
		finalGrav, berr := RefractoFg(OriginalBrix, FinalBrix)
		if berr != nil {
			return "", nil
		}
		og, _ := BrixToOg(OriginalBrix)
		abv, _ := Abv(og, finalGrav)
		message = fmt.Sprintf("Final calculated gravity = %.4f with ABV of %.2f%%", finalGrav, abv)
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

func _refN1(OrigBrix, FinalBrix float64) float64 {
	return 1.001843 - 0.002318474*OrigBrix - 0.000007775*(math.Pow(OrigBrix, 2.0)) - 0.000000034*(math.Pow(OrigBrix, 3.0)) + 0.00574*FinalBrix + 0.00003344*(math.Pow(FinalBrix, 2.0)) + 0.000000086*(math.Pow(FinalBrix, 3.0))
}

func _refN2(OrigBrix, FinalBrix float64) float64 {
	return 1.000898 + 0.003859118*OrigBrix + 0.00001370735*(math.Pow(OrigBrix, 2.0)) + 0.00000003742517*(math.Pow(OrigBrix, 3.0))
}

func _refN3(OrigBrix, FinalBrix float64) float64 {
	return 668.72*_refN2(OrigBrix, FinalBrix) - 463.37 - 205.347*(math.Pow(_refN2(OrigBrix, FinalBrix), 2.0))
}

/*
	RefractoFg will convert refractometer Final gravity from original and final Brix readings from refractometer
    http://seanterrill.com/2010/07/20/toward-a-better-refractometer-correlation/

        OrigBrix (RIi)  =   Original Brix reading
        FinalBrix (RIf) =   Final Brix Reading from refractometer
        FG =                Final Gravity measured
        FG = (1.001843 – 0.002318474*RIi – 0.000007775*RIi² – 0.000000034*RIi³ + 0.00574*RIf + 0.00003344*RIf² + 0.000000086*RIf³) + 0.0216*LN(1 –
            (0.1808*(668.72*(1.000898 + 0.003859118*RIi + 0.00001370735*RIi² + 0.00000003742517*RIi³) – 463.37 –
            205.347*(1.000898 + 0.003859118*RIi + 0.00001370735*RIi² + 0.00000003742517*RIi³)²) +
            0.8192*(668.72*(1.001843 – 0.002318474*RIi – 0.000007775*RIi² – 0.000000034*RIi³ + 0.00574*RIf + 0.00003344*RIf² + 0.000000086*RIf³) – 463.37 –
            205.347*(1.001843 – 0.002318474*RIi – 0.000007775*RIi² – 0.000000034*RIi³ + 0.00574*RIf + 0.00003344*RIf² + 0.000000086*RIf³)²)) /
            (668.72*(1.000898 + 0.003859118*RIi + 0.00001370735*RIi² + 0.00000003742517*RIi³) – 463.37 –
            205.347*(1.000898 + 0.003859118*RIi + 0.00001370735*RIi² + 0.00000003742517*RIi³)²)) + 0.0116
*/
func RefractoFg(OrigBrix, FinalBrix float64) (float64, error) {
	refN1 := _refN1(OrigBrix, FinalBrix)
	refN2 := _refN2(OrigBrix, FinalBrix)
	refN3 := _refN3(OrigBrix, FinalBrix)

	FG := refN1 + 0.0216*math.Log(1-(0.1808*(668.72*refN2-463.37-205.347*math.Pow(refN2, 2.0))+0.8192*(668.72*refN1-463.37-205.347*math.Pow(refN1, 2.0)))/refN3) + 0.0116
	return FG, nil
}
