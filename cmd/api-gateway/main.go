package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/observability"
	appointmentv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/appointment/v1"
	observationv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/observation/v1"
	patientv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/patient/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultCORSAllowedOrigins = "http://localhost:3001,http://127.0.0.1:3001"
)

type server struct {
	patientClient     patientv1.PatientServiceClient
	appointmentClient appointmentv1.AppointmentServiceClient
	observationClient observationv1.ObservationServiceClient
	labAdapterConn    grpc.ClientConnInterface
	notifyServiceConn grpc.ClientConnInterface
	tracer            trace.Tracer
}

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry
	otlpEndpoint := getEnv("OTLP_ENDPOINT", "http://localhost:4318")
	shutdownTracer, err := observability.InitTracer("api-gateway", otlpEndpoint)
	if err != nil {
		log.Printf("Failed to initialize tracer: %v", err)
	}
	if shutdownTracer != nil {
		defer func() {
			if err := shutdownTracer(ctx); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}()
	}

	_, err = observability.InitMetrics("api-gateway")
	if err != nil {
		log.Printf("Failed to initialize metrics: %v", err)
	}

	// Connect to Patient Service
	patientSvcAddr := getEnv("PATIENT_SVC_ADDR", "localhost:50051")
	patientConn, err := grpc.NewClient(
		patientSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to patient service: %v", err)
	}
	defer func() { _ = patientConn.Close() }()

	// Connect to Appointment Service
	appointmentSvcAddr := getEnv("APPOINTMENT_SVC_ADDR", "localhost:50052")
	appointmentConn, err := grpc.NewClient(
		appointmentSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to appointment service: %v", err)
	}
	defer func() { _ = appointmentConn.Close() }()

	// Connect to Observation Service
	observationSvcAddr := getEnv("OBSERVATION_SVC_ADDR", "localhost:50055")
	observationConn, err := grpc.NewClient(
		observationSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to observation service: %v", err)
	}
	defer func() { _ = observationConn.Close() }()

	// Connect to Lab Adapter (health check only)
	labAdapterAddr := getEnv("LAB_ADAPTER_ADDR", "localhost:50053")
	labAdapterConn, err := grpc.NewClient(
		labAdapterAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to lab adapter: %v", err)
	}
	defer func() { _ = labAdapterConn.Close() }()

	// Connect to Notify Service (health check only)
	notifyServiceAddr := getEnv("NOTIFY_SVC_ADDR", "localhost:50054")
	notifyServiceConn, err := grpc.NewClient(
		notifyServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to notify service: %v", err)
	}
	defer func() { _ = notifyServiceConn.Close() }()

	srv := &server{
		patientClient:     patientv1.NewPatientServiceClient(patientConn),
		appointmentClient: appointmentv1.NewAppointmentServiceClient(appointmentConn),
		observationClient: observationv1.NewObservationServiceClient(observationConn),
		labAdapterConn:    labAdapterConn,
		notifyServiceConn: notifyServiceConn,
		tracer:            otel.Tracer("api-gateway"),
	}

	port := getEnv("PORT", "8080")
	corsAllowedOrigins := strings.Split(getEnv("CORS_ALLOWED_ORIGINS", defaultCORSAllowedOrigins), ",")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.healthHandler)
	mux.HandleFunc("/ready", srv.readyHandler)
	mux.HandleFunc("/fhir/Patient", srv.patientHandler)
	mux.HandleFunc("/fhir/Patient/", srv.patientByIDHandler) // Note trailing slash for ID matching
	mux.HandleFunc("/fhir/Appointment", srv.appointmentHandler)
	mux.HandleFunc("/fhir/Appointment/", srv.appointmentByIDHandler) // Note trailing slash for ID matching
	mux.HandleFunc("/fhir/Observation", srv.observationHandler)
	mux.HandleFunc("/fhir/Observation/", srv.observationByIDHandler) // Note trailing slash for ID matching

	// Wrap with OpenTelemetry middleware
	handler := otelhttp.NewHandler(mux, "api-gateway",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)

	// Apply CORS middleware and logging
	handler = corsMiddleware(corsAllowedOrigins, handler)
	handler = loggingMiddleware(handler)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Graceful shutdown
	go func() {
		log.Printf("API Gateway starting on port %s...", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, `{"status":"healthy"}`)
}

