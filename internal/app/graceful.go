package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunWithGracefulShutdown(server *http.Server, timeout time.Duration) error {
	// Channel to listen for errors from listeners
	servErr := make(chan error, 1)

	go func() {
		log.Printf("Server starting on %s", server.Addr)
		err := server.ListenAndServe()
		servErr <- err
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// Tells the runtime: “Stop delivering OS signals to this channel”.Think of it as unsubscribe, not close.
	defer signal.Stop(shutdown)

	select {
	case err := <-servErr:
		log.Printf("Error starting server: %v", err)
		return err
	case sig := <-shutdown:
		log.Printf("Shutdown signal received: %v", sig)
		log.Println("Starting graceful shutdown...")

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			log.Println("Forcing server to close...")

			// Force close if graceful shutdown fails
			if closeErr := server.Close(); closeErr != nil {
				// Return both errors
				return errors.Join(err, closeErr)
			}
			return err
		}
	}
	return nil
}
