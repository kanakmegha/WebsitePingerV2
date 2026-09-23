package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	// ErrInvalidHash indicates the encoded Argon2id hash string is malformed.
	ErrInvalidHash = errors.New("invalid Argon2id hash format")
	// ErrIncompatibleVersion indicates the hash algorithm is not supported.
	ErrIncompatibleVersion = errors.New("incompatible password hash version")
	// ErrPasswordTooShort indicates the provided password does not meet the minimum length requirement.
	ErrPasswordTooShort = errors.New("password must be at least 6 characters long")
)

// Params holds the Argon2id parameters.
type Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

// DefaultParams provides production-ready Argon2id security settings.
var DefaultParams = &Params{
	Time:    1,
	Memory:  64 * 1024, // 64 MB
	Threads: 4,
	KeyLen:  32,
	SaltLen: 16,
}

// HashPassword creates a secure Argon2id hash of the given password.
// Enforces minimum password length of 6 characters.
// Encoded format: argon2id$<time>$<memory>$<threads>$<salt_base64>$<hash_base64>
func HashPassword(password string) (string, error) {
	if len(password) < 6 {
		return "", ErrPasswordTooShort
	}

	p := DefaultParams
	salt, err := generateSalt(p.SaltLen)
	if err != nil {
		return "", fmt.Errorf("failed to generate random salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)

	return encodeHash(p, salt, hash), nil
}

// VerifyPassword checks if a plain password matches an encoded Argon2id hash.
// Uses constant-time comparison to prevent timing attacks.
func VerifyPassword(encodedHash string, password string) (bool, error) {
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	derivedHash := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)

	if subtle.ConstantTimeCompare(hash, derivedHash) == 1 {
		return true, nil
	}

	return false, nil
}

func generateSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func encodeHash(p *Params, salt []byte, hash []byte) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("argon2id$%d$%d$%d$%s$%s", p.Time, p.Memory, p.Threads, b64Salt, b64Hash)
}

func decodeHash(encodedHash string) (*Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	if parts[0] != "argon2id" {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	timeVal, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: invalid time parameter", ErrInvalidHash)
	}

	memVal, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: invalid memory parameter", ErrInvalidHash)
	}

	threadsVal, err := strconv.ParseUint(parts[3], 10, 8)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: invalid threads parameter", ErrInvalidHash)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: invalid salt encoding", ErrInvalidHash)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: invalid hash encoding", ErrInvalidHash)
	}

	p := &Params{
		Time:    uint32(timeVal),
		Memory:  uint32(memVal),
		Threads: uint8(threadsVal),
		KeyLen:  uint32(len(hash)),
		SaltLen: uint32(len(salt)),
	}

	return p, salt, hash, nil
}
