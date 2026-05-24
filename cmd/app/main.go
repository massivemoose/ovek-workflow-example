package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/massivemoose/ovek-workflow-example/internal/app"
	"github.com/massivemoose/ovek-workflow-example/internal/config"
	"github.com/massivemoose/ovek-workflow-example/internal/pocketbase"
)

func main() {
	cfg := config.LoadApp(os.Getenv)
	pb := pocketbase.NewClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if cfg.CollectionEnsure {
		if err := pb.EnsureRequiredCollections(ctx); err != nil {
			log.Fatalf("ensure PocketBase collections: %v", err)
		}
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app.Routes(pb),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
