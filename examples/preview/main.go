// Example: start a Mintlify dev server and wait for it to be ready.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/albertocavalcante/mintlify-cli-sdk-go/mintlify"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	client, err := mintlify.New(dir)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Println("Starting Mintlify dev server...")
	server, err := client.StartDev(ctx, mintlify.DevOptions{Port: 3333})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Waiting for server at %s ...\n", server.URL())
	if err := server.WaitReady(ctx); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	fmt.Printf("Server is ready at %s\n", server.URL())
	fmt.Println("Press Ctrl+C to stop.")

	// Wait for signal or server exit.
	if err := server.Wait(); err != nil {
		fmt.Printf("Server exited: %v\n", err)
	}
}
