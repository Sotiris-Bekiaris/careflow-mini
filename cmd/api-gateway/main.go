package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// TODO: Initialize OpenTelemetry
	// TODO: Connect to gRPC services (patient-svc, appointment-svc)
	// TODO: Setup REST routes with FHIR-style endpoints

	port := getEnv("PORT", "8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler)
	mux.HandleFunc("/fhir/Patient", patientHandler)
	mux.HandleFunc("/fhir/Appointment", appointmentHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		log.Printf("API Gateway starting on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status":"healthy"}`)
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Check gRPC service connections
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status":"ready"}`)
}

func patientHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Forward to patient-svc via gRPC
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintln(w, `{"message":"Patient API - not implemented yet"}`)
}

func appointmentHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Forward to appointment-svc via gRPC
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintln(w, `{"message":"Appointment API - not implemented yet"}`)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
