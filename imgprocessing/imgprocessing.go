package imgprocessing

import (
	"fmt"
	"math"
	"sort"
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
	maxHeight := getMaxNum(len((*matrix1)[0]), len((*matrix2)[0]))

	newMatrix := [][][3]uint8{}

	for x := 0; x < maxWidth; x++ {
		newX := [][3]uint8{}

		for y := 0; y < maxHeight; y++ {
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
	height := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			(*matrix)[x][y][0] = onPixel((*matrix)[x][y][0])
			(*matrix)[x][y][1] = onPixel((*matrix)[x][y][1])
			(*matrix)[x][y][2] = onPixel((*matrix)[x][y][2])
		}
	}
}

func ConvertMatrixToGrayscale(matrix *[][][3]uint8) {
	width := len(*matrix)
	height := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
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
	height := len((*matrix)[0])

	pixelTotalSum := 0

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			grayValueFloat32 := (float32((*matrix)[x][y][0]) + float32((*matrix)[x][y][1]) + float32((*matrix)[x][y][2])) / 3

			grayValue := uint8(grayValueFloat32)

			pixelTotalSum += int(grayValue)

			(*matrix)[x][y][0] = grayValue
			(*matrix)[x][y][1] = grayValue
			(*matrix)[x][y][2] = grayValue
		}
	}

	threshold := uint8(pixelTotalSum / (width * height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
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
	height := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
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
	height := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
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

	matrixSize := float64(width * height)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
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

func GetColorPixelValues(matrix *[][][3]uint8) [3][]uint8 {
	width := len(*matrix)
	height := len((*matrix)[0])

	var values [3][]uint8

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			for z := 0; z < 3; z++ {
				colorIndexValue := (*matrix)[x][y][z]
				values[z] = append(values[z], colorIndexValue)
			}
		}
	}

	return values
}

