package setup

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDecideAdminBootstrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		totalUsers int64
		adminUsers int64
		should     bool
		reason     string
	}{
		{
			name:       "empty database should create admin",
			totalUsers: 0,
			adminUsers: 0,
			should:     true,
			reason:     adminBootstrapReasonEmptyDatabase,
		},
		{
			name:       "admin exists should skip",
			totalUsers: 10,
			adminUsers: 1,
			should:     false,
			reason:     adminBootstrapReasonAdminExists,
		},
		{
			name:       "users exist without admin should skip",
			totalUsers: 5,
			adminUsers: 0,
			should:     false,
			reason:     adminBootstrapReasonUsersExistWithoutAdmin,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := decideAdminBootstrap(tc.totalUsers, tc.adminUsers)
			if got.shouldCreate != tc.should {
				t.Fatalf("shouldCreate=%v, want %v", got.shouldCreate, tc.should)
			}
			if got.reason != tc.reason {
				t.Fatalf("reason=%q, want %q", got.reason, tc.reason)
			}
		})
	}
}

func TestSetupDefaultAdminConcurrency(t *testing.T) {
	t.Run("simple mode admin uses higher concurrency", func(t *testing.T) {
		t.Setenv("RUN_MODE", "simple")
		if got := setupDefaultAdminConcurrency(); got != simpleModeAdminConcurrency {
			t.Fatalf("setupDefaultAdminConcurrency()=%d, want %d", got, simpleModeAdminConcurrency)
		}
	})

	t.Run("standard mode keeps existing default", func(t *testing.T) {
		t.Setenv("RUN_MODE", "standard")
		if got := setupDefaultAdminConcurrency(); got != defaultUserConcurrency {
			t.Fatalf("setupDefaultAdminConcurrency()=%d, want %d", got, defaultUserConcurrency)
		}
	})
}

func TestWriteConfigFileKeepsDefaultUserConcurrency(t *testing.T) {
	t.Setenv("RUN_MODE", "simple")
	t.Setenv("DATA_DIR", t.TempDir())

	if err := writeConfigFile(&SetupConfig{}); err != nil {
		t.Fatalf("writeConfigFile() error = %v", err)
	}

	data, err := os.ReadFile(GetConfigFilePath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(data), "user_concurrency: 5") {
		t.Fatalf("config missing default user concurrency, got:\n%s", string(data))
	}
}

func TestSetupTokenValidation(t *testing.T) {
	token, tokenHash, err := NewSetupToken()
	if err != nil {
		t.Fatalf("NewSetupToken() error = %v", err)
	}
	if token == "" || tokenHash == "" {
		t.Fatal("NewSetupToken() returned empty token or hash")
	}
	if !validateSetupToken(token, tokenHash) {
		t.Fatal("valid setup token was rejected")
	}
	if validateSetupToken(token+"x", tokenHash) {
		t.Fatal("invalid setup token was accepted")
	}
	if validateSetupToken("", tokenHash) || validateSetupToken(token, "") {
		t.Fatal("empty setup token inputs should be rejected")
	}
}

func TestSetupLocalRequestGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		remoteAddr string
		want       bool
	}{
		{name: "ipv4_loopback", remoteAddr: "127.0.0.1:1234", want: true},
		{name: "ipv6_loopback", remoteAddr: "[::1]:1234", want: true},
		{name: "remote", remoteAddr: "203.0.113.10:1234", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/setup/install", nil)
			c.Request.RemoteAddr = tc.remoteAddr
			if got := isLocalSetupRequest(c); got != tc.want {
				t.Fatalf("isLocalSetupRequest()=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestAutoSetupGeneratedAdminPasswordOptIn(t *testing.T) {
	t.Setenv("AUTO_SETUP_ALLOW_GENERATED_ADMIN_PASSWORD", "")
	if autoSetupAllowGeneratedAdminPassword() {
		t.Fatal("generated admin password should be disabled by default")
	}

	t.Setenv("AUTO_SETUP_ALLOW_GENERATED_ADMIN_PASSWORD", "true")
	if !autoSetupAllowGeneratedAdminPassword() {
		t.Fatal("generated admin password opt-in was not recognized")
	}
}
