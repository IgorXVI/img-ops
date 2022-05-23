package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
)

func loadImg(filePath string) (*[][][3]uint8, error) {
	imgFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	image, _, err := image.Decode(imgFile)
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

func loadImgs(paths []string) ([]*[][][3]uint8, error) {
	matrixes := []*[][][3]uint8{}

	for i := 0; i < len(paths); i++ {
		path := paths[i]

		matrix, err := loadImg(path)

		if err != nil {
			return nil, err
		}

		matrixes = append(matrixes, matrix)
	}

	return matrixes, nil
}

type AnyInteger interface {
	int | uint8
}

func getMaxNum[T AnyInteger](num1 T, num2 T) T {
	var greatestNum T = 0
	if num1 > num2 {
		greatestNum = num1
	} else {
		greatestNum = num2
	}
	return greatestNum
}

func addPixels(pixel1 uint8, pixel2 uint8) uint8 {
	newPixel := pixel1 + pixel2

	maxPixel := getMaxNum(pixel1, pixel2)

	if newPixel > 255 || newPixel < maxPixel {
		newPixel = 255
	}

	return newPixel
}

func subtractPixels(pixel1 uint8, pixel2 uint8) uint8 {
	var newPixel uint8

	if pixel1 > pixel2 {
		newPixel = pixel1 - pixel2
	} else {
		newPixel = pixel2 - pixel1
	}

	return newPixel
}

func multiplyPixels(pixel1 uint8, pixel2 uint8) uint8 {
	var newPixel float32 = float32(pixel1) * float32(pixel2)

	return uint8(newPixel / 255)
}

func dividePixels(pixel1 uint8, pixel2 uint8) uint8 {
	var newPixel float32

	if pixel1 > pixel2 {
		newPixel = float32(pixel2) / float32(pixel1)
	} else {
		newPixel = float32(pixel1) / float32(pixel2)
	}

	return uint8(newPixel * 255)
}

func avgPixels(pixel1 uint8, pixel2 uint8) uint8 {
	var newPixel float32 = (float32(pixel1) + float32(pixel2)) / 2

	return uint8(newPixel)
}

func blendPixels(blendFactor float32, pixel1 uint8, pixel2 uint8) uint8 {
	var newPixel float32 = blendFactor*float32(pixel1) + (1-blendFactor)*float32(pixel2)

	return uint8(newPixel)
}

func blendPixelsCurry(blendFactor float32) func(pixel1 uint8, pixel2 uint8) uint8 {
	return func(pixel1, pixel2 uint8) uint8 {
		return blendPixels(blendFactor, pixel1, pixel2)
	}
}

func andPixels(pixel1 uint8, pixel2 uint8) uint8 {
	return pixel1 & pixel2
}

func orPixels(pixel1 uint8, pixel2 uint8) uint8 {
	return pixel1 | pixel2
}

func xorPixels(pixel1 uint8, pixel2 uint8) uint8 {
	return pixel1 ^ pixel2
}

func notPixels(matrix *[][][3]uint8) {
	var maxRed uint8 = 0
	var maxGreen uint8 = 0
	var maxBlue uint8 = 0

	width := len(*matrix)
	heigth := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			red := (*matrix)[x][y][0]
			green := (*matrix)[x][y][1]
			blue := (*matrix)[x][y][2]

			if red > maxRed {
				maxRed = red
			}

			if green > maxGreen {
				maxGreen = green
			}

			if blue > maxBlue {
				maxBlue = blue
			}
		}
	}

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			(*matrix)[x][y][0] = maxRed - (*matrix)[x][y][0]
			(*matrix)[x][y][1] = maxGreen - (*matrix)[x][y][1]
			(*matrix)[x][y][2] = maxBlue - (*matrix)[x][y][2]
		}
	}
}

