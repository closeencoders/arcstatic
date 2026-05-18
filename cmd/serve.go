package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
)

func executeServe(location string, port int) {

	if port > 60_000 || port <= 0 {
		slog.Error("Invalid Port To Serve Files From", "port", port)
		return
	}

	fileServer := http.FileServer(http.Dir(location))
	http.Handle("/", fileServer)

	slog.Info("Serving Http", "Location", location, "Port", port)
	portToken := fmt.Sprintf(":%d", port)
	if err := http.ListenAndServe(portToken, nil); err != nil {
		slog.Error("Program failure", "error", err)
		panic(err)
	}
}
