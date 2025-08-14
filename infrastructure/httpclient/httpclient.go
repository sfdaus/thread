package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// ============================ Public API ============================

// JSONDoer memudahkan mocking di unit test.
type JSONDoer interface {
	GetJSON(ctx context.Context, target string, out any, headers map[string]string) error
	PostJSON(ctx context.Context, target string, body any, out any, headers map[string]string) error
	PutJSON(ctx context.Context, target string, body any, out any, headers map[string]string) error
	PatchJSON(ctx context.Context, target string, body any, out any, headers map[string]string) error
	DeleteJSON(ctx context.Context, target string, out any, headers map[string]string) error
}

// ErrorMapper bisa dipakai buat konversi ke error tipe internal project (mis. utils.HttpErr).
type ErrorMapper func(resp *http.Response, body []byte) error

type Hooks struct {
	// Before dipanggil setiap attempt tepat sebelum request dikirim.
	Before func(req *http.Request, attempt int)
	// After dipanggil setiap attempt tepat setelah response/err diterima.
	After func(req *http.Request, resp *http.Response, err error, duration time.Duration, attempt int)
}

type Config struct {
	BaseURL string

	// Default header untuk semua request (bisa dioverride di argumen method).
	DefaultHeaders map[string]string
	UserAgent      string // default: "<pkg>/1.0"

	// Retry
	MaxRetries         int           // default: 2 (total 3 attempt)
	RetryNonIdempotent bool          // default: false
	BackoffInitial     time.Duration // default: 200ms
	BackoffMax         time.Duration // default: 5s
	RespectRetryAfter  bool          // default: true
	FollowRedirects    bool          // default: true (pakai default client)
	SuccessBodyLimit   int64         // default: 2 MiB (untuk decode sukses)
	ErrorBodyLimit     int64         // default: 1 MiB (untuk baca error)
	DefaultReqTimeout  time.Duration // default: 10s (dipakai jika ctx tidak punya deadline)
	Transport          http.RoundTripper
	ErrorMapper        ErrorMapper // optional
	Hooks              Hooks       // optional
}

type Client struct {
	hc   *http.Client
	conf Config
}

// New membuat client dengan transport yang sehat; tidak memakai http.Client.Timeout.
// Timeout diatur per-request dari context (ctx) atau DefaultReqTimeout kalau ctx tidak punya deadline.
func New(conf Config) *Client {
	if conf.UserAgent == "" {
		conf.UserAgent = "httpclient-prakarsa/1.0"
	}
	if conf.MaxRetries < 0 {
		conf.MaxRetries = 0
	}
	if conf.BackoffInitial <= 0 {
		conf.BackoffInitial = 200 * time.Millisecond
	}
	if conf.BackoffMax <= 0 {
		conf.BackoffMax = 5 * time.Second
	}
	if conf.SuccessBodyLimit <= 0 {
		conf.SuccessBodyLimit = 2 << 20 // 2 MiB
	}
	if conf.ErrorBodyLimit <= 0 {
		conf.ErrorBodyLimit = 1 << 20 // 1 MiB
	}
	if conf.DefaultReqTimeout <= 0 {
		conf.DefaultReqTimeout = 10 * time.Second
	}
	if conf.Transport == nil {
		// clone transport default agar aman dipakai parallel & bisa kita tuning.
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.MaxIdleConns = 100
		tr.MaxIdleConnsPerHost = 10
		tr.IdleConnTimeout = 90 * time.Second
		tr.TLSHandshakeTimeout = 10 * time.Second
		tr.ResponseHeaderTimeout = 5 * time.Second
		conf.Transport = tr
	}
	hc := &http.Client{
		Transport: conf.Transport,
		// Biarkan Timeout kosong; pakai ctx deadline per request.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if conf.FollowRedirects {
				return nil
			}
			return http.ErrUseLastResponse
		},
	}
	return &Client{hc: hc, conf: conf}
}

// Helpers JSON ergonomis
func (c *Client) GetJSON(ctx context.Context, target string, out any, headers map[string]string) error {
	return c.doJSON(ctx, http.MethodGet, target, nil, out, headers)
}
func (c *Client) PostJSON(ctx context.Context, target string, body any, out any, headers map[string]string) error {
	return c.doJSON(ctx, http.MethodPost, target, body, out, headers)
}
func (c *Client) PutJSON(ctx context.Context, target string, body any, out any, headers map[string]string) error {
	return c.doJSON(ctx, http.MethodPut, target, body, out, headers)
}
func (c *Client) PatchJSON(ctx context.Context, target string, body any, out any, headers map[string]string) error {
	return c.doJSON(ctx, http.MethodPatch, target, body, out, headers)
}
func (c *Client) DeleteJSON(ctx context.Context, target string, out any, headers map[string]string) error {
	return c.doJSON(ctx, http.MethodDelete, target, nil, out, headers)
}

