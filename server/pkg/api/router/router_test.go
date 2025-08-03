// server/pkg/api/router/router_test.go
package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"dployr.io/internal/testdb"
	"dployr.io/internal/testutil"
	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/api/middleware"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/queue"
	"dployr.io/pkg/repository"
)

// setupTestRepos creates test repositories for router tests
func setupTestRepos(t *testing.T, db *sqlx.DB) *repository.AppRepos {
	t.Helper()

	return &repository.AppRepos{
		UserRepo:         repository.NewUserRepo(db),
		MagicTokenRepo:   repository.NewMagicTokenRepo(db),
		ProjectRepo:      repository.NewProjectRepo(db),
		EventRepo:        repository.NewEventRepo(db),
		DeploymentRepo:   repository.NewDeploymentRepo(db),
		LogRepo:          repository.NewLogRepo(db),
		RefreshTokenRepo: repository.NewRefreshTokenRepo(db),
	}
}

func TestRouter_New(t *testing.T) {
	// Setup dependencies
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repos := setupTestRepos(t, db)

	q := queue.NewQueue(1, time.Second, queue.CreateHandler())
	defer q.Stop()

	ssh := platform.NewSshManager()
	rl := middleware.NewRateLimiter(10, time.Minute)
	jwtManager := auth.NewJWTManager()
	staticFiles := testutil.MockStaticFiles{}

	// Create router
	r := New(repos, q, ssh, rl, jwtManager, staticFiles)

	assert.NotNil(t, r)
}

func TestRouter_PublicRoutes(t *testing.T) {
	// Setup dependencies
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repos := setupTestRepos(t, db)

	q := queue.NewQueue(1, time.Second, queue.CreateHandler())
	defer q.Stop()

	ssh := platform.NewSshManager()
	rl := middleware.NewRateLimiter(10, time.Minute)
	jwtManager := auth.NewJWTManager()
	staticFiles := testutil.MockStaticFiles{}

	r := New(repos, q, ssh, rl, jwtManager, staticFiles)

	publicRoutes := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/", http.StatusOK},
		{"GET", "/health", http.StatusOK},
		{"GET", "/favicon.ico", http.StatusOK},
	}

	for _, route := range publicRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, route.status, w.Code)
		})
	}
}

func TestRouter_ProtectedRoutes(t *testing.T) {
	// Setup dependencies
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repos := setupTestRepos(t, db)

	q := queue.NewQueue(1, time.Second, queue.CreateHandler())
	defer q.Stop()

	ssh := platform.NewSshManager()
	rl := middleware.NewRateLimiter(10, time.Minute)
	jwtManager := auth.NewJWTManager()
	staticFiles := testutil.MockStaticFiles{}

	r := New(repos, q, ssh, rl, jwtManager, staticFiles)

	protectedRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/api/projects"},
		{"POST", "/v1/api/projects"},
		{"GET", "/v1/api/deployments"},
		{"POST", "/v1/api/deployments"},
	}

	for _, route := range protectedRoutes {
		t.Run(route.method+" "+route.path+" without auth", func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// Should return 401 without authentication
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestRouter_CORS(t *testing.T) {
	// Setup dependencies
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repos := setupTestRepos(t, db)

	q := queue.NewQueue(1, time.Second, queue.CreateHandler())
	defer q.Stop()

	ssh := platform.NewSshManager()
	rl := middleware.NewRateLimiter(10, time.Minute)
	jwtManager := auth.NewJWTManager()
	staticFiles := testutil.MockStaticFiles{}

	r := New(repos, q, ssh, rl, jwtManager, staticFiles)

	// Test CORS preflight request
	req := httptest.NewRequest("OPTIONS", "/v1/api/projects", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should handle CORS preflight
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}