func operateOnTwoMatrixes(
	matrix1 *[][][3]uint8,
	matrix2 *[][][3]uint8,
	onPixel func(pixel1 uint8, pixel2 uint8) uint8,
) [][][3]uint8 {
	maxWidth := getMaxNum(len(*matrix1), len(*matrix2))
	maxHeigth := getMaxNum(len((*matrix1)[0]), len((*matrix2)[0]))

	newMatrix := [][][3]uint8{}

	for x := 0; x < maxWidth; x++ {
		newX := [][3]uint8{}

		for y := 0; y < maxHeigth; y++ {
			newY := [3]uint8{}

			for j := 0; j < 3; j++ {
				var pixel1 uint8 = 0
				var pixel2 uint8 = 0

				if x < len(*matrix1) && y < len((*matrix1)[0]) {
					pixel1 = (*matrix1)[x][y][j]
				}

				if x < len(*matrix2) && y < len((*matrix2)[0]) {
					pixel2 = (*matrix2)[x][y][j]
				}

				newY[j] = onPixel(pixel1, pixel2)
			}

			newX = append(newX, newY)
		}

		newMatrix = append(newMatrix, newX)
	}

	return newMatrix
}

func createImgFromMatrix(matrix *[][][3]uint8) *image.NRGBA {
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

func showImg(img *image.NRGBA) {
	a := app.New()
	w := a.NewWindow("result")

	canvasImg := canvas.NewImageFromImage(img)
	w.SetContent(canvasImg)
	w.Resize(fyne.NewSize(640, 480))

	w.ShowAndRun()
}

func main() {
	fmt.Println("Running...")

	if len(os.Args) < 2 {
		fmt.Println(
			"inform the args in the following order separated by 1 space:\n" +
				"<path of the image folder between single quotes (obligatory)>\n" +
				"<operation: ADD, SUB, MULT, DIV, AVG, BLEND, AND, OR, XOR, NOT (obligatory)>\n" +
				"<name of the first image between single quotes (obligatory)>\n" +
				"<name of the second image between single quotes (optional only when operation is NOT)>\n" +
				"<blend factor (obligatory only when opeartion is BLEND)>\n" +
				"example: ./img-ops.exe 'C:\\Users\\inazu\\OneDrive\\Documentos\\Faculdade\\processamento_imagens\\Matlab\\' ADD 'c-preto.png' 'q-preto.png'",
		)
		return
	}

	if len(os.Args) < 3 {
		fmt.Println("Not enough args!")
		return
	}

	validOperations := map[string]bool{
		"ADD":   true,
		"SUB":   true,
		"MULT":  true,
		"DIV":   true,
		"AVG":   true,
		"BLEND": true,
		"AND":   true,
		"OR":    true,
		"XOR":   true,
		"NOT":   true,
	}

	functionOperationsMap := map[string]func(pixel1 uint8, pixel2 uint8) uint8{
		"ADD":  addPixels,
		"SUB":  subtractPixels,
		"MULT": multiplyPixels,
		"DIV":  dividePixels,
		"AVG":  avgPixels,
		"AND":  andPixels,
		"OR":   orPixels,
		"XOR":  xorPixels,
	}

	imgFolder := os.Args[1]

	operation := os.Args[2]

	var newMatrix [][][3]uint8

	if !validOperations[operation] {
		fmt.Println("Invalid operation!")
		return
	} else if operation != "NOT" {
		if len(os.Args) < 5 {
			fmt.Println("Not enough args!")
			return
		}

		imgPath1 := imgFolder + os.Args[3]
		imgPath2 := imgFolder + os.Args[4]

		matrixes, err := loadImgs([]string{
			imgPath1,
			imgPath2,
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		var opFunc func(pixel1 uint8, pixel2 uint8) uint8

		if operation == "BLEND" {
			if len(os.Args) < 6 {
				fmt.Println("Not enough args!")
				return
			}

			blendFactor, err := strconv.ParseFloat(os.Args[5], 32)
			if err != nil {
				fmt.Println(err)
				return
			}

			opFunc = blendPixelsCurry(float32(blendFactor))
		} else {
			opFunc = functionOperationsMap[operation]
		}

		newMatrix = operateOnTwoMatrixes(matrixes[0], matrixes[1], opFunc)
	} else {
		if len(os.Args) < 4 {
			fmt.Println("Not enough args!")
			return
		}

		imgPath := imgFolder + os.Args[3]

		matrix, err := loadImg(imgPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		notPixels(matrix)

		newMatrix = *matrix
	}

	newImg := createImgFromMatrix(&newMatrix)

	fmt.Println("Completed!")

	showImg(newImg)
}
