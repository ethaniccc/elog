package elog

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"os"

	"golang.org/x/exp/rand"
)

type SeedOpts struct {
	Seed uint64
	F    SeedFunc
}

func Decrypt(encrypted, decrypted *os.File, opts SeedOpts) error {
	encrypted_data, err := io.ReadAll(encrypted)
	if err != nil {
		return err
	}
	encrypted_length := len(encrypted_data)
	current_index := 0
	current_hash := make([]byte, 32)
	current_seed := opts.Seed
	randomizer := rand.New(rand.NewSource(opts.Seed))

	for current_index < encrypted_length {
		entry_length_bytes := encrypted_data[current_index : current_index+8]
		entry_length := int(binary.LittleEndian.Uint64(entry_length_bytes))
		current_index += 8

		ciphertext := encrypted_data[current_index : current_index+entry_length]
		current_index += entry_length

		aes, err := aes.NewCipher(current_hash)
		if err != nil {
			return err
		}
		gcm, err := cipher.NewGCM(aes)
		if err != nil {
			return err
		}

		// We predict the nonce using the randomizer w/ the given seed.
		nonce := make([]byte, gcm.NonceSize())
		randomizer.Read(nonce)
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		// Unless the inital seed or seed function is wrong, this should never happen.
		if err != nil {
			return err
		}

		// Write the plain-text log entry into the plain-text log file.
		decrypted.Write(plaintext)
		new_hash := sha256.Sum256(plaintext)
		copy(current_hash, new_hash[:])
		current_seed = opts.F(current_seed)
		randomizer.Seed(current_seed)
	}

	return nil
}
