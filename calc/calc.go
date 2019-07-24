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

// HandleCommand calls appropriate functions with parameters used and returns response
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

	} else if words[1] == "refractotofg" || words[1] == "fg" {
		if len(words) < 4 {
			return "refractotofg requires 2 arguments <Og> <measured Fg>", nil
		}
		OriginalBrix, err := strconv.ParseFloat(words[2], 64)
		FinalBrix, err2 := strconv.ParseFloat(words[3], 64)

		if err != nil || err2 != nil {
			return "", nil
		}

		if OriginalBrix < 2 {
			OriginalBrix, _ = OgToBrix(OriginalBrix)
		}

		if FinalBrix < 1.06 {
			FinalBrix, _ = OgToBrix(FinalBrix)
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

// BrixToOg converts Brix to Gravity
func BrixToOg(brix float64) (float64, error) {
	og := (brix / (258.6 - ((brix / 258.2) * 227.1))) + 1

	return og, nil
}

// OgToBrix converts Gravity to Brix
func OgToBrix(og float64) (float64, error) {
	brix := (((182.4601*og-775.6821)*og+1262.7794)*og - 669.5622)

	return brix, nil
}

// Abv returns Alcohol By Volume in percentage
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

/*
GetPercentUtilization used to Calculate the correct percent Utilization
based on time in boil
boilTime - Boil time in minutes
Return the Utilization in percent
*/
func GetPercentUtilization(boilTime int32) float64 {
	var percUtil float64
	if boilTime < 6 {
		percUtil = 5.0
	} else if boilTime < 11 {
		percUtil = 6.0
	} else if boilTime < 16 {
		percUtil = 8.0
	} else if boilTime < 21 {
		percUtil = 10.1
	} else if boilTime < 26 {
		percUtil = 12.1
	} else if boilTime < 31 {
		percUtil = 15.3
	} else if boilTime < 41 {
		percUtil = 22.8
	} else if boilTime < 46 {
		percUtil = 26.9
	} else if boilTime < 51 {
		percUtil = 28.1
	} else {
		percUtil = 30.0
	}

	return percUtil
}

// GetBoilGravity calculates boil gravity when doing a partial boil
// batchSize - actual batch size in gallons
// boilQnty - Size of boil size in gallons
// origGravity - calculated original gravity for batch
func GetBoilGravity(batchSize, boilQnty, origGravity float64) (float64, error) {
	if boilQnty >= batchSize {
		return origGravity, nil
	}
	return ((origGravity - 1) * (batchSize / boilQnty)) + 1, nil
}

// GetGravityAdjustment returns gravity adjustment for calculating hop utilization
// boilGravity - boil gravity in gravity reading
func GetGravityAdjustment(boilGravity float64) (float64, error) {
	if boilGravity < 1.050 {
		return 0, nil
	}
	ga := (boilGravity - 1.050) / 0.2
	return ga, nil
}

// GetIBU returns amount of IBUs for hop addition
// waterQuantity - Batch size in gallons
// boilTime - Boil time for hop in minutes
// amount - quantity of hops in ounces
// percAlpha - Alpha acids in percentage for hop being used
// OriginalGrav - gravity of batch during boil
// boilQuantity - amount of water in gallons used in boil
func GetIBU(waterQuantity float64, boilTime int32, amount int32, percAlpha float64, OriginalGrav float64, boilQuantity float64) (int32, error) {
	Utilization := GetPercentUtilization(boilTime) / 100.0
	boilGravity, err := GetBoilGravity(waterQuantity, boilQuantity, OriginalGrav)
	if err != nil {
		return 0.0, nil
	}
	gravAdjustment, err2 := GetGravityAdjustment(boilGravity)
	if err2 != nil {
		return 0.0, nil
	}
	ibu := (float64(amount) * Utilization * (percAlpha / 100) * 7462.0) / (waterQuantity * (1 + gravAdjustment))
	return int32(ibu), nil
}

// DefHopsForDesiredIBU gets amount of hops needed for a hop with known alpha acid for desired IBU at a perticular boil time
// waterQuantity - Batch size in gallons
// boilTime - Boil time for hop in minutes
// percAlpha - Alpha acids in percentage for hop being used
// boilGravity - gravity of batch during boil
func DefHopsForDesiredIBU(desiredIBU int32, waterQuantity float64, boilTime int32, percAlpha float64, boilGravity float64) (float64, error) {
	Utilization := GetPercentUtilization(boilTime) / 100.0
	gravAdjustment, err := GetGravityAdjustment(boilGravity)
	if err != nil {
		return 0.0, nil
	}
	hopAmountForIBU := (waterQuantity * (1 + gravAdjustment) * float64(desiredIBU)) / (Utilization * (percAlpha / 100.0) * 7462.0)
	return hopAmountForIBU, nil
}

// KilToOz will convert Kilograms to Onces
func KilToOz(kilAmount float64) (float64, error) {
	return kilAmount / 0.02834952, nil
}

// KilToGal will convert Kilograms to gallons
func KilToGal(kilAmount float64) (float64, error) {
	return kilAmount * 0.26417, nil
}

// KilToLb will convert Kilograms to pounds
func KilToLb(kilAmount float64) (float64, error) {
	return kilAmount / 0.453592374, nil
}

// McuToSrm will convert MCU color units to SRM color units
func McuToSrm(mcuAmount float64) (float64, error) {
	return 1.4922 * math.Pow(mcuAmount, 0.6859), nil
}
