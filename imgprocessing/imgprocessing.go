package imgprocessing

import "math"

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

func AddPixels(pixel1 uint8, pixel2 uint8) uint8 {
	newPixel := pixel1 + pixel2

	maxPixel := getMaxNum(pixel1, pixel2)

	if newPixel > 255 || newPixel < maxPixel {
		newPixel = 255
	}

	return newPixel
}

func SubtractPixels(pixel1 uint8, pixel2 uint8) uint8 {
	var newPixel uint8

	if pixel1 > pixel2 {
		newPixel = pixel1 - pixel2
	} else {
		newPixel = pixel2 - pixel1
	}

	return newPixel
}

func blendPixels(factor float32, pixel1 uint8, pixel2 uint8) uint8 {
	var newPixel float32 = factor*float32(pixel1) + (1-factor)*float32(pixel2)

	return uint8(newPixel)
}

func BlendPixelsCurry(factor float32) func(pixel1 uint8, pixel2 uint8) uint8 {
	return func(pixel1, pixel2 uint8) uint8 {
		return blendPixels(factor, pixel1, pixel2)
	}
}

func AvgPixels(pixel1 uint8, pixel2 uint8) uint8 {
	var newPixel float32 = (float32(pixel1) + float32(pixel2)) / 2

	return uint8(newPixel)
}

func ANDPixels(pixel1 uint8, pixel2 uint8) uint8 {
	return pixel1 & pixel2
}

func ORPixels(pixel1 uint8, pixel2 uint8) uint8 {
	return pixel1 | pixel2
}

func XORPixels(pixel1 uint8, pixel2 uint8) uint8 {
	return pixel1 ^ pixel2
}

func OperateOnTwoMatrixes(
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

			for z := 0; z < 3; z++ {
				var pixel1 uint8 = 0
				var pixel2 uint8 = 0

				if x < len(*matrix1) && y < len((*matrix1)[0]) {
					pixel1 = (*matrix1)[x][y][z]
				}

				if x < len(*matrix2) && y < len((*matrix2)[0]) {
					pixel2 = (*matrix2)[x][y][z]
				}

				newY[z] = onPixel(pixel1, pixel2)
			}

			newX = append(newX, newY)
		}

		newMatrix = append(newMatrix, newX)
	}

	return newMatrix
}

func multiplyPixel(factor float32, pixel uint8) uint8 {
	newPixel := factor * float32(pixel)

	if newPixel > 255 {
		newPixel = 255
	}

	return uint8(newPixel)
}

func MultiplyPixelCurry(factor float32) func(pixel uint8) uint8 {
	return func(pixel uint8) uint8 {
		return multiplyPixel(factor, pixel)
	}
}

func OperateOnMatrix(
	matrix *[][][3]uint8,
	onPixel func(pixel uint8) uint8,
) {
	width := len(*matrix)
	heigth := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			(*matrix)[x][y][0] = onPixel((*matrix)[x][y][0])
			(*matrix)[x][y][1] = onPixel((*matrix)[x][y][1])
			(*matrix)[x][y][2] = onPixel((*matrix)[x][y][2])
		}
	}
}

func ConvertMatrixToGrayscale(matrix *[][][3]uint8) {
	width := len(*matrix)
	heigth := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			grayValueFloat32 := (float32((*matrix)[x][y][0]) + float32((*matrix)[x][y][1]) + float32((*matrix)[x][y][2])) / 3

			grayValue := uint8(grayValueFloat32)

			(*matrix)[x][y][0] = grayValue
			(*matrix)[x][y][1] = grayValue
			(*matrix)[x][y][2] = grayValue
		}
	}
}

func ConvertMatrixToBinary(matrix *[][][3]uint8) {
	width := len(*matrix)
	heigth := len((*matrix)[0])

	pixelTotalSum := 0

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			grayValueFloat32 := (float32((*matrix)[x][y][0]) + float32((*matrix)[x][y][1]) + float32((*matrix)[x][y][2])) / 3

			grayValue := uint8(grayValueFloat32)

			pixelTotalSum += int(grayValue)

			(*matrix)[x][y][0] = grayValue
			(*matrix)[x][y][1] = grayValue
			(*matrix)[x][y][2] = grayValue
		}
	}

	threshold := uint8(pixelTotalSum / (width * heigth))

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			pixelValue := (*matrix)[x][y][0]

			var newPixelValue uint8 = 0

			if pixelValue >= threshold {
				newPixelValue = 255
			}

			(*matrix)[x][y][0] = newPixelValue
			(*matrix)[x][y][1] = newPixelValue
			(*matrix)[x][y][2] = newPixelValue
		}
	}
}

func NOTMatrix(matrix *[][][3]uint8) {
	width := len(*matrix)
	heigth := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			(*matrix)[x][y][0] = 255 - (*matrix)[x][y][0]
			(*matrix)[x][y][1] = 255 - (*matrix)[x][y][1]
			(*matrix)[x][y][2] = 255 - (*matrix)[x][y][2]
		}
	}
}

func EqualizeMatrixHistogram(matrix *[][][3]uint8) {
	var hist [3][256]int

	for i := 0; i < 3; i++ {
		for j := 0; j < 256; j++ {
			hist[i][j] = 0
		}
	}

	width := len(*matrix)
	heigth := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			for z := 0; z < 3; z++ {
				colorValue := (*matrix)[x][y][z]
				hist[z][colorValue]++
			}
		}
	}

	var histCFD [3][256]int

	for i := 0; i < 3; i++ {
		histCFD[i][0] = hist[i][0]
	}

	for i := 0; i < 3; i++ {
		for j := 1; j < 256; j++ {
			histCFD[i][j] = histCFD[i][j-1] + hist[i][j]
		}
	}

	matrixSize := float64(width * heigth)

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			for z := 0; z < 3; z++ {
				colorValue := (*matrix)[x][y][z]

				histCFDValue := float64(histCFD[z][colorValue])

				histCFDMin := float64(histCFD[z][0])

				result := math.Floor((histCFDValue - histCFDMin) / (matrixSize - histCFDMin) * 255)

				(*matrix)[x][y][z] = uint8(result)
			}
		}
	}
}

func GetColorPixelValues(colorIndex int, matrix *[][][3]uint8) []uint8 {
	width := len(*matrix)
	heigth := len((*matrix)[0])

	var values []uint8

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			colorIndexValue := (*matrix)[x][y][colorIndex]
			values = append(values, colorIndexValue)
		}
	}

	return values
}

func ReplaceMatrixBlackForColor(colorIndex int, matrix *[][][3]uint8) {
	width := len(*matrix)
	heigth := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < heigth; y++ {
			for z := 0; z < 3; z++ {
				if (*matrix)[x][y][0] != 255 || (*matrix)[x][y][1] != 255 || (*matrix)[x][y][2] != 255 {
					(*matrix)[x][y][0] = 0
					(*matrix)[x][y][1] = 0
					(*matrix)[x][y][2] = 0

					(*matrix)[x][y][colorIndex] = 255
				}
			}
		}
	}
}

func CombineMatrixesHorizontally(matrixes []*[][][3]uint8) *[][][3]uint8 {
	var newMatrix [][][3]uint8

	for i := 0; i < len(matrixes); i++ {
		newMatrix = append(newMatrix, *matrixes[i]...)
	}

	return &newMatrix
}