// ============================ Core ============================

func (c *Client) doJSON(ctx context.Context, method, target string, body any, out any, headers map[string]string) error {
	// Pastikan ada deadline: kalau ctx belum ada, apply default.
	if _, ok := ctx.Deadline(); !ok && c.conf.DefaultReqTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.conf.DefaultReqTimeout)
		defer cancel()
	}

	reqURL, err := c.buildURL(target)
	if err != nil {
		return &HTTPError{StatusCode: http.StatusBadRequest, Message: "invalid url: " + err.Error()}
	}

	// Marshal body JSON sekali agar idempotent untuk retry.
	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return &HTTPError{StatusCode: http.StatusBadRequest, Message: "failed to encode json body: " + err.Error()}
		}
	}

	maxAttempts := c.conf.MaxRetries + 1
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// build request baru tiap attempt (body reader fresh).
		var reader io.ReadCloser
		if bodyBytes != nil {
			reader = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
		if err != nil {
			return &HTTPError{StatusCode: http.StatusBadRequest, Message: "failed to build request: " + err.Error()}
		}

		// Default headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.conf.UserAgent)
		for k, v := range c.conf.DefaultHeaders {
			if v != "" && req.Header.Get(k) == "" {
				req.Header.Set(k, v)
			}
		}
		// Body content-type
		if bodyBytes != nil && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
		// Override dari pemanggil
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		if c.conf.Hooks.Before != nil {
			c.conf.Hooks.Before(req, attempt)
		}
		start := time.Now()
		resp, err := c.hc.Do(req)
		dur := time.Since(start)
		if c.conf.Hooks.After != nil {
			c.conf.Hooks.After(req, resp, err, dur, attempt)
		}

		// Network error → bisa retry?
		if err != nil {
			if shouldRetryNetwork(err) && c.canRetry(method, attempt, maxAttempts) {
				time.Sleep(c.backoff(attempt, nil))
				continue
			}
			return wrapNetErr(err)
		}

		// Kita punya response
		if is2xx(resp.StatusCode) {
			// success path
			if out == nil {
				// habiskan body sedikit agar koneksi reusable
				drainAndClose(resp.Body)
				return nil
			}
			// baca terbatas → decode
			content, readErr := readLimited(resp.Body, c.conf.SuccessBodyLimit)
			_ = readErr // readErr berupa EOF biasa, aman diabaikan
			_ = resp.Body.Close()

			ct := resp.Header.Get("Content-Type")
			// Guard: kalau bukan JSON, coba deteksi heuristik kecil lalu fallback error bila gagal.
			if !isJSONContentType(ct) && !looksLikeJSON(content) {
				return &HTTPError{
					StatusCode: resp.StatusCode,
					Message:    "unexpected content-type for JSON: " + ct,
					Body:       string(sample(content, 512)),
					Headers:    resp.Header.Clone(),
				}
			}
			if err := json.Unmarshal(content, out); err != nil {
				return &HTTPError{
					StatusCode: http.StatusBadGateway,
					Message:    "failed to decode response: " + err.Error(),
					Body:       string(sample(content, 1024)),
					Headers:    resp.Header.Clone(),
				}
			}
			return nil
		}

		// Non-2xx → cek apakah eligible retry
		errBody, _ := readLimited(resp.Body, c.conf.ErrorBodyLimit)
		_ = resp.Body.Close()
		if shouldRetryStatus(resp.StatusCode) && c.canRetry(method, attempt, maxAttempts) {
			time.Sleep(c.backoff(attempt, resp))
			continue
		}

		// Map error
		if c.conf.ErrorMapper != nil {
			return c.conf.ErrorMapper(resp, errBody)
		}
		return defaultMapHTTPError(resp, errBody)
	}

	// Harusnya gak nyampe sini.
	return &HTTPError{StatusCode: http.StatusInternalServerError, Message: "unreachable retry state"}
}

// ============================ Internals ============================

