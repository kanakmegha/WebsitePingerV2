package main

import (
	"fmt"
	"log"

	"github.com/kanakmegha/WebsitePingerV2/internal/auth"
)

func main() {
	password := "123456"

	fmt.Println("=== Testing Argon2id Password Hashing ===")
	fmt.Printf("Input Password: %s (length: %d)\n\n", password, len(password))

	// 1. Hash the password
	hashedOutput, err := auth.HashPassword(password)
	if err != nil {
		fmt.Printf("Error Hashing Password: %v\n", err)
		return
	}

	fmt.Printf("Generated Hash:\n%s\n\n", hashedOutput)

	// 2. Verify the password
	isValid, err := auth.VerifyPassword(hashedOutput, password)
	if err != nil {
		log.Fatalf("Error verifying password: %v", err)
	}

	fmt.Printf("Verification Result: %v\n", isValid)
}
