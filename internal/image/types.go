package image

type Rgb struct {
	R uint8
	G uint8
	B uint8
}

type Image struct {
	W   int
	H   int
	Pix []Rgb
}
