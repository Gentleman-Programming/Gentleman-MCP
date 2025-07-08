package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	mcpv1 "github.com/Gentleman-Programming/gentleman-mcp/proto/mcp/v1"
)

var (
	serverAddr  = flag.String("addr", "localhost:50051", "Server address")
	caCertFile  = flag.String("ca-cert", "certs/ca-cert.pem", "Path to CA certificate")
	certFile    = flag.String("cert", "certs/client-cert.pem", "Path to client certificate (for mTLS)")
	keyFile     = flag.String("key", "certs/client-key.pem", "Path to client private key (for mTLS)")
	enableMTLS  = flag.Bool("mtls", false, "Enable mutual TLS")
	useInsecure = flag.Bool("insecure", false, "Connect without TLS")
	tenantID    = flag.String("tenant", "demo-tenant", "Tenant ID")
	agentID     = flag.String("agent", "demo-agent", "Agent ID")
	model       = flag.String("model", "gemma3:4b", "Model to use")
	message     = flag.String("message", "Hello Gemma! How are you today?", "Message to send")
)

func main() {
	flag.Parse()

	// Create connection
	conn, err := createConnection()
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create clients
	handshakeClient := mcpv1.NewHandshakeServiceClient(conn)
	agentClient := mcpv1.NewAgentServiceClient(conn)

	log.Printf("🚀 Gentleman MCP Gateway Client")
	log.Printf("📡 Connecting to: %s", *serverAddr)
	log.Printf("🔐 TLS: %v (mTLS: %v)", !*useInsecure, *enableMTLS)
	log.Printf("")

	// Step 1: Register and get session
	log.Printf("1️⃣ Registering with gateway...")
	sessionID, jwtToken, err := registerSession(handshakeClient)
	if err != nil {
		log.Fatalf("❌ Registration failed: %v", err)
	}
	log.Printf("✅ Registration successful!")
	log.Printf("   📋 Session ID: %s", sessionID)
	log.Printf("   🎫 JWT Token: %s...", jwtToken[:20])
	log.Printf("")

	// Step 2: Authenticate token
	log.Printf("2️⃣ Authenticating token...")
	if err := authenticateToken(handshakeClient, jwtToken); err != nil {
		log.Fatalf("❌ Authentication failed: %v", err)
	}
	log.Printf("✅ Authentication successful!")
	log.Printf("")

	// Step 3: Send chat message
	log.Printf("3️⃣ Sending chat message...")
	log.Printf("   💬 Message: %s", *message)
	response, err := sendChatMessage(agentClient, sessionID, *message)
	if err != nil {
		log.Fatalf("❌ Chat failed: %v", err)
	}
	log.Printf("✅ Response received!")
	log.Printf("   🤖 Gemma: %s", response)
	log.Printf("")

	log.Printf("🎉 Demo completed successfully!")
}

// createConnection establishes a gRPC connection with appropriate credentials
func createConnection() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if *useInsecure {
		// Insecure connection
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		log.Printf("⚠️  Using insecure connection")
	} else {
		// TLS connection
		creds, err := loadTLSCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	// Create connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	return conn, nil
}

// loadTLSCredentials loads TLS credentials for the client
func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load CA certificate
	caCert, err := ioutil.ReadFile(*caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	config := &tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS13, // Force TLS 1.3
	}

	if *enableMTLS {
		// Load client certificate for mTLS
		clientCert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		config.Certificates = []tls.Certificate{clientCert}
	}

	return credentials.NewTLS(config), nil
}

// registerSession registers with the gateway and returns session info
func registerSession(client mcpv1.HandshakeServiceClient) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &mcpv1.RegisterRequest{
		TenantId: *tenantID,
		AgentId:  *agentID,
		Model:    *model,
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		return "", "", err
	}

	return resp.SessionId, resp.JwtToken, nil
}

// authenticateToken validates the JWT token
func authenticateToken(client mcpv1.HandshakeServiceClient, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &mcpv1.AuthRequest{
		JwtToken: token,
	}

	resp, err := client.Authenticate(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Valid {
		return fmt.Errorf("token is invalid or expired")
	}

	log.Printf("   ✅ Token valid for tenant: %s, agent: %s", resp.TenantId, resp.AgentId)
	return nil
}

// sendChatMessage sends a chat message and returns the response
func sendChatMessage(client mcpv1.AgentServiceClient, sessionID, message string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &mcpv1.SingleChatRequest{
		SessionId: sessionID,
		Content:   message,
		Model:     *model,
	}

	resp, err := client.SingleChat(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
