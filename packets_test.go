package spdy

import (
	"bytes"
	"http"
	"os"
	"reflect"
	"testing"
	"url"
)

type tester struct {
	t     *testing.T
	data  []byte
	buf   bytes.Buffer
	zip   compressor
	unzip decompressor
}

func newTester(t *testing.T) *tester {
	return &tester{t: t}
}

func (s *tester) test(f frame, parse func() (frame, os.Error)) {
	s.buf.Reset()
	if err := f.WriteFrame(&s.buf, &s.zip); err != nil {
		s.t.Fatalf("%v %+v", err, f)
	}

	s.data = s.buf.Bytes()
	f2, err := parse()

	if err != nil {
		s.t.Fatalf("%v %+v", err, f)
	}

	if !reflect.DeepEqual(f, f2) {
		s.t.Fatalf("%#v\n%#v", f, f2)
	}
}

var testurl, _ = url.Parse("https://www.example.com/foo?bar=3")

var requests = []synStreamFrame{
	{
		Finished:           true,
		Unidirectional:     true,
		StreamId:           3,
		AssociatedStreamId: 2,
		URL:                testurl,
		Proto:              "HTTP/1.1",
		ProtoMajor:         1,
		ProtoMinor:         1,
		Method:             "GET",
	},
	{
		Finished:       false,
		Unidirectional: false,
		URL:            testurl,
		Proto:          "HTTP/1.1",
		ProtoMajor:     1,
		ProtoMinor:     1,
	},
}

var replies = []synReplyFrame{
	{
		Finished:   true,
		StreamId:   50,
		Status:     "202 OK",
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     nil,
	},
	{
		StreamId:   50,
		Status:     "202 OK",
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Foo":  {"bar", "bar3"},
			"Foo2": {"bar5", "bar6"},
		},
	},
}

var headers = []headersFrame{
	{
		Finished: true,
		StreamId: 3,
		Header:   nil,
	},
	{
		Finished: false,
		StreamId: 5,
		Header:   http.Header{},
	},
	{
		Finished: false,
		StreamId: 7,
		Header: http.Header{
			"Foo": {"bar", "bar2"},
		},
	},
}

func TestSynStreamFrame(t *testing.T) {
	s := newTester(t)
	for _, f := range requests {
		f.Version = 2
		s.test(&f, func() (frame, os.Error) {
			return parseSynStream(s.data, &s.unzip)
		})

		f.Version = 3
		s.test(&f, func() (frame, os.Error) {
			return parseSynStream(s.data, &s.unzip)
		})
	}
}

func TestSynReplyFrame(t *testing.T) {
	s := newTester(t)
	for _, f := range replies {
		f.Version = 2
		s.test(&f, func() (frame, os.Error) {
			return parseSynReply(s.data, &s.unzip)
		})

		f.Version = 3
		s.test(&f, func() (frame, os.Error) {
			return parseSynReply(s.data, &s.unzip)
		})
	}
}

func TestHeadersFrame(t *testing.T) {
	s := newTester(t)
	for _, f := range headers {
		f.Version = 2
		s.test(&f, func() (frame, os.Error) {
			return parseHeaders(s.data, &s.unzip)
		})

		f.Version = 3
		s.test(&f, func() (frame, os.Error) {
			return parseHeaders(s.data, &s.unzip)
		})
	}
}
