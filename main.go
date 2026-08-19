package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mine-vent/internal/handler"
)

func main() {
	dbPath := os.Getenv("MINE_VENT_DB")
	if dbPath == "" {
		dbPath = "mine-vent.db"
	}
	addr := os.Getenv("MINE_VENT_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	router, err := handler.NewRouter(dbPath)
	if err != nil {
		log.Fatalf("failed to create router: %v", err)
	}
	defer router.Close()

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		seedData(router)
		log.Printf("mine-vent server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func seedData(router *handler.Router) {
	resp, err := http.Get("http://localhost" + mustGetAddr() + "/api/sensors")
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			return
		}
	}

	baseTime := time.Now()

	for i, name := range []string{"主通风口CH4传感器", "回风巷CO传感器", "掘进面风速传感器"} {
		body := fmt.Sprintf(`{"id":"sensor-00%d","name":"%s","type":"%s","direction":"%s","location":"level1-section%d","area_id":"area-1","unit":"%s","min_threshold":0,"max_threshold":100,"is_active":true}`, i+1, name, []string{"CH4", "CO", "AIRFLOW"}[i], []string{"intake", "exhaust", "intake"}[i], i+1, []string{"ppm", "ppm", "m/s"}[i])
		req, _ := http.NewRequest("POST", "http://localhost"+mustGetAddr()+"/api/sensors", bytes.NewBufferString(body))
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}

	for i := 1; i <= 2; i++ {
		body := fmt.Sprintf(`{"id":"fan-00%d","name":"主风机%d号","area_id":"area-1","capacity":5000,"current_rpm":0,"status":"stopped","power_kw":75,"installed_at":"%s","last_service_at":"%s","is_active":true}`, i, i, baseTime.Add(-30*24*time.Hour).Format(time.RFC3339), baseTime.Add(-30*24*time.Hour).Format(time.RFC3339))
		req, _ := http.NewRequest("POST", "http://localhost"+mustGetAddr()+"/api/fans", bytes.NewBufferString(body))
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}

	log.Println("seed data inserted")
}

func mustGetAddr() string {
	addr := os.Getenv("MINE_VENT_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	if addr[0] == ':' {
		return addr
	}
	return ":" + addr
}
