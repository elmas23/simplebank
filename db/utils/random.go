package utils

import (
	"math/rand"
	"strings"
	"time"
)

// This file is used to generate random int and string for testing

const _alphabet = "abcdefghijklmnopqrstuvwxyz"

// This method will be called whenever we used this package
// we will set the random seed here so that it changes everytime
func init() {
	// we set the seed to the current time and transform to Unix since
	// Seed take an int64
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomInt generates a random integer between min and max
func GenerateRandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1) // +1 to include
}

// GenerateRandomString generates a random string of length n
func GenerateRandomString(n int) string {
	var sb strings.Builder // Initialize a variable for String builder
	k := len(_alphabet)

	// Loop until we have the length our desired string
	for i := 0; i < n; i++ {
		c := _alphabet[rand.Intn(k)] // take a string at a random position in our alphabet
		sb.WriteByte(c)              // it will append the byte c to the buffer of our string builder
	}
	return sb.String() // return the string of our string builder
}

// GenerateOwner generates a random owner name
func GenerateOwner() string {
	return GenerateRandomString(6)
}

// GenerateBalance generates a random amount for the balance
func GenerateBalance() int64 {
	return GenerateRandomInt(0, 1000)
}

// GenerateCurrency generates a random currency code
func GenerateCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

// GenerateAmount generates a random amount for the amount field
func GenerateAmount() int64 {
	return GenerateRandomInt(0, 1000)
}
