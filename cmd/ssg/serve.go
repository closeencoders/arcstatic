package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func executeServe(path string, port int) error {

	if port > 60_000 || port <= 0 {
		return fmt.Errorf("invalid port to serve files %d", port)
	}

	fileServer := http.FileServer(http.Dir(path))
	http.Handle("/", fileServer)

	slog.Info("Serving Http", "Location", path, "Port", port)
	portToken := fmt.Sprintf(":%d", port)
	if err := http.ListenAndServe(portToken, nil); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	slog.Debug("serve complete")
	return nil
}
