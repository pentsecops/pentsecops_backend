package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime      = 3
	argonMemory    = 64 * 1024 // 64 MB
	argonThreads   = 4
	argonKeyLength = 32
)

// HashPassword generates an Argon2 hash of the provided password
func HashPassword(password string) (string, error) {
	salt, err := generateSalt(16)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLength,
	)

	// Encode in format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonThreads,
		encodedSalt,
		encodedHash,
	), nil
}

// VerifyPassword compares a password with its hash
func VerifyPassword(password, hash string) (bool, error) {
	// Parse the hash to extract parameters
	var version int
	var memory, timeCost, threads uint32
	var salt, hashBytes []byte
	var err error

	_, err = fmt.Sscanf(hash, "$argon2id$v=%d$m=%d,t=%d,p=%d$", &version, &memory, &timeCost, &threads)
	if err != nil {
		return false, err
	}

	// Extract salt and hash from the encoded string
	parts := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$", version, memory, timeCost, threads)
	remaining := hash[len(parts):]

	// Decode salt and hash
	for i, ch := range remaining {
		if ch == '$' {
			salt, _ = base64.RawStdEncoding.DecodeString(remaining[:i])
			hashBytes, _ = base64.RawStdEncoding.DecodeString(remaining[i+1:])
			break
		}
	}

	// Recompute hash with extracted parameters
	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		timeCost,
		memory,
		uint8(threads),
		uint32(len(hashBytes)),
	)

	// Compare
	if len(newHash) != len(hashBytes) {
		return false, nil
	}

	for i, b := range newHash {
		if b != hashBytes[i] {
			return false, nil
		}
	}

	return true, nil
}

// generateSalt creates a random salt for password hashing
func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// GenerateRandomPassword generates a random password of specified length
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 8 // Minimum password length
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)

	for i := 0; i < length; i++ {
		randomByte := make([]byte, 1)
		_, err := rand.Read(randomByte)
		if err != nil {
			return "", err
		}
		password[i] = charset[randomByte[0]%byte(len(charset))]
	}

	return string(password), nil
}
