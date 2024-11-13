package elog

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	MAX_PREFERED_EXTRA_DATA_SIZE int = 2048
	MAX_PREFERED_LOG_SIZE        int = (1 << 16) - 1 - MAX_PREFERED_EXTRA_DATA_SIZE
)

type entry struct {
	entry_type string
	entry_mode uint64

	msg       []byte
	extra_dat []byte

	log     *Logger
	has_msg bool
}

func (e *entry) Msg(msg string) {
	if e.has_msg {
		panic("message has already been set for this logger entry")
	}
	if !HasType(e.log.Modes, e.entry_mode) {
		free_entry(e)
		return
	}

	if len(e.extra_dat) > 0 {
		e.extra_dat = append(e.extra_dat, ']')
		e.msg = append(e.msg, e.extra_dat...)
		e.msg = append(e.msg, ' ')
	}

	e.msg = append(e.msg, '[')
	e.msg = append(e.msg, e.entry_type...)
	e.msg = append(e.msg, ']')
	e.msg = append(e.msg, ':')

	e.msg = append(e.msg, ' ')
	e.msg = append(e.msg, '"')
	e.msg = append(e.msg, msg...)
	e.msg = append(e.msg, '"')
	e.msg = append(e.msg, '\n')
	e.has_msg = true

	if _, err := e.log.CleanOut.Write(e.msg); err != nil {
		fmt.Println("failed to write to clean out:", err)
	}
	if err := e.log.encrypt_entry(e); err != nil {
		fmt.Println("failed to encrypt entry:", err)
	}
	free_entry(e)
}

func (e *entry) Time() *entry {
	return e.Int("time", time.Now().UnixMilli())
}

func (e *entry) Str(name string, v string) *entry {
	e.init_extradat()
	e.append_prefix(name)
	e.extra_dat = append(e.extra_dat, v...)

	return e
}

func (e *entry) UInt(name string, v uint64) *entry {
	e.init_extradat()
	e.append_prefix(name)
	e.extra_dat = strconv.AppendUint(e.extra_dat, v, 10)

	return e
}

func (e *entry) Int(name string, v int64) *entry {
	e.init_extradat()
	e.append_prefix(name)
	e.extra_dat = strconv.AppendInt(e.extra_dat, v, 10)

	return e
}

func (e *entry) append_prefix(prefix string) {
	if len(e.extra_dat) > 1 {
		e.extra_dat = append(e.extra_dat, ' ')
	}

	e.extra_dat = append(e.extra_dat, prefix...)
	e.extra_dat = append(e.extra_dat, '=')
}

func (e *entry) init_extradat() {
	if len(e.extra_dat) == 0 {
		e.extra_dat = append(e.extra_dat, '[')
	}
}

var entry_pool = sync.Pool{
	New: func() any {
		return &entry{
			msg:       make([]byte, 0, MAX_PREFERED_LOG_SIZE),
			extra_dat: make([]byte, 0, MAX_PREFERED_EXTRA_DATA_SIZE),
		}
	},
}

func free_entry(e *entry) {
	// See: https://github.com/golang/go/issues/23199
	// TLDR: Objects inside a pool should have a (roughly) the same amount of memory utilized.
	if cap(e.msg) > MAX_PREFERED_LOG_SIZE {
		return
	}

	e.log = nil
	entry_pool.Put(e)
}

func new_entry(l *Logger, mode uint64) *entry {
	e := entry_pool.Get().(*entry)
	e.entry_mode = mode
	e.entry_type = Modes[mode]
	e.msg = e.msg[:0]
	e.extra_dat = e.extra_dat[:0]

	e.log = l
	e.has_msg = false

	return e
}