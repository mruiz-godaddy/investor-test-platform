package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"backend/bidding"
	"backend/config"
	"backend/db"
	"backend/handler"
	"backend/lifecycle"
	"backend/scenario"
	"backend/store"
)

func main() {
	port := flag.Int("port", 8080, "HTTP port")
	dbPath := flag.String("db", "biddings.db", "SQLite database path")
	autoFinalize := flag.Bool("auto-finalize", true, "Auto-finalize expired listings")
	seed := flag.Bool("seed", true, "Seed default shoppers")
	upstream := flag.String("upstream", "", "Upstream API host for reverse proxy (e.g. https://api.test-godaddy.com)")
	findUpstream := flag.String("find-upstream", "", "Upstream Find API host for search proxy (e.g. https://entourage.prod.aws.godaddy.com)")
	flag.Parse()

	// Init database
	database := db.New(*dbPath)
	defer database.Close()

	// Init store
	s := store.New(database)

	// Seed
	if *seed {
		s.SeedDefaults()
	}

	// Config
	cfg := config.New()
	cfg.SetAutoFinalize(*autoFinalize)

	// Bidding engine
	eng := bidding.NewEngine(s)

	// Lifecycle manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lm := lifecycle.NewManager(s, cfg)
	go lm.Run(ctx)

	// Scenario loader
	sc := scenario.NewLoader(s, cfg, eng)

	r := mux.NewRouter()

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if req.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, req)
		})
	})

	// Request logging middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			log.Printf("%s %s", req.Method, req.URL.Path)
			next.ServeHTTP(w, req)
			_ = start
		})
	})

	// App-facing endpoints
	appH := handler.NewAppHandler(s, cfg, eng, *upstream, *findUpstream)
	r.HandleFunc("/v1/aftermarket/domains/listings/{listingId}", appH.GetListing).Methods("GET")
	r.HandleFunc("/v1/aftermarket/domains/listings/{listingId}/bids", appH.PlaceBid).Methods("POST")
	r.HandleFunc("/v1/aftermarket/domains/bidding", appH.GetBiddingListings).Methods("GET")
	r.HandleFunc("/v1/aftermarket/domains/won", appH.GetWonListings).Methods("GET")
	r.HandleFunc("/v1/aftermarket/domains/didNotWin", appH.GetLostListings).Methods("GET")
	r.HandleFunc("/v4/aftermarket/find/auction/recommend", appH.SearchListings).Methods("GET")
	r.HandleFunc("/v1/aftermarket/domains/member/authorized", appH.GetMemberAuthorized).Methods("GET")

	// Admin endpoints
	adminH := handler.NewAdminHandler(s, cfg, eng, sc)
	r.HandleFunc("/admin/listings", adminH.CreateListing).Methods("POST")
	r.HandleFunc("/admin/listings", adminH.ListListings).Methods("GET")
	r.HandleFunc("/admin/listings/{id}", adminH.GetListing).Methods("GET")
	r.HandleFunc("/admin/listings/{id}/status", adminH.UpdateStatus).Methods("PUT")
	r.HandleFunc("/admin/listings/{id}/endtime", adminH.UpdateEndTime).Methods("PUT")
	r.HandleFunc("/admin/listings/{id}/sniper-bid", adminH.SniperBid).Methods("POST")
	r.HandleFunc("/admin/shoppers", adminH.CreateShopper).Methods("POST")
	r.HandleFunc("/admin/shoppers", adminH.ListShoppers).Methods("GET")
	r.HandleFunc("/admin/shoppers/{id}", adminH.GetShopper).Methods("GET")
	r.HandleFunc("/admin/reset", adminH.Reset).Methods("POST")
	r.HandleFunc("/admin/wipe", adminH.WipeDB).Methods("POST")
	r.HandleFunc("/admin/export", adminH.ExportDB).Methods("GET")
	r.HandleFunc("/admin/import", adminH.ImportDB).Methods("POST")
	r.HandleFunc("/admin/setup", adminH.SetupSystem).Methods("POST")
	r.HandleFunc("/admin/scenarios/{name}", adminH.LoadScenario).Methods("POST")
	r.HandleFunc("/admin/config", adminH.UpdateConfig).Methods("PUT")
	r.HandleFunc("/admin/config", adminH.GetConfig).Methods("GET")
	r.HandleFunc("/admin/time", adminH.UpdateTime).Methods("PUT")
	r.HandleFunc("/admin/time", adminH.GetTime).Methods("GET")

	// Reverse proxy catch-all (registered LAST)
	if *upstream != "" {
		upstreamURL, err := url.Parse(*upstream)
		if err != nil {
			log.Fatalf("Invalid upstream URL: %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			req.Host = upstreamURL.Host
			proxy.ServeHTTP(w, req)
			log.Printf("PROXY %s %s → %s (%dms)",
				req.Method, req.URL.Path, upstreamURL.Host, time.Since(start).Milliseconds())
		})
		log.Printf("Reverse proxy enabled → %s", *upstream)
	} else {
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error":"no upstream configured","path":%q}`, req.URL.Path)
		})
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: r,
	}

	go func() {
		log.Printf("Mock auction server starting on :%d", *port)
		if *upstream != "" {
			log.Printf("Reverse proxy: unmatched routes → %s", *upstream)
		} else {
			log.Printf("No upstream configured: unmatched routes → 404")
		}
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down...")
	cancel() // Stop lifecycle manager
	srv.Shutdown(context.Background())
	log.Println("Server stopped")
}
