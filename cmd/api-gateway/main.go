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
	patientv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/patient/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	patientClient     patientv1.PatientServiceClient
	appointmentClient appointmentv1.AppointmentServiceClient
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

	srv := &server{
		patientClient:     patientv1.NewPatientServiceClient(patientConn),
		appointmentClient: appointmentv1.NewAppointmentServiceClient(appointmentConn),
		tracer:            otel.Tracer("api-gateway"),
	}

	port := getEnv("PORT", "8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.healthHandler)
	mux.HandleFunc("/ready", srv.readyHandler)
	mux.HandleFunc("/fhir/Patient", srv.patientHandler)
	mux.HandleFunc("/fhir/Patient/", srv.patientByIDHandler) // Note trailing slash for ID matching
	mux.HandleFunc("/fhir/Appointment", srv.appointmentHandler)
	mux.HandleFunc("/fhir/Appointment/", srv.appointmentByIDHandler) // Note trailing slash for ID matching

	// Wrap with OpenTelemetry middleware
	handler := otelhttp.NewHandler(mux, "api-gateway",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: loggingMiddleware(handler),
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
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check Patient Service connection
	_, err := s.patientClient.ListPatients(ctx, &patientv1.ListPatientsRequest{PageSize: 1})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintln(w, `{"status":"not ready","error":"patient service unavailable"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, `{"status":"ready"}`)
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

	req := &patientv1.ListPatientsRequest{
		PageToken: pageToken,
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
		FamilyName: f.Name.Family,
		GivenNames: f.Name.Given,
		Gender:     f.Gender,
		BirthDate:  f.BirthDate,
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
		Name: fhir.HumanName{
			Family: p.FamilyName,
			Given:  p.GivenNames,
		},
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

	// Convert Proto appointments to FHIR
	fhirAppointments := make([]fhir.Appointment, len(resp.Appointments))
	for i, p := range resp.Appointments {
		fhirAppointments[i] = *protoAppointmentToFHIR(p)
	}

	// Create FHIR Bundle response
	bundle := struct {
		ResourceType string             `json:"resourceType"`
		Type         string             `json:"type"`
		Entry        []fhir.Appointment `json:"entry"`
		NextLink     string             `json:"link,omitempty"`
	}{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry:        fhirAppointments,
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
			// Naive approach: assume first participant is patient, others are practitioners
			if p.PatientId == "" && part.Actor.Display != "" {
				p.PatientId = part.Actor.Reference
			} else if p.PractitionerId == "" {
				p.PractitionerId = part.Actor.Reference
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
				Reference: p.PatientId,
				Display:   "Patient",
			},
			Status: "accepted",
		})
	}

	if p.PractitionerId != "" {
		f.Participant = append(f.Participant, fhir.Participant{
			Actor: fhir.Reference{
				Reference: p.PractitionerId,
				Display:   "Practitioner",
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
