// Run starts the server
// @title Prim API
// @version 1.0
// @description This is the Prim API
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8081
// @BasePath /api/v1
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/m-mahmoud-alsaid/prim-backend/docs"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/app"
)

func main() {
	app := &app.App{}

	go func() {
		if err := app.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()

	app.Shutdown()
}
