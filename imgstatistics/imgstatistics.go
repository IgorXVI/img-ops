package imgstatistics

import (
	"bytes"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"img-ops/imgconversion"
	"img-ops/imgprocessing"
)

func makePixelHist(colorName string, pixelValues []uint8) (*bytes.Buffer, error) {
	var values plotter.Values
	for i := 0; i < len(pixelValues); i++ {
		values = append(values, float64(pixelValues[i]))
	}

	p := plot.New()
	p.Title.Text = colorName

	hist, err := plotter.NewHist(values, 256)
	if err != nil {
		return nil, err
	}

	p.Add(hist)

	writer, err := p.WriterTo(6*vg.Centimeter, 6*vg.Centimeter, "png")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	writer.WriteTo(buf)

	return buf, nil
}

func GetMatrixHistRGB(matrix *[][][3]uint8) (*[][][3]uint8, error) {
	var matrixes []*[][][3]uint8

	colorNames := [3]string{"red", "green", "blue"}

	colorPixelValues := imgprocessing.GetColorPixelValues(matrix)

	for i := 0; i < 3; i++ {
		histBuf, err := makePixelHist(colorNames[i], colorPixelValues[i])
		if err != nil {
			return nil, err
		}

		histMatrix, err := imgconversion.LoadImg(histBuf)
		if err != nil {
			return nil, err
		}

		imgprocessing.ReplaceMatrixBlackForColor(i, histMatrix)

		matrixes = append(matrixes, histMatrix)
	}

	newMatrix := imgprocessing.CombineMatrixesHorizontally(matrixes)

	return newMatrix, nil
}

func CompareHistograms(matrix1 *[][][3]uint8, matrix2 *[][][3]uint8) (*[][][3]uint8, error) {
	histMatrix1, err := GetMatrixHistRGB(matrix1)
	if err != nil {
		return nil, err
	}

	histMatrix2, err := GetMatrixHistRGB(matrix2)
	if err != nil {
		return nil, err
	}

	matrix1Resized := imgprocessing.ResizeNearestNeighbor(matrix1, 500, 500)

	histMatrix1Resized := imgprocessing.ResizeNearestNeighbor(histMatrix1, 1500, 500)

	result1 := imgprocessing.CombineMatrixesHorizontally([]*[][][3]uint8{matrix1Resized, histMatrix1Resized})

	matrix2Resized := imgprocessing.ResizeNearestNeighbor(matrix2, 500, 500)

	histMatrix2Resized := imgprocessing.ResizeNearestNeighbor(histMatrix2, 1500, 500)

	result2 := imgprocessing.CombineMatrixesHorizontally([]*[][][3]uint8{matrix2Resized, histMatrix2Resized})

	combinedResult := imgprocessing.CombineMatrixesVertically([]*[][][3]uint8{result1, result2})

	return combinedResult, nil
}
