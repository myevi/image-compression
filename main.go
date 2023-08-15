package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nfnt/resize"
)

type RequestImage struct {
	Link string `json:"link"`
}

var (
	port string
	logger *slog.Logger
)

func init() {
	initLogger()
}

func main() {
	port = os.Getenv("PORT")
	if port == "" {
		logger.Error("port is undefined")
		port = ":8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/thmbnl", thumbnailHandler)

	server := &http.Server{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Handler:      mux,
	}

	defer server.Close()

	listener, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error(
			"net.Listen error",
			"err", err.Error(),
		)
	}

	go func() {
		server.Serve(listener)
	}()

	logger.Info(
		"server started",
		"port", port,
	)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	logger.Info("server stoped with signal", <- ch)
}

func initLogger() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger = slog.New(handler)
}

func thumbnailHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		logger.Error(
			"method not allowed",
			"handler", "thumbnailHandler",
		)

		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Request body error", http.StatusBadRequest)
		logger.Error(
			"Request body error",
			"Handler", "thumbnailHandler",
			"Error", err.Error(),
		)

		return
	}

	requestImage := RequestImage{}
	err = json.Unmarshal(body, &requestImage)
	if err != nil {
		http.Error(writer, "Unsuccesful decoding", http.StatusBadRequest)
		logger.Error(
			"Unsuccesful decoding",
			"Handler", "thumbnailHandler",
			"Error", err.Error(),
		)

		return
	}

	originalImageBuffer, format := getImage(&requestImage)
	resizedImageBuffer := resizeImage(originalImageBuffer, format)
	buffer := new(bytes.Buffer)
	err = jpeg.Encode(buffer, resizedImageBuffer, nil)
	if err != nil {
		return
	}

    writer.WriteHeader(http.StatusOK)
    writer.Header().Set("Content-Type", "image/jpeg")
    writer.Write(buffer.Bytes())
	logger.Info("image has been send")

	return
}

func resizeImage(originalImageBuffer image.Image, format string) image.Image {
	var resizedImage image.Image
	newWidth := 300
    ratio := float64(originalImageBuffer.Bounds().Dx()) / float64(newWidth)
    newHeight := int(float64(originalImageBuffer.Bounds().Dy()) / ratio)
    resizedImage = resize.Resize(uint(newWidth), uint(newHeight), originalImageBuffer, resize.Lanczos3)
	return resizedImage
}

func getImage(requestImage *RequestImage) (image.Image, string) {

	resp, err := http.Get(requestImage.Link)
	if err != nil {
		logger.Error(
			"Error in getting image",
			"method", "getImage",
		)
		return nil, ""
	}

	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(
			"body is empty",
			"method", "getImage",
			"error", err.Error(),
		)
		return nil, ""
	}

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		logger.Error(
			"image decode error",
			"method", "getImage",
			"error", err.Error(),
		)
		return nil, ""
	}

	return img, format
}