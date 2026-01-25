// @title PieMDM API
// @version 1.0
// @description PieMDM Master Data Management System API Documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@piemdm.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8787
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"log/slog"

	"piemdm/internal/configs"
	"piemdm/pkg/configloader"
	"piemdm/pkg/http"
	"piemdm/pkg/log"
)

func main() {
	// 1. Load config (Viper)
	v, err := configloader.Load()
	if err != nil {
		panic(err)
	}

	// 2. Unmarshal to struct (Strongly Typed)
	var cfg configs.Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	// 3. Initialize Logger
	// Note: log.NewLog still accepts *viper.Viper for now.
	// Future refactor: Update log.NewLog to accept configs.Config or specific LogConfig
	logger := log.NewLog(v)

	// cron server
	go func() {
		// Pass viper 'v' for backward compatibility
		cronServ, cleanup, err := newCronApp(v, logger)
		if err != nil {
			panic(err)
		}
		cronServ.Start()
		defer cleanup()
	}()

	// webhook server
	go func() {
		webhookServ, cleanup, err := newWebhookApp(v, logger)
		if err != nil {
			panic(err)
		}
		webhookServ.Start()
		defer cleanup()
	}()

	// web server
	servers, cleanup, err := newApp(v, logger)
	if err != nil {
		panic(err)
	}

	// 4. Start Feishu Event Loop (Long Connection)
	if servers.FeishuService != nil {
		go func() {
			slog.Info("Starting Feishu Event Loop...")
			if err := servers.FeishuService.StartEventLoop(context.Background()); err != nil {
				slog.Error("Failed to start Feishu Event Loop", "error", err)
			}

			// 5. 自动同步现有审批定义的订阅状态
			if servers.ApprovalService != nil {
				slog.Info("Syncing Feishu approval subscriptions...")
				if err := servers.ApprovalService.SyncFeishuSubscriptions(context.Background()); err != nil {
					slog.Error("Failed to sync Feishu subscriptions", "error", err)
				}
			}
		}()
	}

	slog.Info("Server Start.", "host", fmt.Sprintf("0.0.0.0:%d", v.GetInt("http.port")))

	// servers.
	http.Run(servers.ServerHTTP, fmt.Sprintf("0.0.0.0:%d", v.GetInt("http.port")))
	defer cleanup()
}
