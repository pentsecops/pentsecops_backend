package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/google/uuid"
	"github.com/pentsecops/backend/pkg/utils"
	"golang.org/x/term"
)

func main() {
	generateFlag := flag.Bool("generate", false, "Generate a password hash")
	idOnlyFlag := flag.Bool("id-only", false, "Only generate a UUID")
	flag.Parse()

	if *idOnlyFlag {
		// Just generate a UUID
		id := uuid.New().String()
		fmt.Println(id)
		return
	}

	if *generateFlag {
		generatePasswordHash()
	} else {
		showUsage()
	}
}

func generatePasswordHash() {
	fmt.Println("=== Password Hash Generator ===")
	fmt.Println()

	// Prompt for password
	fmt.Print("Enter password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Error reading password: %v", err)
	}
	fmt.Println() // New line after password input

	password := string(bytePassword)
	if password == "" {
		log.Fatalf("Password cannot be empty")
	}

	// // Optional: confirm password
	// fmt.Print("Confirm password: ")
	// bytePasswordConfirm, err := term.ReadPassword(int(syscall.Stdin))
	// if err != nil {
	// 	log.Fatalf("Error reading password confirmation: %v", err)
	// // }
	// fmt.Println()

	// if password != string(bytePasswordConfirm) {
	// 	log.Fatalf("Passwords do not match")
	// }

	// Generate hash using existing utils
	hash, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Generate ID
	id := uuid.New().String()

	fmt.Println()
	fmt.Println("=== Generated Credentials ===")
	fmt.Printf("ID: %s\n", id)
	fmt.Printf("Password Hash: %s\n", hash)
	fmt.Println()
	fmt.Println("=== SQL Insert Command ===")
	fmt.Println()

	// For Admin
	fmt.Println("-- For creating admin:")
	fmt.Printf("INSERT INTO admin (id, email, password_hash) VALUES ('%s', 'admin@example.com', '%s');\n", id, hash)
	fmt.Println()

	// For User
	fmt.Printf("-- For creating user (pentester/stakeholder):\n")
	fmt.Printf("INSERT INTO users (id, email, password_hash, role, status) VALUES ('%s', 'user@example.com', '%s', 'pentester', 'active');\n", uuid.New().String(), hash)
	fmt.Println()
	fmt.Println("=== Notes ===")
	fmt.Println("- Replace 'admin@example.com' or 'user@example.com' with actual email")
	fmt.Println("- Copy the password hash exactly as shown")
	fmt.Println("- Execute the SQL command in your PostgreSQL database")
}

func showUsage() {
	fmt.Println("Password Hash Generator - PentSecOps Backend")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd/hashgen/main.go -generate      Generate a password hash")
	fmt.Println("  go run ./cmd/hashgen/main.go -id-only        Generate only a UUID")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  go run ./cmd/hashgen/main.go -generate")
	fmt.Println()
	fmt.Println("The tool will prompt for a password and generate:")
	fmt.Println("  - UUID for database ID")
	fmt.Println("  - Argon2id password hash (using pkg/utils/argon2.go)")
	fmt.Println("  - SQL INSERT commands ready to use")
}
