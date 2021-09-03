package land

type tileInput struct {
	zoom   *int
	x      *int
	y      *int
	format *string
}

type env struct {
	xMin float64
	xMax float64
	yMin float64
	yMax float64
}