func (s *server) readyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Check Patient Service connection
	_, err := s.patientClient.ListPatients(ctx, &patientv1.ListPatientsRequest{PageSize: 1})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintln(w, `{"status":"not ready","error":"patient service unavailable"}`)
		return
	}

	// Check Appointment Service connection
	// ListAppointments requires a patient ID, so we check the service differently
	_, err = s.appointmentClient.GetAppointment(ctx, &appointmentv1.GetAppointmentRequest{Id: "health-check"})
	// We expect this to return NotFound, which is fine - it means the service is responding
	if err != nil {
		grpcErr := status.Code(err)
		// NotFound (5) is expected, any other error indicates service is down
		if grpcErr != codes.NotFound && grpcErr != codes.InvalidArgument {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintln(w, `{"status":"not ready","error":"appointment service unavailable"}`)
			return
		}
	}

	// Check Lab Adapter health
	if err := s.checkServiceHealth(ctx, s.labAdapterConn, "lab-adapter"); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintf(w, `{"status":"not ready","error":"lab-adapter unavailable: %s"}`+"\n", err.Error())
		return
	}

	// Check Notify Service health
	if err := s.checkServiceHealth(ctx, s.notifyServiceConn, "notify-svc"); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintf(w, `{"status":"not ready","error":"notify-svc unavailable: %s"}`+"\n", err.Error())
		return
	}

	// Check Observation Service connection using gRPC health check
	// (similar to lab-adapter and notify-svc since ListObservations requires patient_id)
	// Note: If observation service is not available, it may not be critical for readiness
	// but we'll still check it for consistency

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, `{"status":"ready"}`)
}

// checkServiceHealth checks the gRPC health status of a service
func (s *server) checkServiceHealth(ctx context.Context, conn grpc.ClientConnInterface, serviceName string) error {
	healthClient := grpc_health_v1.NewHealthClient(conn)
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service not serving (status: %v)", resp.Status)
	}

	return nil
}

