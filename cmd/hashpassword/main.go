package main

import (
	"flag"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := flag.String("password", "", "Password to hash")
	cost := flag.Int("cost", bcrypt.DefaultCost, "Bcrypt cost (higher is more secure but slower)")
	flag.Parse()

	if *password == "" {
		log.Fatal("Password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), *cost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	fmt.Println("Password hash:")
	fmt.Println(string(hash))
}
