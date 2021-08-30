package event

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/tmaxmax/go-sse/internal/parser"
)

type field interface {
	name() parser.FieldName
	repr() []byte

	Option
}

// Event is the representation of a single message. Use the New constructor to create one.
type Event struct {
	expiresAt  time.Time
	fields     []field
	nameIndex  int
	idIndex    int
	retryIndex int
}

func (e *Event) WriteTo(w io.Writer) (int64, error) {
	fw := &fieldWriter{
		w: w,
		s: parser.ChunkScanner{},
	}

	n, m, err := 0, 0, error(nil)

	for _, f := range e.fields {
		m, err = fw.writeField(f)
		n += m
		if err != nil {
			return int64(n), err
		}
	}

	m, err = w.Write(newline)
	n += m

	return int64(n), err
}

func (e *Event) MarshalText() ([]byte, error) {
	b := &bytes.Buffer{}
	_, _ = e.WriteTo(b)

	return b.Bytes(), nil
}

func (e *Event) String() string {
	s := &strings.Builder{}
	_, _ = e.WriteTo(s)

	return s.String()
}

// ID returns the event's ID. It returns an empty string if the event doesn't have an ID.
func (e *Event) ID() ID {
	if e.idIndex == -1 {
		return ""
	}
	return e.fields[e.idIndex].(ID)
}

// ExpiresAt returns the timestamp when the event expires.
func (e *Event) ExpiresAt() time.Time {
	return e.expiresAt
}

// New creates a new event. It takes as parameters the event's desired fields and an expiry time configuration
// (TTL or ExpiresAt). If no expiry time is specified, the event expires immediately.
func New(options ...Option) *Event {
	e := &Event{
		nameIndex:  -1,
		idIndex:    -1,
		retryIndex: -1,
	}

	for _, option := range options {
		option.apply(e)
	}

	return e
}

// From creates a new event using the provided one as a base. It does not modify the base event.
func From(base *Event, options ...Option) *Event {
	e := &Event{
		nameIndex:  base.nameIndex,
		idIndex:    base.idIndex,
		retryIndex: base.retryIndex,
		expiresAt:  base.expiresAt,
		fields:     make([]field, 0, len(base.fields)),
	}

	e.fields = append(e.fields, base.fields...)

	for _, option := range options {
		option.apply(e)
	}

	return e
}