func ReplaceMatrixBlackForColor(colorIndex int, matrix *[][][3]uint8) {
	width := len(*matrix)
	height := len((*matrix)[0])

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
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

func makeVerticalLine(width int, RGB [3]uint8, amountOfPixels int) *[][][3]uint8 {
	var verticalLine [][][3]uint8

	for x := 0; x < amountOfPixels; x++ {
		newX := [][3]uint8{}

		for y := 0; y < width; y++ {
			newX = append(newX, RGB)
		}

		verticalLine = append(verticalLine, newX)
	}

	return &verticalLine
}

func makeHorizontalLine(width int, RGB [3]uint8, amountOfPixels int) *[][][3]uint8 {
	var horizontalLine [][][3]uint8

	for x := 0; x < width; x++ {
		newX := [][3]uint8{}

		for y := 0; y < amountOfPixels; y++ {
			newX = append(newX, RGB)
		}

		horizontalLine = append(horizontalLine, newX)
	}

	return &horizontalLine
}

func CombineMatrixesHorizontally(matrixes []*[][][3]uint8, separatorWidth int) *[][][3]uint8 {
	var newMatrix [][][3]uint8

	for i := 0; i < len(matrixes); i++ {
		verticalLine := makeVerticalLine(len((*matrixes[i])[0]), [3]uint8{255, 255, 255}, separatorWidth)

		newMatrix = append(newMatrix, *verticalLine...)

		newMatrix = append(newMatrix, *matrixes[i]...)

		newMatrix = append(newMatrix, *verticalLine...)
	}

	return &newMatrix
}

func CombineMatrixesVertically(matrixes []*[][][3]uint8, separatorWidth int) *[][][3]uint8 {
	var newMatrix [][][3]uint8

	width := len(*matrixes[0])

	firstHorizontalLine := makeHorizontalLine(width, [3]uint8{255, 255, 255}, 0)

	newMatrix = append(newMatrix, *firstHorizontalLine...)

	for i := 0; i < len(matrixes); i++ {
		horizontalLine := makeHorizontalLine(len(*matrixes[i]), [3]uint8{255, 255, 255}, separatorWidth)

		for x := 0; x < width; x++ {
			newMatrix[x] = append(newMatrix[x], (*horizontalLine)[x]...)

			newMatrix[x] = append(newMatrix[x], (*matrixes[i])[x]...)

			newMatrix[x] = append(newMatrix[x], (*horizontalLine)[x]...)
		}
	}

	return &newMatrix
}

func ResizeNearestNeighbor(matrix *[][][3]uint8, newWidth uint64, newHeight uint64) *[][][3]uint8 {
	width := len(*matrix)
	height := len((*matrix)[0])

	scaleX := 1 / (float64(newWidth) / float64(width))
	scaleY := 1 / (float64(newHeight) / float64(height))

	var newMatrix [][][3]uint8

	for x := 0; x < int(newWidth); x++ {
		newX := [][3]uint8{}

		oldX := int(math.Min(float64(x)*scaleX, float64(width-1)))

		for y := 0; y < int(newHeight); y++ {
			oldY := int(math.Min(float64(y)*scaleY, float64(height-1)))

			newY := (*matrix)[oldX][oldY]

			newX = append(newX, newY)
		}

		newMatrix = append(newMatrix, newX)
	}

	return &newMatrix
}

func CopyMatrix(matrix *[][][3]uint8) *[][][3]uint8 {
	width := len(*matrix)
	height := len((*matrix)[0])

	var newMatrix [][][3]uint8

	for x := 0; x < width; x++ {
		newX := [][3]uint8{}

		for y := 0; y < height; y++ {
			newY := (*matrix)[x][y]

			newX = append(newX, newY)
		}

		newMatrix = append(newMatrix, newX)
	}

	return &newMatrix
}

//função generica para qualquer filtro

func ApplyFilter(matrix *[][][3]uint8, mask [][]float64, operation func(pixels []float64) uint8) *[][][3]uint8 {
	width := len(*matrix)
	height := len((*matrix)[0])

	maskSize := len(mask)

	maskCenter := maskSize / 2

	var newMatrix [][][3]uint8

	for x := 0; x < width; x++ {
		newX := [][3]uint8{}

		for y := 0; y < height; y++ {

			if x == 0 || y == 0 || x == width-1 || y == height-1 {
				newX = append(newX, (*matrix)[x][y])
				continue
			}

			var redPixels []float64
			var greenPixels []float64
			var bluePixels []float64

			for maskX := 0; maskX < maskSize; maskX++ {
				neighborX := x + maskX - maskCenter

				for maskY := 0; maskY < maskSize; maskY++ {
					neighborY := y + maskY - maskCenter

					neighbor := [3]uint8{0, 0, 0}

					if neighborX >= 0 && neighborX < width && neighborY >= 0 && neighborY < height {
						neighbor = (*matrix)[neighborX][neighborY]
					}

					redPixel := float64(neighbor[0]) * mask[maskX][maskY]
					greenPixel := float64(neighbor[1]) * mask[maskX][maskY]
					bluePixel := float64(neighbor[2]) * mask[maskX][maskY]

					redPixels = append(redPixels, redPixel)
					greenPixels = append(greenPixels, greenPixel)
					bluePixels = append(bluePixels, bluePixel)
				}
			}

			redResult := operation(redPixels)
			greenResult := operation(greenPixels)
			blueResult := operation(bluePixels)

			result := [3]uint8{redResult, greenResult, blueResult}

			newX = append(newX, result)
		}

		newMatrix = append(newMatrix, newX)
	}

	return &newMatrix
}

//funções para criar mascaras de filtros

func MakeMaskOfOnes(size int) [][]float64 {
	mask := [][]float64{}

	for i := 0; i < size; i++ {
		maskRow := []float64{}
		for j := 0; j < size; j++ {
			maskRow = append(maskRow, 1)
		}
		mask = append(mask, maskRow)
	}

	return mask
}

func calcGaussian2d(x float64, y float64, sigma float64) float64 {
	y2 := math.Pow(y, 2)
	x2 := math.Pow(x, 2)
	sigma2 := math.Pow(sigma, 2)

	eExp := -(x2 + y2) / (2 * sigma2)

	return math.Pow(math.E, eExp) / (2 * math.Pi * sigma2)
}

func Matrix2String(t *[][]float64) string {
	s := ""

	for i := range *t {
		for _, n := range (*t)[i] {
			s += fmt.Sprintf("%v ", math.Ceil(n*255))
		}
		s += "\n"
	}

	return s
}

func MakeGaussMask(size int, sigma float64) [][]float64 {
	halfSize := size / 2

	sum := 0.0

	mask := [][]float64{}

	for x := -halfSize; x <= halfSize; x++ {
		maskRow := []float64{}

		for y := -halfSize; y <= halfSize; y++ {
			gauss := calcGaussian2d(float64(x), float64(y), sigma)

			sum += gauss

			maskRow = append(maskRow, gauss)
		}

		mask = append(mask, maskRow)
	}

	fmt.Println(Matrix2String(&mask))

	for x := range mask {
		for y := range mask[x] {
			mask[x][y] /= sum
		}
	}

	return mask
}

//funções para calcular valor do pixel alvo nos filtros

func PixelsMax(pixels []float64) uint8 {
	var maxPixel float64 = math.Inf(-1)

	for i := 0; i < len(pixels); i++ {
		if maxPixel < pixels[i] {
			maxPixel = pixels[i]
		}
	}

	return uint8(maxPixel)
}

func PixelsMin(pixels []float64) uint8 {
	minPixel := math.Inf(1)

	for i := 0; i < len(pixels); i++ {
		if minPixel > pixels[i] {
			minPixel = pixels[i]
		}
	}

	return uint8(minPixel)
}

func PixelsAvg(pixels []float64) uint8 {
	var sum float64 = 0

	arrSize := len(pixels)

	for i := 0; i < arrSize; i++ {
		sum += pixels[i]
	}

	avg := uint8(sum / float64(arrSize))

	return avg
}

func PixelsMean(pixels []float64) uint8 {
	arrCenter := len(pixels) / 2

	sort.Float64s(pixels)

	return uint8(pixels[arrCenter])
}

func PixelsSum(pixels []float64) uint8 {
	sum := 0.0
	for _, pixel := range pixels {
		sum += pixel
	}
	return uint8(sum)
}

func GetPixelByIndexInSortedArr(pixels []float64, index int) uint8 {
	sort.Float64s(pixels)

	return uint8(pixels[index])
}

func GetPixelByIndexInSortedArrCurry(index int) func(pixels []float64) uint8 {
	return func(pixels []float64) uint8 {
		return GetPixelByIndexInSortedArr(pixels, index)
	}
}

func GetPixelBoundedByNeighborsRange(pixels []float64) uint8 {
	arrCenter := len(pixels) / 2

	centerPixel := uint8(pixels[arrCenter])

	nonCenterPixels := []float64{}
	for i, pixel := range pixels {
		if i == arrCenter {
			continue
		}

		nonCenterPixels = append(nonCenterPixels, pixel)
	}

	max := PixelsMax(nonCenterPixels)

	min := PixelsMin(nonCenterPixels)

	var result uint8

	if max < centerPixel {
		result = max
	} else if min > centerPixel {
		result = min
	} else {
		result = centerPixel
	}

	return result
}
