package imgconversion

import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"mime/multipart"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

//parte que lida com conversão de dados

func LoadImg(multipartFile *multipart.File) (*[][][3]uint8, error) {
	image, _, err := image.Decode(*multipartFile)
	if err != nil {
		return nil, err
	}

	bounds := image.Bounds()

	RGBMatrix := [][][3]uint8{}

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		newX := [][3]uint8{}

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := image.At(x, y).RGBA()

			newY := [3]uint8{uint8(r / 257), uint8(g / 257), uint8(b / 257)}

			newX = append(newX, newY)
		}

		RGBMatrix = append(RGBMatrix, newX)
	}

	return &RGBMatrix, nil
}

func CreateImgFromMatrix(matrix *[][][3]uint8) *image.NRGBA {
	width := len(*matrix)
	height := len((*matrix)[0])

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewNRGBA(image.Rectangle{upLeft, lowRight})

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{(*matrix)[x][y][0], (*matrix)[x][y][1], (*matrix)[x][y][2], 255})
		}
	}

	return img
}

func CreatePNGBufferFromMatrix(matrix *[][][3]uint8) (*bytes.Buffer, error) {
	img := CreateImgFromMatrix(matrix)

	buf := new(bytes.Buffer)

	err := png.Encode(buf, img)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func GetMatrixColorHistAsPNGBuffer(color string, matrix *[][][3]uint8) (*bytes.Buffer, error) {
	colorIndexMap := map[string]int{
		"red":   0,
		"green": 1,
		"blue":  2,
	}

	colorIndex := colorIndexMap[color]

	width := len(*matrix)
	heigth := len((*matrix)[0])

	var values plotter.Values

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			colorIndexValue := float64((*matrix)[x][y][colorIndex])
			values = append(values, colorIndexValue)
		}
	}

	p := plot.New()
	p.Title.Text = color + " histogram"

	hist, err := plotter.NewHist(values, 10000)
	if err != nil {
		return nil, err
	}

	p.Add(hist)

	writer, err := p.WriterTo(20*vg.Centimeter, 20*vg.Centimeter, "png")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	writer.WriteTo(buf)

	return buf, nil
}
