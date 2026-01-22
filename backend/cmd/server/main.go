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
	"fmt"
	"log/slog"

	"piemdm/pkg/config"
	"piemdm/pkg/http"
	"piemdm/pkg/log"
)

func main() {
	conf := config.NewConfig()
	logger := log.NewLog(conf)

	// cron server
	go func() {
		cronServ, cleanup, err := newCronApp(conf, logger)
		if err != nil {
			panic(err)
		}
		cronServ.Start()
		defer cleanup()
	}()

	// webhook server
	go func() {
		webhookServ, cleanup, err := newWebhookApp(conf, logger)
		if err != nil {
			panic(err)
		}
		webhookServ.Start()
		defer cleanup()
	}()

	// web server
	servers, cleanup, err := newApp(conf, logger)
	if err != nil {
		panic(err)
	}

	slog.Info("Server Start.", "host", fmt.Sprintf("0.0.0.0:%d", conf.GetInt("http.port")))

	// servers.
	http.Run(servers.ServerHTTP, fmt.Sprintf("0.0.0.0:%d", conf.GetInt("http.port")))
	defer cleanup()
}
