package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/sairaviteja27/nova-infra-task/config"
	"github.com/sairaviteja27/nova-infra-task/utils"
	"github.com/sairaviteja27/nova-infra-task/wallet"
)

type Server struct {
	addr       string
	httpServer *http.Server
	svc        *wallet.Service
}

func NewServer(addr string) *Server {
	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := connectMongo(context.Background(), cfg.MongoURI, cfg.MongoDBName)
	if err != nil {
		log.Fatalf("mongo setup error: %v", err)
	}

	client := wallet.NewClient(cfg.RPCEndpoint)
	svc := wallet.NewService(cfg.CacheTTL, client.Fetch)

	mux := http.NewServeMux()
	s := &Server{
		addr: addr,
		httpServer: &http.Server{
			Addr: addr,
		},
		svc: svc,
	}

	mux.Handle("/api/get-balance",
		APIKeyAuth(db)(http.HandlerFunc(s.walletsHandler)),
	)
	gzWrapped := gziphandler.GzipHandler(mux)
	limiter := utils.NewIPRateLimiter(cfg.RateLimitPerMin, time.Minute)
	s.httpServer.Handler = limiter.Middleware(gzWrapped)

	return s
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

type requestBody struct {
	Wallets []string `json:"wallets"`
}

func (s *Server) walletsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req requestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	results, errs := s.svc.FetchMany(r.Context(), req.Wallets)

	for addr, err := range errs {
		log.Printf("addr %s error: %v", addr, err)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}
