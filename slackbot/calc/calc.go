package calc

import (
	"fmt"
	"math"
	"strconv"
	//"io/ioutil"
	//"log"
	//"os"
	//"strings"
)

// HandleCommand calls appropriate functions with parameters used and returns response
func HandleCommand(words []string, orMes string) (string, error) {
	message := ""

	switch {
	case words[1] == "brixtoog" || words[1] == "og":
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

	case words[1] == "ogtobrix" || words[1] == "brix":
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

	case words[1] == "abv":
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

	case words[1] == "refractotofg" || words[1] == "fg":
		if len(words) < 4 {
			return "refractotofg requires at least 2 arguments <Og> <measured Fg> [opt: temperature]", nil
		}
		OriginalBrix, err := strconv.ParseFloat(words[2], 64)
		FinalBrix, err2 := strconv.ParseFloat(words[3], 64)
		var temp float64 = 0.0
		if len(words) > 4 {
			temp, _ = strconv.ParseFloat(words[4], 64)
		}
		if err != nil || err2 != nil {
			return "", nil
		}

		if OriginalBrix < 2 {
			OriginalBrix, _ = OgToBrix(OriginalBrix)
		}

		if FinalBrix < 1.06 {
			FinalBrix, _ = OgToBrix(FinalBrix)
		}

		finalGrav, adjFinalBrix, berr := RefractoFg(OriginalBrix, FinalBrix, temp)
		if berr != nil {
			return "", nil
		}

		og, _ := BrixToOg(OriginalBrix)
		abv, _ := Abv(og, finalGrav)

		message = fmt.Sprintf("Final calculated gravity = %.4f with ABV of %.2f%%  (Adjusted final Brix = %0.4f)", finalGrav, abv, adjFinalBrix)

	default:
		message = "Unrecognized command '" + words[1] + "'"
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

/*
RefractoFg will convert refractometer Final gravity from original and final Brix readings from refractometer

   Calculations from MoreBeer excel spreadsheet "THJ Refractometer_Beer.xls"

  SG = 1.001843 - 0.002318474(OB) - 0.000007775(OB^2) - 0.000000034(OB^3) + 0.00574(AB) + 0.00003344(AB^2) + 0.000000086(AB^3)
  temp adjust = (1.313454-0.132674*temp+0.002057793*(temp^2)-0.000002627634*(temp^3))*0.001
  adjusted final Brix = -676.67+1286.4*Fg-800.47*(Fg^2)+190.74*(Fg^3)
  OrigBrix (RIi)  =   Original Brix reading
  FinalBrix (RIf) =   Final Brix Reading from refractometer
  FG =                Final Gravity measured
*/
func RefractoFg(OrigBrix float64, FinalBrix float64, temp float64) (float64, float64, error) {
	var calcTemp float64 = 0.0
	var adjFinalBrix float64 = 0.0

	Fg := 1.001843 - 0.002318474*(OrigBrix) - 0.000007775*(math.Pow(OrigBrix, 2.0)) - 0.000000034*(math.Pow(OrigBrix, 3.0)) + 0.00574*(FinalBrix) + 0.00003344*(math.Pow(FinalBrix, 2)) + 0.000000086*(math.Pow(FinalBrix, 3))

	if temp > 10 {
		calcTemp = (1.313454 - 0.132674*temp + 0.002057793*(math.Pow(temp, 2.0)) - 0.000002627634*(math.Pow(temp, 3.0))) * 0.001
		fmt.Printf("Calculated temperature adjustent = %0.5f\n", calcTemp)
		Fg += calcTemp
	}

	adjFinalBrix = -676.67 + 1286.4*Fg - 800.47*(math.Pow(Fg, 2.0)) + 190.74*(math.Pow(Fg, 3.0))

	return Fg, adjFinalBrix, nil
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
RefractoFgComplex will convert refractometer Final gravity from original and final Brix readings from refractometer

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
func RefractoFgComplex(OrigBrix, FinalBrix float64) (float64, error) {
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

	switch {
	case boilTime < 6:
		percUtil = 5.0
	case boilTime < 11:
		percUtil = 6.0
	case boilTime < 16:
		percUtil = 8.0
	case boilTime < 21:
		percUtil = 10.1
	case boilTime < 26:
		percUtil = 12.1
	case boilTime < 31:
		percUtil = 15.3
	case boilTime < 41:
		percUtil = 22.8
	case boilTime < 46:
		percUtil = 26.9
	case boilTime < 51:
		percUtil = 28.1
	default:
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
