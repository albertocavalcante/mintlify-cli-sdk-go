// Example: validate a Mintlify docs project.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	fmt.Printf("Using runner: %s (%s)\n", client.Runner().Name, client.Runner().Cmd)

	result, err := client.Validate(context.Background(), mintlify.ValidateOptions{Strict: true})
	if err != nil {
		log.Fatal(err)
	}

	if result.OK {
		fmt.Println("All checks passed!")
		return
	}

	fmt.Printf("Found %d issue(s):\n", len(result.Errors))
	for _, e := range result.Errors {
		if e.File != "" {
			fmt.Printf("  %s:%d:%d — %s\n", e.File, e.Line, e.Column, e.Message)
		} else {
			fmt.Printf("  %s\n", e.Message)
		}
	}
	os.Exit(1)
}
