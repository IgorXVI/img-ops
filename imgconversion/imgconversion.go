package imgconversion

import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

//parte que lida com conversão de dados

func LoadImg(data io.Reader) (*[][][3]uint8, error) {
	image, _, err := image.Decode(data)
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

func getMatrixHistForOneColor(color string, matrix *[][][3]uint8) (*[][][3]uint8, error) {
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
	p.Title.Text = color

	hist, err := plotter.NewHist(values, 256)
	if err != nil {
		return nil, err
	}

	p.Add(hist)

	writer, err := p.WriterTo(7.5*vg.Centimeter, 7.5*vg.Centimeter, "png")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	writer.WriteTo(buf)

	newMatrix, err := LoadImg(buf)
	if err != nil {
		return nil, err
	}

	nWidth := len(*newMatrix)
	nHeigth := len((*newMatrix)[0])

	for x := 0; x < nWidth; x++ {
		for y := 0; y < nHeigth; y++ {
			for z := 0; z < 3; z++ {
				if (*newMatrix)[x][y][0] != 255 || (*newMatrix)[x][y][1] != 255 || (*newMatrix)[x][y][2] != 255 {
					(*newMatrix)[x][y][0] = 0
					(*newMatrix)[x][y][1] = 0
					(*newMatrix)[x][y][2] = 0

					(*newMatrix)[x][y][colorIndex] = 255
				}
			}
		}
	}

	return newMatrix, nil
}

func GetMatrixHistRGB(matrix *[][][3]uint8) (*bytes.Buffer, error) {
	redHistMatrix, err := getMatrixHistForOneColor("red", matrix)
	if err != nil {
		return nil, err
	}

	greenHistMatrix, err := getMatrixHistForOneColor("green", matrix)
	if err != nil {
		return nil, err
	}

	blueHistMatrix, err := getMatrixHistForOneColor("blue", matrix)
	if err != nil {
		return nil, err
	}

	var newMatrix [][][3]uint8

	newMatrix = append(newMatrix, *redHistMatrix...)

	newMatrix = append(newMatrix, *greenHistMatrix...)

	newMatrix = append(newMatrix, *blueHistMatrix...)

	newBuf, err := CreatePNGBufferFromMatrix(&newMatrix)
	if err != nil {
		return nil, err
	}

	return newBuf, nil
}
