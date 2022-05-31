package server

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"image/png"
	"net/http"
	"strconv"

	"img-ops/imgconversion"
	"img-ops/imgprocessing"
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
	img := imgconversion.CreateImgFromMatrix(matrix)

	buf := new(bytes.Buffer)

	err := png.Encode(buf, img)
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

	matrix, err := imgconversion.LoadImg(&multipartFile)
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 3000000)

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

	router.Run("localhost:9090")
}