package elog

import (
	"fmt"
	"os"
	"testing"
)

func TestDecryptingLog(t *testing.T) {
	os.Remove("my.log")
	os.Remove("my.log.decrypted")
	f, err := os.OpenFile("my.log", os.O_RDWR|os.O_CREATE, 0744)
	if err != nil {
		panic(err)
	}

	l := New(f, WithFlags(LoggerModeTypeInfo, LoggerModeTypeError), WithFlags(LoggerOptIncludeTime), 69420, func(seed uint64) uint64 {
		seed += 1000
		seed *= 2

		if seed > 1_000_000_000 {
			seed -= seed / 4
		}

		return seed
	})
	for i := 0; i < 10; i++ {
		l.Log(LoggerModeTypeInfo).Str("name", "Benjamin Saulon").Int("iteration", int64(i)).Msg("this is a test :)")
	}

	dec, err := os.OpenFile("my.log.decrypted", os.O_CREATE|os.O_RDWR, 0744)
	if err != nil {
		panic(err)
	}

	l.Out.Close()
	l.Out, err = os.OpenFile("my.log", os.O_RDONLY, 0744)
	if err != nil {
		panic(err)
	}

	err = Decrypt(l.Out.(*os.File), dec, SeedOpts{
		Seed: 69420,
		F: func(seed uint64) uint64 {
			seed += 1000
			seed *= 2

			if seed > 1_000_000_000 {
				seed -= seed / 4
			}

			return seed
		},
	})
	fmt.Println(err)
}

func BenchmarkBasicLog(b *testing.B) {
	os.Remove("my.log")
	f, err := os.OpenFile("my.log", os.O_RDWR|os.O_CREATE, 0744)
	if err != nil {
		panic(err)
	}

	var count int64
	l := New(f, WithFlags(LoggerModeTypeInfo, LoggerModeTypeError), WithFlags(LoggerOptIncludeTime), 0, func(seed uint64) uint64 {
		seed += 1000
		seed *= 2

		if seed > 1_000_000_000 {
			seed -= seed / 4
		}

		return seed
	})
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			count++
			l.Log(LoggerModeTypeInfo).
				Int("id", count).
				Msg("hello!")
		}
	})
}