func (c *Client) buildURL(target string) (string, error) {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target, nil
	}
	if c.conf.BaseURL == "" {
		return "", errors.New("BaseURL empty and target is relative")
	}
	base, err := url.Parse(c.conf.BaseURL)
	if err != nil {
		return "", err
	}
	// join path + preserve query dari target
	tgt, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	base.Path = path.Join(strings.TrimSuffix(base.Path, "/"), tgt.Path)
	q := base.Query()
	for k, vs := range tgt.Query() {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	base.RawQuery = q.Encode()
	return base.String(), nil
}

func (c *Client) canRetry(method string, attempt, maxAttempts int) bool {
	if attempt >= maxAttempts-1 {
		return false
	}
	if c.conf.RetryNonIdempotent {
		return true
	}
	return isIdempotent(method)
}

func (c *Client) backoff(attempt int, resp *http.Response) time.Duration {
	// Hormati Retry-After bila ada
	if resp != nil && c.conf.RespectRetryAfter {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			// numeric seconds
			if secs, _ := strconv.Atoi(ra); secs > 0 {
				return clamp(time.Duration(secs)*time.Second, 0, c.conf.BackoffMax)
			}
			// HTTP-date
			if t, err := http.ParseTime(ra); err == nil {
				d := time.Until(t)
				if d > 0 {
					return clamp(d, 0, c.conf.BackoffMax)
				}
			}
		}
	}
	// Exponential backoff + jitter
	base := c.conf.BackoffInitial * (1 << attempt)
	if base > c.conf.BackoffMax {
		base = c.conf.BackoffMax
	}
	jitter := time.Duration(rand.Int64N(int64(base/5 + 1)))
	return base/2 + jitter // sebaran [base/2, ~base*0.7]
}

func shouldRetryNetwork(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) && (ne.Timeout() || ne.Temporary()) {
		return true
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "broken pipe"),
		strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "no such host"),
		strings.Contains(msg, "server misbehaving"):
		return true
	}
	return false
}

func shouldRetryStatus(code int) bool {
	return code == http.StatusTooManyRequests || (code >= 500 && code <= 599)
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodPut, http.MethodOptions:
		return true
	default:
		return false
	}
}

func is2xx(code int) bool { return code >= 200 && code < 300 }

func isJSONContentType(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "application/json") || strings.Contains(ct, "+json")
}

func looksLikeJSON(b []byte) bool {
	s := strings.TrimSpace(string(sample(b, 1)))
	return len(s) > 0 && (s[0] == '{' || s[0] == '[')
}

func readLimited(rc io.ReadCloser, limit int64) ([]byte, error) {
	defer func() {
		// kalau masih tersisa, drain sedikit biar koneksi reusable
		drainAndClose(rc)
	}()
	lr := io.LimitReader(rc, limit)
	return io.ReadAll(lr)
}

func drainAndClose(rc io.ReadCloser) {
	if rc == nil {
		return
	}
	// drain sedikit saja agar koneksi bisa di-reuse oleh pool
	io.CopyN(io.Discard, rc, 512)
	rc.Close()
}

func clamp(d, min, max time.Duration) time.Duration {
	if d < min {
		return min
	}
	if d > max {
		return max
	}
	return d
}

func sample(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[:n]
}

// ============================ Errors ============================

// HTTPError error generik untuk HTTP non-2xx / kesalahan client.
type HTTPError struct {
	StatusCode int
	Message    string
	Body       string
	Headers    http.Header
}

func (e *HTTPError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Message != "" {
		return fmt.Sprintf("http error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("http error %d", e.StatusCode)
}

func wrapNetErr(err error) error {
	// Map network error → 503 agar calling layer bisa uniform.
	return &HTTPError{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "request failed: " + err.Error(),
	}
}

func defaultMapHTTPError(resp *http.Response, body []byte) error {
	// Coba parse JSON error umum: {error:, message:, details:, status:}
	var payload struct {
		Status  int         `json:"status"`
		Error   string      `json:"error"`
		Message string      `json:"message"`
		Details interface{} `json:"details"`
	}
	_ = json.Unmarshal(body, &payload)

	msg := strings.TrimSpace(firstNonEmpty(payload.Error, payload.Message, http.StatusText(resp.StatusCode)))
	if msg == "" {
		msg = "unexpected response status"
	}
	return &HTTPError{
		StatusCode: resp.StatusCode,
		Message:    msg,
		Body:       string(sample(body, 2048)),
		Headers:    resp.Header.Clone(),
	}
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
