package telegram

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withDefaultTransport(t *testing.T, rt roundTripFunc) {
	t.Helper()
	original := http.DefaultTransport
	http.DefaultTransport = rt
	t.Cleanup(func() {
		http.DefaultTransport = original
	})
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func newTestLogger() *logrus.Entry {
	return logrus.NewEntry(logrus.New())
}

type mockSubscription struct {
	config   any
	messages []any
}

func (m *mockSubscription) Configuration() any   { return m.config }
func (m *mockSubscription) SendMessage(msg any) error {
	m.messages = append(m.messages, msg)
	return nil
}
