# brewbrat
Slackbot that answers beer brewing and beer recipe related questions

 Note: commands are not case sensitive
 
  **calc**      OgToBrix|BrixToOg|RefactoToFg|ABV

      OgToBrix Converts Original Gravity to Brix
      BrixToOg Converts Brix to Original Gravity
      ABV <Og> <Fg> Returns Alcohol by Volume in perv\centage
      RefactoToFg or "fg" for short - takes original Brix and Final Brix measured from refractometer and returns the actual brix
          RefactoToFg <brix Og> <brix Fg>


  **Example:**

---
      "@umbot calc ogtobrix 1.089"   (returns "Brix is 21.35")
      "@umbot calc BrixToOg 21.35"   (Returns "Original Gravity is 1.0890")
      "@umbot calc RefractoToFg 19.3 11.3"  (Returns "Final gravity is 1.0228" with ABV of 7.50%)
---

