package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"img-ops/imgconversion"
	"img-ops/imgprocessing"
	"img-ops/imgstatistics"
)

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

func sendMatrixAsImg(context *gin.Context, matrix *[][][3]uint8) {
	buf, err := imgconversion.CreatePNGBufferFromMatrix(matrix)
	if err != nil {
		sendInputError(context, err)
		return
	}

	context.Data(http.StatusOK, "image/png", buf.Bytes())
}

func loadImgFromParams(context *gin.Context, name string) (*[][][3]uint8, error) {
	multipartFile, _, err := context.Request.FormFile(name)
	if err != nil {
		return nil, err
	}

	matrix, err := imgconversion.LoadImg(multipartFile)
	if err != nil {
		return nil, err
	}

	return matrix, nil
}

func getFactorFromParams(context *gin.Context) (float32, error) {
	factorStr := context.Param("factor")

	factor64, err := strconv.ParseFloat(factorStr, 32)
	if err != nil {
		return 0, err
	}

	factor := float32(factor64)

	return factor, nil
}

func getMaskSizeFromParams(context *gin.Context) (int, error) {
	maskSizeInt64, err := strconv.ParseInt(context.Param("maskSize"), 10, 64)
	if err != nil {
		return 0, err
	}
	maskSize := int(maskSizeInt64)

	return maskSize, nil
}

func handleTwoImages(context *gin.Context, pixelOperation func(pixel1 uint8, pixel2 uint8) uint8) {
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

	newMatrix := imgprocessing.OperateOnTwoMatrixes(matrix1, matrix2, pixelOperation)

	sendMatrixAsImg(context, &newMatrix)
}

func handleOneImage(context *gin.Context, pixelOperation func(pixel uint8) uint8) {
	matrix, err := loadImgFromParams(context, "img")
	if err != nil {
		sendInputError(context, err)
		return
	}

	imgprocessing.OperateOnMatrix(matrix, pixelOperation)

	sendMatrixAsImg(context, matrix)
}

func handleMaskOfOnesFilter(context *gin.Context, operation func(pixels []float64) uint8) {
	matrix, err := loadImgFromParams(context, "img")
	if err != nil {
		sendInputError(context, err)
		return
	}

	maskSize, err := getMaskSizeFromParams(context)
	if err != nil {
		sendInputError(context, err)
		return
	}

	mask := imgprocessing.MakeMaskOfOnes(maskSize)

	result := imgprocessing.ApplyFilter(matrix, mask, operation)

	sendMatrixAsImg(context, result)
}

func corsMiddleware(context *gin.Context) {
	context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	context.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	context.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	if context.Request.Method == "OPTIONS" {
		context.AbortWithStatus(204)
		return
	}

	context.Next()
}

func maxBodySizeMiddleware(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 3000000000)

	c.Next()
}

