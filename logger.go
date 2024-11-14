package elog

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"os"
	"sync"

	"golang.org/x/exp/rand"
)

type SeedFunc func(seed uint64) uint64

type Logger struct {
	sync.Mutex

	CleanOut io.Writer
	Out      io.WriteCloser

	Modes uint64
	Opts  uint64

	last_hash  []byte
	last_seed  uint64
	seed_func  SeedFunc
	randomizer *rand.Rand
}

func New(f io.WriteCloser, modes, opts, seed uint64, sf SeedFunc) *Logger {
	l := &Logger{
		CleanOut: os.Stdout,
		Out:      f,

		Modes: modes,
		Opts:  opts,

		last_hash:  make([]byte, 32),
		last_seed:  seed,
		seed_func:  sf,
		randomizer: rand.New(rand.NewSource(seed)),
	}

	return l
}

func (l *Logger) Log(log_mode uint64) *entry {
	e := new_entry(l, log_mode)
	if HasType(l.Opts, LoggerOptIncludeTime) {
		e = e.Time()
	}

	return e
}

func (l *Logger) encrypt_entry(e *entry) error {
	l.Lock()
	defer l.Unlock()

	if l.Out == nil {
		return nil
	}

	aes, err := aes.NewCipher(l.last_hash)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return err
	}

	nonce_size := gcm.NonceSize()
	nonce := make([]byte, nonce_size)
	if _, err := l.randomizer.Read(nonce); err != nil {
		return err
	}

	// We remove the nonce from the encrypted entry as we should be able manually
	// calculate it when decoding w/ the seed function.
	msg := e.msg.Bytes()
	ciphertext := append(make([]byte, 8), gcm.Seal(nonce, nonce, msg, nil)[nonce_size:]...)
	binary.LittleEndian.PutUint64(ciphertext, uint64(len(ciphertext)-8))
	_, err = l.Out.Write(ciphertext)

	data_hash := sha256.Sum256(msg)
	copy(l.last_hash, data_hash[:])
	new_seed := l.seed_func(l.last_seed)
	if new_seed == l.last_seed {
		panic("bad seed function: repeated same seed twice")
	}
	l.last_seed = new_seed
	l.randomizer.Seed(l.last_seed)

	return err
}