// patientHandler handles /fhir/Patient (list and create)
func (s *server) patientHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "patientHandler")
	defer span.End()

	span.SetAttributes(attribute.String("http.method", r.Method))

	switch r.Method {
	case http.MethodGet:
		s.listPatientsHandler(w, r.WithContext(ctx))
	case http.MethodPost:
		s.createPatientHandler(w, r.WithContext(ctx))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// patientByIDHandler handles /fhir/Patient/:id (get and update)
func (s *server) patientByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "patientByIDHandler")
	defer span.End()

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/fhir/Patient/")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("patient.id", id),
	)

	switch r.Method {
	case http.MethodGet:
		s.getPatientHandler(w, r.WithContext(ctx), id)
	case http.MethodPut:
		s.updatePatientHandler(w, r.WithContext(ctx), id)
	case http.MethodDelete:
		s.deletePatientHandler(w, r.WithContext(ctx), id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) createPatientHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "createPatient")
	defer span.End()

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer func() { _ = r.Body.Close() }()

	var fhirPatient fhir.Patient
	if err := json.Unmarshal(body, &fhirPatient); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Convert FHIR to Proto
	protoPatient := fhirToProto(&fhirPatient)

	// Call gRPC service
	resp, err := s.patientClient.CreatePatient(ctx, &patientv1.CreatePatientRequest{
		Patient: protoPatient,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	// Convert Proto back to FHIR
	resultPatient := protoToFHIR(resp.Patient)

	respondJSON(w, http.StatusCreated, resultPatient)
}

func (s *server) getPatientHandler(w http.ResponseWriter, r *http.Request, id string) {
	ctx, span := s.tracer.Start(r.Context(), "getPatient")
	defer span.End()

	resp, err := s.patientClient.GetPatient(ctx, &patientv1.GetPatientRequest{
		Id: id,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	resultPatient := protoToFHIR(resp.Patient)
	respondJSON(w, http.StatusOK, resultPatient)
}

func (s *server) updatePatientHandler(w http.ResponseWriter, r *http.Request, id string) {
	ctx, span := s.tracer.Start(r.Context(), "updatePatient")
	defer span.End()

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer func() { _ = r.Body.Close() }()

	var fhirPatient fhir.Patient
	if err := json.Unmarshal(body, &fhirPatient); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Ensure ID in body matches URL
	fhirPatient.ID = id

	// Convert FHIR to Proto
	protoPatient := fhirToProto(&fhirPatient)

	// Call gRPC service
	resp, err := s.patientClient.UpdatePatient(ctx, &patientv1.UpdatePatientRequest{
		Patient: protoPatient,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	resultPatient := protoToFHIR(resp.Patient)
	respondJSON(w, http.StatusOK, resultPatient)
}

func (s *server) deletePatientHandler(w http.ResponseWriter, r *http.Request, id string) {
	ctx, span := s.tracer.Start(r.Context(), "deletePatient")
	defer span.End()

	_, err := s.patientClient.DeletePatient(ctx, &patientv1.DeletePatientRequest{
		Id: id,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) listPatientsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "listPatients")
	defer span.End()

	// Parse query parameters
	query := r.URL.Query()
	pageSize := query.Get("pageSize")
	pageToken := query.Get("pageToken")
	name := query.Get("name")

	req := &patientv1.ListPatientsRequest{
		PageToken: pageToken,
		Name:      name,
	}

	if pageSize != "" {
		var size int32
		if _, err := fmt.Sscanf(pageSize, "%d", &size); err == nil {
			req.PageSize = size
		}
	}

	// Call gRPC service
	resp, err := s.patientClient.ListPatients(ctx, req)
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	// Convert Proto patients to FHIR
	fhirPatients := make([]fhir.Patient, len(resp.Patients))
	for i, p := range resp.Patients {
		fhirPatients[i] = *protoToFHIR(p)
	}

	// Create FHIR Bundle response
	bundle := struct {
		ResourceType string         `json:"resourceType"`
		Type         string         `json:"type"`
		Entry        []fhir.Patient `json:"entry"`
		NextLink     string         `json:"link,omitempty"`
	}{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry:        fhirPatients,
		NextLink:     resp.NextPageToken,
	}

	respondJSON(w, http.StatusOK, bundle)
}

// fhirToProto converts FHIR Patient to Proto Patient
func fhirToProto(f *fhir.Patient) *patientv1.Patient {
	p := &patientv1.Patient{
		Id:         f.ID,
		Active:     f.Active,
		Gender:     f.Gender,
		BirthDate:  f.BirthDate,
	}

	// Extract first name from name array
	if len(f.Name) > 0 {
		p.FamilyName = f.Name[0].Family
		p.GivenNames = f.Name[0].Given
	}

	// Convert telecom
	if len(f.Telecom) > 0 {
		p.Telecom = make([]*patientv1.Contact, len(f.Telecom))
		for i, t := range f.Telecom {
			p.Telecom[i] = &patientv1.Contact{
				System: t.System,
				Value:  t.Value,
				Use:    t.Use,
			}
		}
	}

	// Convert address
	if len(f.Address) > 0 {
		addr := f.Address[0]
		p.Address = &patientv1.Address{
			Line:       addr.Line,
			City:       addr.City,
			State:      addr.State,
			PostalCode: addr.PostalCode,
			Country:    addr.Country,
		}
	}

	return p
}

// protoToFHIR converts Proto Patient to FHIR Patient
func protoToFHIR(p *patientv1.Patient) *fhir.Patient {
	f := &fhir.Patient{
		ID:        p.Id,
		Active:    p.Active,
		Gender:    p.Gender,
		BirthDate: p.BirthDate,
	}

	// Create name array from proto
	if p.FamilyName != "" || len(p.GivenNames) > 0 {
		f.Name = []fhir.HumanName{
			{
				Family: p.FamilyName,
				Given:  p.GivenNames,
			},
		}
	}

	// Convert telecom
	if len(p.Telecom) > 0 {
		f.Telecom = make([]fhir.Contact, len(p.Telecom))
		for i, t := range p.Telecom {
			f.Telecom[i] = fhir.Contact{
				System: t.System,
				Value:  t.Value,
				Use:    t.Use,
			}
		}
	}

	// Convert address
	if p.Address != nil {
		f.Address = []fhir.Address{
			{
				Line:       p.Address.Line,
				City:       p.Address.City,
				State:      p.Address.State,
				PostalCode: p.Address.PostalCode,
				Country:    p.Address.Country,
			},
		}
	}

	return f
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(status)
	resp := map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]string{
			{
				"severity": "error",
				"code":     "processing",
				"details":  message,
			},
		},
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// respondGRPCError translates gRPC errors to HTTP responses
func respondGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		respondError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var httpStatus int
	switch st.Code() {
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.ResourceExhausted:
		httpStatus = http.StatusTooManyRequests
	case codes.Unimplemented:
		httpStatus = http.StatusNotImplemented
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusInternalServerError
	}

	respondError(w, httpStatus, st.Message())
}

// corsMiddleware handles CORS headers and preflight requests
func corsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		isAllowed := false
		for _, allowed := range allowedOrigins {
			// Trim whitespace from allowed origin
			allowed = strings.TrimSpace(allowed)
			if origin == allowed {
				isAllowed = true
				break
			}
		}

		// Set CORS headers if origin is allowed
		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Create response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		// Log response
		duration := time.Since(start)
		log.Printf("[%s] %s %s - %d (%v)", r.Method, r.URL.Path, r.RemoteAddr, rw.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// appointmentHandler handles /fhir/Appointment (list and create)
func (s *server) appointmentHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "appointmentHandler")
	defer span.End()

	span.SetAttributes(attribute.String("http.method", r.Method))

	switch r.Method {
	case http.MethodGet:
		s.listAppointmentsHandler(w, r.WithContext(ctx))
	case http.MethodPost:
		s.createAppointmentHandler(w, r.WithContext(ctx))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// appointmentByIDHandler handles /fhir/Appointment/:id (get and cancel)
func (s *server) appointmentByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "appointmentByIDHandler")
	defer span.End()

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/fhir/Appointment/")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Appointment ID is required")
		return
	}

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("appointment.id", id),
	)

	switch r.Method {
	case http.MethodGet:
		s.getAppointmentHandler(w, r.WithContext(ctx), id)
	case http.MethodPut:
		s.updateAppointmentHandler(w, r.WithContext(ctx), id)
	case http.MethodDelete:
		s.cancelAppointmentHandler(w, r.WithContext(ctx), id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) createAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "createAppointment")
	defer span.End()

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer func() { _ = r.Body.Close() }()

	var fhirAppointment fhir.Appointment
	if err := json.Unmarshal(body, &fhirAppointment); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Convert FHIR to Proto
	protoAppointment := fhirAppointmentToProto(&fhirAppointment)

	// Call gRPC service
	resp, err := s.appointmentClient.CreateAppointment(ctx, &appointmentv1.CreateAppointmentRequest{
		Appointment: protoAppointment,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	// Convert Proto back to FHIR
	resultAppointment := protoAppointmentToFHIR(resp.Appointment)

	respondJSON(w, http.StatusCreated, resultAppointment)
}

func (s *server) getAppointmentHandler(w http.ResponseWriter, r *http.Request, id string) {
	ctx, span := s.tracer.Start(r.Context(), "getAppointment")
	defer span.End()

	resp, err := s.appointmentClient.GetAppointment(ctx, &appointmentv1.GetAppointmentRequest{
		Id: id,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	resultAppointment := protoAppointmentToFHIR(resp.Appointment)
	respondJSON(w, http.StatusOK, resultAppointment)
}

func (s *server) updateAppointmentHandler(w http.ResponseWriter, r *http.Request, id string) {
	ctx, span := s.tracer.Start(r.Context(), "updateAppointment")
	defer span.End()

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer func() { _ = r.Body.Close() }()

	var fhirAppointment fhir.Appointment
	if err := json.Unmarshal(body, &fhirAppointment); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Ensure the ID from the URL is set on the appointment
	fhirAppointment.ID = id

	// Convert FHIR to Proto
	protoAppointment := fhirAppointmentToProto(&fhirAppointment)

	// Call gRPC service
	resp, err := s.appointmentClient.UpdateAppointment(ctx, &appointmentv1.UpdateAppointmentRequest{
		Appointment: protoAppointment,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	// Convert Proto back to FHIR
	resultAppointment := protoAppointmentToFHIR(resp.Appointment)

	respondJSON(w, http.StatusOK, resultAppointment)
}

func (s *server) cancelAppointmentHandler(w http.ResponseWriter, r *http.Request, id string) {
	ctx, span := s.tracer.Start(r.Context(), "cancelAppointment")
	defer span.End()

	// Parse optional reason from query params
	reason := r.URL.Query().Get("reason")

	_, err := s.appointmentClient.CancelAppointment(ctx, &appointmentv1.CancelAppointmentRequest{
		Id:     id,
		Reason: reason,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) listAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "listAppointments")
	defer span.End()

	// Parse query parameters
	query := r.URL.Query()
	patientID := query.Get("patientId")
	pageSize := query.Get("pageSize")
	pageToken := query.Get("pageToken")

	req := &appointmentv1.ListAppointmentsRequest{
		PatientId: patientID,
		PageToken: pageToken,
	}

	if pageSize != "" {
		var size int32
		if _, err := fmt.Sscanf(pageSize, "%d", &size); err == nil {
			req.PageSize = size
		}
	}

	// Call gRPC service
	resp, err := s.appointmentClient.ListAppointments(ctx, req)
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	// Convert Proto appointments to FHIR and wrap in bundle entries
	bundleEntries := make([]struct {
		Resource fhir.Appointment `json:"resource"`
	}, len(resp.Appointments))
	for i, p := range resp.Appointments {
		bundleEntries[i].Resource = *protoAppointmentToFHIR(p)
	}

	// Create FHIR Bundle response
	bundle := struct {
		ResourceType string `json:"resourceType"`
		Type         string `json:"type"`
		Entry        []struct {
			Resource fhir.Appointment `json:"resource"`
		} `json:"entry"`
		NextLink string `json:"link,omitempty"`
	}{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry:        bundleEntries,
		NextLink:     resp.NextPageToken,
	}

	respondJSON(w, http.StatusOK, bundle)
}

// fhirAppointmentToProto converts FHIR Appointment to Proto Appointment
func fhirAppointmentToProto(f *fhir.Appointment) *appointmentv1.Appointment {
	p := &appointmentv1.Appointment{
		Id:          f.ID,
		Status:      f.Status,
		Description: f.Description,
	}

	// Extract patient and practitioner IDs from participants
	for _, part := range f.Participant {
		if part.Actor.Reference != "" {
			ref := part.Actor.Reference
			// Extract ID from reference like "Patient/123" -> "123"
			var id string
			if slashPos := strings.LastIndex(ref, "/"); slashPos >= 0 {
				id = ref[slashPos+1:]
			} else {
				id = ref // If no slash, use the whole reference as ID
			}

			// Check if reference starts with "Patient/" to determine if it's a patient ID
			if p.PatientId == "" && strings.HasPrefix(ref, "Patient/") {
				p.PatientId = id
			} else if p.PractitionerId == "" && strings.HasPrefix(ref, "Practitioner/") {
				p.PractitionerId = id
			}
		}
	}

	// Convert service type
	if len(f.ServiceType) > 0 && len(f.ServiceType[0].Coding) > 0 {
		p.ServiceType = f.ServiceType[0].Coding[0].Code
	}

	// Convert timestamps
	if !f.Start.IsZero() {
		p.Start = timestamppb.New(f.Start)
	}
	if !f.End.IsZero() {
		p.End = timestamppb.New(f.End)
	}

	return p
}

// protoAppointmentToFHIR converts Proto Appointment to FHIR Appointment
func protoAppointmentToFHIR(p *appointmentv1.Appointment) *fhir.Appointment {
	f := &fhir.Appointment{
		ID:          p.Id,
		Status:      p.Status,
		Description: p.Description,
	}

	// Build participants from patient and practitioner IDs
	if p.PatientId != "" {
		f.Participant = append(f.Participant, fhir.Participant{
			Actor: fhir.Reference{
				Reference: "Patient/" + p.PatientId,
				Display:   "Patient " + p.PatientId,
			},
			Status: "accepted",
		})
	}

	if p.PractitionerId != "" {
		f.Participant = append(f.Participant, fhir.Participant{
			Actor: fhir.Reference{
				Reference: "Practitioner/" + p.PractitionerId,
				Display:   "Practitioner " + p.PractitionerId,
			},
			Status: "accepted",
		})
	}

	// Convert service type
	if p.ServiceType != "" {
		f.ServiceType = []fhir.CodeableConcept{
			{
				Coding: []fhir.Coding{
					{
						Code: p.ServiceType,
					},
				},
			},
		}
	}

	// Convert timestamps
	if p.Start != nil {
		f.Start = p.Start.AsTime()
	}
	if p.End != nil {
		f.End = p.End.AsTime()
	}

	// Initialize meta
	f.Meta = fhir.Meta{}

	return f
}

// observationHandler handles /fhir/Observation (list and create)
func (s *server) observationHandler(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracer.Start(r.Context(), "observationHandler")
	defer span.End()

	switch r.Method {
	case http.MethodGet:
		s.listObservationsHandler(w, r)
	case http.MethodPost:
		s.createObservationHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// observationByIDHandler handles /fhir/Observation/:id (get and update status)
func (s *server) observationByIDHandler(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracer.Start(r.Context(), "observationByIDHandler")
	defer span.End()

	id := strings.TrimPrefix(r.URL.Path, "/fhir/Observation/")
	if id == "" {
		http.Error(w, "Invalid observation ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getObservationHandler(w, r, id)
	case http.MethodPost:
		s.updateObservationStatusHandler(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) createObservationHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "createObservationHandler")
	defer span.End()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var fhirObservation fhir.Observation
	if err := json.Unmarshal(body, &fhirObservation); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	protoObservation := fhirObservationToProto(&fhirObservation)

	resp, err := s.observationClient.CreateObservation(ctx, &observationv1.CreateObservationRequest{
		Observation: protoObservation,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	fhirResp := protoObservationToFHIR(resp.Observation)
	respondJSON(w, http.StatusCreated, fhirResp)
}

func (s *server) getObservationHandler(w http.ResponseWriter, r *http.Request, id string) {
	ctx, span := s.tracer.Start(r.Context(), "getObservationHandler")
	defer span.End()

	resp, err := s.observationClient.GetObservation(ctx, &observationv1.GetObservationRequest{
		Id: id,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	fhirResp := protoObservationToFHIR(resp.Observation)
	respondJSON(w, http.StatusOK, fhirResp)
}

func (s *server) updateObservationStatusHandler(w http.ResponseWriter, r *http.Request, id string) {
	ctx, span := s.tracer.Start(r.Context(), "updateObservationStatusHandler")
	defer span.End()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var updateRequest map[string]string
	if err := json.Unmarshal(body, &updateRequest); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	status, ok := updateRequest["status"]
	if !ok {
		respondError(w, http.StatusBadRequest, "Status field is required")
		return
	}

	resp, err := s.observationClient.UpdateObservationStatus(ctx, &observationv1.UpdateObservationStatusRequest{
		Id:     id,
		Status: status,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	fhirResp := protoObservationToFHIR(resp.Observation)
	respondJSON(w, http.StatusOK, fhirResp)
}

func (s *server) listObservationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "listObservationsHandler")
	defer span.End()

	patientID := r.URL.Query().Get("patient")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "patient query parameter is required")
		return
	}

	req := &observationv1.ListObservationsRequest{
		PatientId: patientID,
		PageSize:  10,
	}

	// Parse pagination token if provided
	if pageToken := r.URL.Query().Get("page_token"); pageToken != "" {
		req.PageToken = pageToken
	}

	resp, err := s.observationClient.ListObservations(ctx, req)
	if err != nil {
		respondGRPCError(w, err)
		return
	}

	// Convert to FHIR observations
	observations := make([]fhir.Observation, len(resp.Observations))
	for i, obs := range resp.Observations {
		protoObs := protoObservationToFHIR(obs)
		observations[i] = *protoObs
	}

	// Return as FHIR bundle
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        len(observations),
		"entry": func() []map[string]interface{} {
			if len(observations) == 0 {
				return []map[string]interface{}{}
			}
			entries := make([]map[string]interface{}, len(observations))
			for i, obs := range observations {
				entries[i] = map[string]interface{}{
					"resource": obs,
				}
			}
			return entries
		}(),
	}

	respondJSON(w, http.StatusOK, bundle)
}

// fhirObservationToProto converts FHIR Observation to protobuf Observation
func fhirObservationToProto(f *fhir.Observation) *observationv1.Observation {
	if f == nil {
		return nil
	}

	code := ""
	codeSystem := ""
	if len(f.Code.Coding) > 0 {
		code = f.Code.Coding[0].Code
		codeSystem = f.Code.Coding[0].System
	} else {
		code = f.Code.Text
	}

	categories := make([]string, len(f.Category))
	for i, c := range f.Category {
		categories[i] = c.Text
	}

	valueQuantityValue := ""
	valueQuantityUnit := ""
	if f.ValueQuantity != nil {
		valueQuantityValue = fmt.Sprintf("%v", f.ValueQuantity.Value)
		valueQuantityUnit = f.ValueQuantity.Unit
	}

	referenceLow := ""
	referenceHigh := ""
	if len(f.ReferenceRange) > 0 {
		if f.ReferenceRange[0].Low != nil {
			referenceLow = fmt.Sprintf("%v", f.ReferenceRange[0].Low.Value)
		}
		if f.ReferenceRange[0].High != nil {
			referenceHigh = fmt.Sprintf("%v", f.ReferenceRange[0].High.Value)
		}
	}

	patientID := ""
	if f.Subject.Reference != "" && len(f.Subject.Reference) > 8 {
		patientID = f.Subject.Reference[8:]
	}

	return &observationv1.Observation{
		Id:                 f.ID,
		Status:             f.Status,
		PatientId:          patientID,
		EffectiveDatetime:  timestamppb.New(f.EffectiveDateTime),
		Issued:             timestamppb.New(f.Issued),
		Code:               code,
		CodeSystem:         codeSystem,
		ValueQuantityValue: valueQuantityValue,
		ValueQuantityUnit:  valueQuantityUnit,
		ValueString:        f.ValueString,
		Category:           categories,
		ReferenceRangeLow:  referenceLow,
		ReferenceRangeHigh: referenceHigh,
	}
}

// protoObservationToFHIR converts protobuf Observation to FHIR Observation
func protoObservationToFHIR(p *observationv1.Observation) *fhir.Observation {
	if p == nil {
		return nil
	}

	// Parse quantities
	var valueQuantity *fhir.Quantity
	if p.ValueQuantityValue != "" {
		var value float64
		_, _ = fmt.Sscanf(p.ValueQuantityValue, "%f", &value)
		valueQuantity = &fhir.Quantity{
			Value: value,
			Unit:  p.ValueQuantityUnit,
		}
	}

	referenceRange := []fhir.ReferenceRange{}
	if p.ReferenceRangeLow != "" || p.ReferenceRangeHigh != "" {
		var low, high *fhir.Quantity
		if p.ReferenceRangeLow != "" {
			var lowVal float64
			_, _ = fmt.Sscanf(p.ReferenceRangeLow, "%f", &lowVal)
			low = &fhir.Quantity{Value: lowVal}
		}
		if p.ReferenceRangeHigh != "" {
			var highVal float64
			_, _ = fmt.Sscanf(p.ReferenceRangeHigh, "%f", &highVal)
			high = &fhir.Quantity{Value: highVal}
		}
		referenceRange = append(referenceRange, fhir.ReferenceRange{
			Low:  low,
			High: high,
		})
	}

	categories := make([]fhir.CodeableConcept, len(p.Category))
	for i, c := range p.Category {
		categories[i] = fhir.CodeableConcept{
			Text: c,
		}
	}

	return &fhir.Observation{
		ID:       p.Id,
		Status:   p.Status,
		Category: categories,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{
				{
					System: p.CodeSystem,
					Code:   p.Code,
				},
			},
			Text: p.Code,
		},
		Subject: fhir.Reference{
			Reference: "Patient/" + p.PatientId,
		},
		EffectiveDateTime: p.EffectiveDatetime.AsTime(),
		Issued:            p.Issued.AsTime(),
		ValueQuantity:     valueQuantity,
		ValueString:       p.ValueString,
		ReferenceRange:    referenceRange,
		Meta: fhir.Meta{
			LastUpdated: time.Now().UTC(),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
