package googlephotos

type Category string
type Categories []Category

func (Ω Categories) contains(needle Category) bool {
	for _, c := range Ω {
		if c == needle {
			return true
		}
	}
	return false
}

func (Ω Categories) addToSet(c *Category) Categories {
	if c == nil {
		return Ω
	}
	if Ω.contains(*c) {
		return Ω
	}
	return append(Ω, *c)
}

const (
	CategoryAnimals   = Category("ANIMALS")
	CategoryLandmarks = Category("LANDMARKS")
	// "PETS",
	// "UTILITY",
	// "BIRTHDAYS",
	CategoryLandscapes = Category("LANDSCAPES")
	// "RECEIPTS",
	CategoryWeddings   = Category("WEDDINGS")
	CategoryCityScapes = Category("CITYSCAPES")
	CategoryNight      = Category("NIGHT")
	// "SCREENSHOTS",
	// "WHITEBOARDS",
	// "DOCUMENTS",
	CategoryPeople       = Category("PEOPLE")
	CategorySelfies      = Category("SELFIES")
	CategoryFood         = Category("FOOD")
	CategoryPerformances = Category("PERFORMANCES")
	CategorySport        = Category("SPORT")
)

var knownCategories = []Category{
	CategoryAnimals, CategoryLandmarks, CategoryLandscapes, CategoryWeddings, CategoryCityScapes, CategoryNight,
	CategoryWeddings, CategoryPeople, CategorySelfies, CategoryFood, CategoryPerformances, CategorySport,
}