func StartServer() {
	router := gin.Default()

	//lidar com duas imagens

	router.POST("/process-img/add", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleTwoImages(context, imgprocessing.AddPixels)
	})

	router.POST("/process-img/subtract", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleTwoImages(context, imgprocessing.SubtractPixels)
	})

	router.POST("/process-img/blend/:factor", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		factor, err := getFactorFromParams(context)
		if err != nil {
			sendInputError(context, err)
			return
		}

		handleTwoImages(context, imgprocessing.BlendPixelsCurry(factor))
	})

	router.POST("/process-img/avg", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleTwoImages(context, imgprocessing.AvgPixels)
	})

	router.POST("/process-img/and", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleTwoImages(context, imgprocessing.ANDPixels)
	})

	router.POST("/process-img/or", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleTwoImages(context, imgprocessing.ORPixels)
	})

	router.POST("/process-img/xor", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleTwoImages(context, imgprocessing.XORPixels)
	})

	//lidar com uma imagem

	router.POST("/process-img/multiply/:factor", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		factor, err := getFactorFromParams(context)
		if err != nil {
			sendInputError(context, err)
			return
		}

		handleOneImage(context, imgprocessing.MultiplyPixelCurry(factor))
	})

	router.POST("/process-img/divide/:factor", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		factor, err := getFactorFromParams(context)
		if err != nil {
			sendInputError(context, err)
			return
		}

		handleOneImage(context, imgprocessing.MultiplyPixelCurry(1/factor))
	})

	router.POST("/process-img/not", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		matrix, err := loadImgFromParams(context, "img")
		if err != nil {
			sendInputError(context, err)
			return
		}

		imgprocessing.NOTMatrix(matrix)

		sendMatrixAsImg(context, matrix)
	})

	router.POST("/process-img/grayscale", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		matrix, err := loadImgFromParams(context, "img")
		if err != nil {
			sendInputError(context, err)
			return
		}

		imgprocessing.ConvertMatrixToGrayscale(matrix)

		sendMatrixAsImg(context, matrix)
	})

	router.POST("/process-img/binary", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		matrix, err := loadImgFromParams(context, "img")
		if err != nil {
			sendInputError(context, err)
			return
		}

		imgprocessing.ConvertMatrixToBinary(matrix)

		sendMatrixAsImg(context, matrix)
	})

	router.POST("/process-img/equalize-histogram", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		matrix, err := loadImgFromParams(context, "img")
		if err != nil {
			sendInputError(context, err)
			return
		}

		imgprocessing.EqualizeMatrixHistogram(matrix)

		sendMatrixAsImg(context, matrix)
	})

	router.POST("/process-img/histogram", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		matrix, err := loadImgFromParams(context, "img")
		if err != nil {
			sendInputError(context, err)
			return
		}

		histMatrix, err := imgstatistics.GetMatrixHistRGB(matrix)
		if err != nil {
			sendInputError(context, err)
			return
		}

		sendMatrixAsImg(context, histMatrix)
	})

	router.POST("/process-img/compare-histograms", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
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

		result, err := imgstatistics.CompareHistograms(matrix1, matrix2)
		if err != nil {
			sendInputError(context, err)
			return
		}

		sendMatrixAsImg(context, result)
	})

	router.POST("/process-img/equalize-and-compare-histograms", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		matrix, err := loadImgFromParams(context, "img")
		if err != nil {
			sendInputError(context, err)
			return
		}

		matrixOld := imgprocessing.CopyMatrix(matrix)

		imgprocessing.EqualizeMatrixHistogram(matrix)

		result, err := imgstatistics.CompareHistograms(matrixOld, matrix)
		if err != nil {
			sendInputError(context, err)
			return
		}

		sendMatrixAsImg(context, result)
	})

	router.POST("/process-img/filter/max/:maskSize", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleMaskOfOnesFilter(context, imgprocessing.PixelsMax)
	})

	router.POST("/process-img/filter/min/:maskSize", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleMaskOfOnesFilter(context, imgprocessing.PixelsMin)
	})

	router.POST("/process-img/filter/avg/:maskSize", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleMaskOfOnesFilter(context, imgprocessing.PixelsAvg)
	})

	router.POST("/process-img/filter/mean/:maskSize", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleMaskOfOnesFilter(context, imgprocessing.PixelsMean)
	})

	router.POST("/process-img/filter/conservative-smoothing/:maskSize", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		handleMaskOfOnesFilter(context, imgprocessing.GetPixelBoundedByNeighborsRange)
	})

	router.POST("/process-img/filter/order/:maskSize/:index", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		maskSize, err := getMaskSizeFromParams(context)
		if err != nil {
			sendInputError(context, err)
			return
		}

		index64, err := strconv.ParseInt(context.Param("index"), 10, 64)
		if err != nil {
			sendInputError(context, err)
			return
		}
		index := int(index64)

		maxIndex := maskSize*maskSize - 1
		maxIndexStr := strconv.Itoa(int(maxIndex))

		if index < 0 || index > maxIndex {
			sendInputError(context, errors.New("index must be between 0 and "+maxIndexStr))
			return
		}

		matrix, err := loadImgFromParams(context, "img")
		if err != nil {
			sendInputError(context, err)
			return
		}

		mask := imgprocessing.MakeMaskOfOnes(maskSize)

		getPixelByIndexInSortedArr := imgprocessing.GetPixelByIndexInSortedArrCurry(index)

		result := imgprocessing.ApplyFilter(matrix, mask, getPixelByIndexInSortedArr)

		sendMatrixAsImg(context, result)
	})

	router.POST("/process-img/filter/gaussian/:maskSize/:sigma", corsMiddleware, maxBodySizeMiddleware, func(context *gin.Context) {
		matrix, err := loadImgFromParams(context, "img")
		if err != nil {
			sendInputError(context, err)
			return
		}

		maskSize, err := getMaskSizeFromParams(context)
		if err != nil {
			sendInputError(context, err)
			return
		}

		sigma, err := strconv.ParseFloat(context.Param("sigma"), 64)
		if err != nil {
			sendInputError(context, err)
			return
		}

		gaussMask := imgprocessing.MakeGaussMask(maskSize, sigma)

		result := imgprocessing.ApplyFilter(matrix, gaussMask, imgprocessing.PixelsSum)

		sendMatrixAsImg(context, result)
	})

	router.Run("localhost:9090")
}
