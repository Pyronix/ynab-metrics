package units

//Dollars converts milliunits to dollars
func Dollars(milliunits int64) float64 {
	return float64(milliunits / 10) / 100
}
