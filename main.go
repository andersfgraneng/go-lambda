package main

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"golang.org/x/image/draw"
)

type Event struct {
	GetObjectContext `json:"getObjectContext"`
	UserRequest      `json:"userRequest"`
}

type GetObjectContext struct {
	InputS3Url  string `json:"inputS3Url"`
	OutputRoute string `json:"outputRoute"`
	OutputToken string `json:"outputToken"`
}

type UserRequest struct {
	Url string `json:"url"`
}

type AWSResponse struct {
	StatusCode int `json:"status_code"`
}

func HandleRequest(ctx context.Context, event Event) (AWSResponse, error) {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	client := s3.NewFromConfig(cfg)

	objContext := event.GetObjectContext
	s3Url := objContext.InputS3Url

	res, _ := http.Get(s3Url)

	img, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	watermarkedImage := watermarkImage(img)

	client.WriteGetObjectResponse(ctx, &s3.WriteGetObjectResponseInput{
		RequestRoute: &objContext.OutputRoute,
		RequestToken: &objContext.OutputToken,
		Body:         bytes.NewReader(watermarkedImage),
	})

	return AWSResponse{200}, nil
}

func watermarkImage(img []byte) []byte {
	originalImage, _ := jpeg.Decode(bytes.NewReader(img))

	wm, _ := os.Open("watermark.png")
	watermark, _ := png.Decode(wm)
	defer wm.Close()

	destinationWatermark := image.NewRGBA(image.Rect(0, 0, originalImage.Bounds().Max.X/2, originalImage.Bounds().Max.Y/2))
	draw.NearestNeighbor.Scale(destinationWatermark, destinationWatermark.Rect, watermark, watermark.Bounds(), draw.Over, nil)

	offset := image.Pt(200, 200)
	bounds := originalImage.Bounds()
	newImage := image.NewRGBA(bounds)
	draw.Draw(newImage, bounds, originalImage, image.Point{}, draw.Src)
	draw.Draw(newImage, destinationWatermark.Bounds().Add(offset), destinationWatermark, image.Point{}, draw.Over)

	buff := new(bytes.Buffer)
	jpeg.Encode(buff, newImage, &jpeg.Options{Quality: jpeg.DefaultQuality})
	return buff.Bytes()
}

func main() {
	lambda.Start(HandleRequest)
}
