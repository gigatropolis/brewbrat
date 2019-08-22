package ingredients

import (
	"github.com/gigatropolis/beercnv"

	"fmt"
	"os"
	"time"
	//"strconv"
	//"github.com/beerxml"
	//"github.com/stone/beerxml2"
	//"../xml/beerxml"
	//"../../beercnv"
)

var (
	data *beercnv.BeerXML2
)

// $.map($('.recipe-link'), m=>m.href).forEach(recipe=>{window.open(`${recipe}.xml`, '_blank')})

// Init inits
func Init() {
	data = &beercnv.BeerXML2{}
}

// GetBeerXMLFromFile takes a filename as string that has all beer related data in BeerXML 2.0 format
func GetBeerXMLFromFile(fileName string) (*beercnv.BeerXML2, error) {
	//allXML := beercnv.BeerXML{}
	//beer2 := beercnv.BeerXml2{}
	//path := "../recipes/public"
	//path := "../recipes/recipe3/recipe3"
	//path := "../recipes/all"
	//path := "../recipes/home"

	//outName := "AllRecipes1.xml"
	outName2 := "data/AllRecipes2.xml"
	//outName := "home.xml"
	//outName2 := "home2.xml"
	var err error

	xmlFile, err := os.Open(outName2)

	if err != nil {
		panic(err)
	}

	defer xmlFile.Close()
	startTime := time.Now()

	fmt.Printf("Reading in recipe database...\n\n")

	data, err = beercnv.NewBeerXML2(xmlFile)

	if err != nil {
		fmt.Printf("no xml object from %s", outName2)
		return data, nil
	}

	fmt.Printf("%d Recipes\n", len(data.Recipes))
	fmt.Printf("%d hops\n", len(data.HopVarieties))
	fmt.Printf("%d Fermentables\n", len(data.Fermentables))
	fmt.Printf("%d Miscs\n", len(data.Miscs))
	fmt.Printf("%d Cultures\n", len(data.Cultures))
	fmt.Printf("%d Styles\n", len(data.Styles))

	endTime := time.Now()
	totalTime := endTime.Sub(startTime).Seconds()
	fmt.Printf("\n\nLoad time for recipes: %.6f\n\n", totalTime)

	/*
		for _, s := range data.Styles {
			fmt.Printf("%s\n", s.Name)
		}
	*/

	return data, nil
}

// HandleList generates list of beer hops, fermentables, styles, and yeast strains
func HandleList(words []string, orMes string) (string, error) {
	message := ""
	if len(words) < 2 {
		return "short", nil
	}
	switch words[1] {
	case "hops":
		for _, hop := range data.HopVarieties {
			line := fmt.Sprintf("%s | alpha=%.2f\n", hop.Name, hop.AlphaAcidUnits)
			message += line
		}
	case "styles", "style":
		for _, style := range data.Styles {
			line := fmt.Sprintf("%s\n", style.Name)
			message += line
		}
	case "fermentables", "ferm":
		for _, ferm := range data.Fermentables {
			line := fmt.Sprintf("%s | type=%s\n", ferm.Name, ferm.Type)
			message += line
		}
	case "yeast", "cultures":
		for _, yeast := range data.Cultures {
			line := fmt.Sprintf("%s | att=%.2f\n", yeast.Name, yeast.Attenuation)
			message += line
		}
	}
	return message, nil
}

func HandleExplaination(words []string, orMes string) (string, error) {
	message := ""
	if len(words) < 3 {
		return "", nil
	}
	return message, nil
}
