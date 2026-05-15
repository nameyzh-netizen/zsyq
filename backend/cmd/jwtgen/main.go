package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/nameyzh-netizen/zsyq/ent/runtime"
	"github.com/nameyzh-netizen/zsyq/internal/config"
	"github.com/nameyzh-netizen/zsyq/internal/repository"
	"github.com/nameyzh-netizen/zsyq/internal/service"
)

func main() {
	email := flag.String("email", "", "Admin email to issue a JWT for (defaults to first active admin)")
	confirm := flag.Bool("confirm-emergency-admin-token", false, "Confirm this emergency operation may mint an administrator JWT")
	flag.Parse()

	if !*confirm && os.Getenv("SUB2API_ENABLE_JWTGEN") != "1" {
		log.Fatal("refusing to mint admin JWT without --confirm-emergency-admin-token or SUB2API_ENABLE_JWTGEN=1")
	}

	cfg, err := config.LoadForBootstrap()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client, sqlDB, err := repository.InitEnt(cfg)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	userRepo := repository.NewUserRepository(client, sqlDB)
	authService := service.NewAuthService(client, userRepo, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user *service.User
	if *email != "" {
		user, err = userRepo.GetByEmail(ctx, *email)
	} else {
		user, err = userRepo.GetFirstAdmin(ctx)
	}
	if err != nil {
		log.Fatalf("failed to resolve admin user: %v", err)
	}
	if !user.IsAdmin() || !user.IsActive() {
		log.Fatalf("refusing to issue token for non-active admin user: %s", user.Email)
	}

	token, err := authService.GenerateToken(user)
	if err != nil {
		log.Fatalf("failed to generate token: %v", err)
	}

	fmt.Printf("ADMIN_EMAIL=%s\nADMIN_USER_ID=%d\nJWT=%s\n", user.Email, user.ID, token)
}
