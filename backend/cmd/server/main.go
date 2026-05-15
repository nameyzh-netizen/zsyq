package main

//go:generate go run github.com/google/wire/cmd/wire

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/nameyzh-netizen/zsyq/ent/runtime"
	"github.com/nameyzh-netizen/zsyq/internal/config"
	"github.com/nameyzh-netizen/zsyq/internal/handler"
	"github.com/nameyzh-netizen/zsyq/internal/pkg/logger"
	"github.com/nameyzh-netizen/zsyq/internal/server/middleware"
	"github.com/nameyzh-netizen/zsyq/internal/setup"
	"github.com/nameyzh-netizen/zsyq/internal/web"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

//go:embed VERSION
var embeddedVersion string

// Build-time variables (can be set by ldflags)
var (
	Version   = ""
	Commit    = "unknown"
	Date      = "unknown"
	BuildType = "source" // "source" for manual builds, "release" for CI builds (set by ldflags)
)

func init() {
	// 如果 Version 已通过 ldflags 注入（例如 -X main.Version=...），则不要覆盖。
	if strings.TrimSpace(Version) != "" {
		return
	}

	// 默认从 embedded VERSION 文件读取版本号（编译期打包进二进制）。
	Version = strings.TrimSpace(embeddedVersion)
	if Version == "" {
		Version = "0.0.0-dev"
	}
}

// initLogger configures the default slog handler based on gin.Mode().
// In non-release mode, Debug level logs are enabled.
func main() {
	logger.InitBootstrap()
	defer logger.Sync()

	// Parse command line flags
	setupMode := flag.Bool("setup", false, "Run setup wizard in CLI mode")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		log.Printf("智算引擎 %s (commit: %s, built: %s)\n", Version, Commit, Date)
		return
	}

	// CLI setup mode
	if *setupMode {
		if err := setup.RunCLI(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		return
	}

	// Check if setup is needed
	if setup.NeedsSetup() {
		// Check if auto-setup is enabled (for Docker deployment)
		if setup.AutoSetupEnabled() {
			log.Println("Auto setup mode enabled...")
			if err := setup.AutoSetupFromEnv(); err != nil {
				log.Fatalf("Auto setup failed: %v", err)
			}
			// Continue to main server after auto-setup
		} else {
			log.Println("First run detected, starting setup wizard...")
			runSetupServer()
			return
		}
	}

	// Normal server mode
	runMainServer()
}

func runSetupServer() {
	token, tokenHash, err := setup.NewSetupToken()
	if err != nil {
		log.Fatalf("Failed to create setup token: %v", err)
	}
	addr, publicAddr, allowRemote := setupServerAddress()

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(config.CORSConfig{}))
	r.Use(middleware.SecurityHeaders(config.CSPConfig{Enabled: true, Policy: config.DefaultCSPPolicy}, nil))

	// Register setup routes
	setup.RegisterRoutes(r, setup.RouteOptions{SetupTokenHash: tokenHash, AllowRemoteInit: allowRemote})

	// Serve embedded frontend if available
	if web.HasEmbeddedFrontend() {
		r.Use(web.ServeEmbeddedFrontend())
	}

	if allowRemote {
		log.Println("WARNING: remote setup is enabled; keep the setup token private and finish installation immediately")
	}
	log.Printf("Setup wizard available at http://%s/setup?token=%s", publicAddr, token)
	log.Println("Complete the setup wizard to configure 智算引擎")

	server := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(r, &http2.Server{}),
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start setup server: %v", err)
	}
}

func setupServerAddress() (listenAddr, publicAddr string, allowRemote bool) {
	_, port, err := net.SplitHostPort(config.GetServerAddress())
	if err != nil || strings.TrimSpace(port) == "" {
		port = strconv.Itoa(8080)
	}
	allowRemote = isTruthy(os.Getenv("SETUP_ALLOW_REMOTE"))
	host := strings.TrimSpace(os.Getenv("SETUP_BIND_HOST"))
	if host == "" {
		if allowRemote {
			host = strings.TrimSpace(os.Getenv("SERVER_HOST"))
			if host == "" {
				host = "0.0.0.0"
			}
		} else {
			host = "127.0.0.1"
		}
	}
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		allowRemote = true
	}
	listenAddr = net.JoinHostPort(host, port)
	publicHost := host
	if publicHost == "0.0.0.0" || publicHost == "::" || publicHost == "[::]" {
		publicHost = "127.0.0.1"
	}
	publicAddr = net.JoinHostPort(publicHost, port)
	return listenAddr, publicAddr, allowRemote
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func runMainServer() {
	cfg, err := config.LoadForBootstrap()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := logger.Init(logger.OptionsFromConfig(cfg.Log)); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	if cfg.RunMode == config.RunModeSimple {
		log.Println("⚠️  WARNING: Running in SIMPLE mode - billing and quota checks are DISABLED")
	}

	buildInfo := handler.BuildInfo{
		Version:   Version,
		BuildType: BuildType,
	}

	app, err := initializeApplication(buildInfo)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Cleanup()

	// 启动服务器
	go func() {
		if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on %s", app.Server.Addr)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
