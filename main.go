package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
)

//parte que processa as images

func getMaxNum[T int | uint8](num1 T, num2 T) T {
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

//parte que lida com conversão de dados

func loadImg(multipartFile *multipart.File) (*[][][3]uint8, error) {
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

//parte que lida com requisições

type ErrorResponse struct {
	Message string `json:"message"`
}

func sendInputError(context *gin.Context, err error) {
	errMessage := err.Error()

	fmt.Printf("An error ocurred: %v\n", errMessage)

	errorResponse := ErrorResponse{
		Message: errMessage,
	}

	context.JSON(http.StatusBadRequest, errorResponse)
}

func loadImgFromParams(context *gin.Context, name string) (*[][][3]uint8, error) {
	multipartFile, _, err := context.Request.FormFile(name)
	if err != nil {
		return nil, err
	}

	matrix, err := loadImg(&multipartFile)
	if err != nil {
		return nil, err
	}

	return matrix, nil
}

func sendMatrixAsImg(context *gin.Context, matrix *[][][3]uint8) {
	img := createImgFromMatrix(matrix)

	buf := new(bytes.Buffer)

	err := png.Encode(buf, img)
	if err != nil {
		sendInputError(context, err)
		return
	}

	context.Data(http.StatusOK, "image/png", buf.Bytes())
}

func loadAndOperateOnTwoImages(context *gin.Context, pixelOperation func(pixel1 uint8, pixel2 uint8) uint8) {
	matrix1, err := loadImgFromParams(context, "img1")
	if err != nil {
		sendInputError(context, err)
		return
	}

	matrix2, err := loadImgFromParams(context, "img2")
	if err != nil {
		sendInputError(context, err)
		return
	}

	newMatrix := operateOnTwoMatrixes(matrix1, matrix2, pixelOperation)

	sendMatrixAsImg(context, &newMatrix)
}

func main() {
	fmt.Println("Running...")

	router := gin.Default()

	router.POST("/process-img/add", func(context *gin.Context) {
		loadAndOperateOnTwoImages(context, addPixels)
	})

	router.POST("/process-img/subtract", func(context *gin.Context) {
		loadAndOperateOnTwoImages(context, subtractPixels)
	})

	router.POST("/process-img/multiply", func(context *gin.Context) {
		loadAndOperateOnTwoImages(context, multiplyPixels)
	})

	router.POST("/process-img/divide", func(context *gin.Context) {
		loadAndOperateOnTwoImages(context, dividePixels)
	})

	router.POST("/process-img/avg", func(context *gin.Context) {
		loadAndOperateOnTwoImages(context, avgPixels)
	})

	router.POST("/process-img/blend/:blendFactor", func(context *gin.Context) {
		blendFactorStr := context.Param("blendFactor")

		blendFactor64, err := strconv.ParseFloat(blendFactorStr, 32)
		if err != nil {
			sendInputError(context, err)
			return
		}

		blendFactor := float32(blendFactor64)

		loadAndOperateOnTwoImages(context, blendPixelsCurry(blendFactor))
	})

	router.POST("/process-img/and", func(context *gin.Context) {
		loadAndOperateOnTwoImages(context, andPixels)
	})

	router.POST("/process-img/or", func(context *gin.Context) {
		loadAndOperateOnTwoImages(context, orPixels)
	})

	router.POST("/process-img/xor", func(context *gin.Context) {
		loadAndOperateOnTwoImages(context, xorPixels)
	})

	router.POST("/process-img/not", func(context *gin.Context) {
		matrix, err := loadImgFromParams(context, "img")
		if err != nil {
			sendInputError(context, err)
			return
		}

		notPixels(matrix)

		sendMatrixAsImg(context, matrix)
	})

	router.Run("localhost:9090")
}
