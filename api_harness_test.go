package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/middleware/auth"
	"linhazero.com/pvapi/middleware/logmw"
	"linhazero.com/pvapi/models/apitokenmodel"
	"linhazero.com/pvapi/seed"
)

type apiSession struct {
	cookies     []*http.Cookie
	bearerToken string
}

type uploadFile struct {
	Field    string
	Filename string
	Data     []byte
}

type apiHarness struct {
	router             *gin.Engine
	pool               *pgxpool.Pool
	apiClientToken     string
	defaultAPISession *apiSession
}

var png1x1 = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x04, 0x00, 0x00, 0x00, 0xb5, 0x1c, 0x0c,
	0x02, 0x00, 0x00, 0x00, 0x0b, 0x49, 0x44, 0x41,
	0x54, 0x78, 0xda, 0x63, 0xfc, 0xff, 0x1f, 0x00,
	0x03, 0x03, 0x02, 0x00, 0xef, 0x45, 0xca, 0x95,
	0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
	0xae, 0x42, 0x60, 0x82,
}

func newAPIHarness(t *testing.T) *apiHarness {
	t.Helper()

	setDefaultEnv(t, "PV_API_SECURE_COOKIE", "no")
	setDefaultEnv(t, "PV_API_SECRET", "secret")
	setDefaultEnv(t, "PV_API_DB_HOST", "127.0.0.1")
	setDefaultEnv(t, "PV_API_DB_USER", "test_user")
	setDefaultEnv(t, "PV_API_DB_PASS", "testing_database")
	setDefaultEnv(t, "PV_API_DB_BASE", "test_db")
	setDefaultEnv(t, "PV_API_DB_PORT", "5434")
	setDefaultEnv(t, "PV_API_ADMIN_USER", "lzero")
	setDefaultEnv(t, "PV_API_ADMIN_MAIL", "sistemas@lzero.org")
	setDefaultEnv(t, "PV_API_ADMIN_PASS", "TestingPassword")
	setDefaultEnv(t, "PV_API_LOG_LEVEL", "debug")
	setDefaultEnv(t, "PV_API_DOMAIN_NAME", "localhost")
	setDefaultEnv(
		t,
		"PV_API_MP_TOKEN",
		"TEST-placeholder-mercadopago-token",
	)
	setDefaultEnv(t, "PV_API_TEST_MODE", "yes")

	if err := os.MkdirAll("public/uploads", 0o755); err != nil {
		t.Fatalf("create uploads dir: %v", err)
	}

	gin.SetMode(gin.TestMode)

	pool, err := db.NewPool()
	if err != nil {
		t.Skipf("test database not configured: %v", err)
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		t.Skipf("test database is not reachable; start it with test.just first: %v", err)
	}

	out := io.Discard
	logger := slog.New(slog.NewTextHandler(out, nil))

	if err := seed.Plant(logger, context.Background(), pool); err != nil {
		pool.Close()
		t.Fatalf("seed admin user: %v", err)
	}

	params, err := auth.GenerateJwtInitParams(pool)
	if err != nil {
		pool.Close()
		t.Fatalf("generate auth params: %v", err)
	}

	authMiddleware, err := jwt.New(params)
	if err != nil {
		pool.Close()
		t.Fatalf("build auth middleware: %v", err)
	}

	if err := authMiddleware.MiddlewareInit(); err != nil {
		pool.Close()
		t.Fatalf("init auth middleware: %v", err)
	}

	router := gin.New()
	router.Use(logmw.LogMiddleware(logger))
	router.Use(gin.Recovery())
	setupRoutes(router, authMiddleware, pool)

	t.Cleanup(pool.Close)

	// Create a default API client for tests that hit /checkouts/confirm
	_, apiToken, err := apitokenmodel.Create(
		context.Background(), pool,
		"Test API Client", "test-api@lzero.org",
	)
	if err != nil {
		pool.Close()
		t.Fatalf("create API client: %v", err)
	}

	return &apiHarness{
		router:         router,
		pool:           pool,
		apiClientToken: apiToken,
		defaultAPISession: &apiSession{
			bearerToken: apiToken,
		},
	}
}

func setDefaultEnv(t *testing.T, key, value string) {
	t.Helper()
	if _, ok := os.LookupEnv(key); !ok {
		t.Setenv(key, value)
	}
}

func (h *apiHarness) login(t *testing.T, email, password string) apiSession {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/login", nil, map[string]any{
		"email":    email,
		"password": password,
	})
	assertStatus(t, rec, http.StatusOK)

	return apiSession{
		cookies: rec.Result().Cookies(),
	}
}

func (h *apiHarness) doJSON(
	t *testing.T,
	method, path string,
	session *apiSession,
	payload any,
) *httptest.ResponseRecorder {
	t.Helper()

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(data)
	}

	return h.doRequest(t, method, path, session, "application/json", body)
}

func (h *apiHarness) doMultipart(
	t *testing.T,
	method, path string,
	session *apiSession,
	fields map[string]string,
	files []uploadFile,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write multipart field %q: %v", key, err)
		}
	}

	for _, file := range files {
		part, err := writer.CreateFormFile(file.Field, file.Filename)
		if err != nil {
			t.Fatalf("create multipart file %q: %v", file.Field, err)
		}
		if _, err := part.Write(file.Data); err != nil {
			t.Fatalf("write multipart file %q: %v", file.Field, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart body: %v", err)
	}

	return h.doRequest(t, method, path, session, writer.FormDataContentType(), &body)
}

func (h *apiHarness) doRequest(
	t *testing.T,
	method, path string,
	session *apiSession,
	contentType string,
	body io.Reader,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if session != nil {
		for _, cookie := range session.cookies {
			req.AddCookie(cookie)
		}
		if session.bearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+session.bearerToken)
		}
	}

	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)
	return rec
}
