package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	resizeImage(&requestImage)
	fmt.Fprintf(writer, fmt.Sprintf("link is %q", requestImage.Link))

	// file, err := os.ReadFile("static/thumbnail/asdasd.png")
	// if err != nil {
	// 	logger.Error(
	// 		"file doesnt exist",
	// 		"syserr", err.Error(),
	// 	)
	// 	return
	// }

	// writer.WriteHeader(http.StatusOK)
	// writer.Header().Set("Content-Type", "application/octet-stream")
	// writer.Write(file)

	return
}

func resizeImage(requestImage *RequestImage) {

	resp, err := http.Get(requestImage.Link)
	if err != nil {
		logger.Error(
			"Error in getting image",
			"method", resizeImage,
		)
		return
	}

	defer resp.Body.Close()
	fmt.Println(resp)

}