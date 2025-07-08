package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	agentID     = flag.String("agent", "stream-agent", "Agent ID")
	model       = flag.String("model", "gemma3:4b", "Model to use")
)

func main() {
	flag.Parse()

	log.Printf("🚀 Gentleman MCP Gateway - Streaming Client")
	log.Printf("📡 Connecting to: %s", *serverAddr)
	log.Printf("🔐 TLS: %v (mTLS: %v)", !*useInsecure, *enableMTLS)
	log.Printf("")

	// Create connection
	conn, err := createConnection()
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create clients
	handshakeClient := mcpv1.NewHandshakeServiceClient(conn)
	agentClient := mcpv1.NewAgentServiceClient(conn)

	// Step 1: Register and get session
	log.Printf("1️⃣ Registering with gateway...")
	sessionID, jwtToken, err := registerSession(handshakeClient)
	if err != nil {
		log.Fatalf("❌ Registration failed: %v", err)
	}
	log.Printf("✅ Registration successful!")
	log.Printf("   📋 Session ID: %s", sessionID)
	log.Printf("   🎫 JWT Token: %s...", jwtToken[:min(20, len(jwtToken))])
	log.Printf("")

	// Step 2: Start streaming chat
	log.Printf("2️⃣ Starting streaming chat session...")
	log.Printf("   💡 Type messages and press Enter")
	log.Printf("   💡 Type 'quit' or 'exit' to end session")
	log.Printf("   💡 Type 'help' for available commands")
	log.Printf("")

	err = startStreamingChat(agentClient, sessionID)
	if err != nil {
		log.Fatalf("❌ Streaming chat failed: %v", err)
	}

	log.Printf("🎉 Streaming session completed!")
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

// startStreamingChat starts a bidirectional streaming chat session
func startStreamingChat(client mcpv1.AgentServiceClient, sessionID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create streaming connection
	stream, err := client.Chat(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Channel to signal when to close
	done := make(chan struct{})

	// Goroutine to handle incoming messages from server
	go func() {
		defer close(done)
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Printf("📡 Server closed the stream")
				return
			}
			if err != nil {
				log.Printf("❌ Stream receive error: %v", err)
				return
			}

			// Display received message
			timestamp := ""
			if msg.Timestamp != nil {
				timestamp = msg.Timestamp.AsTime().Format("15:04:05")
			}

			switch msg.Type {
			case mcpv1.MessageType_MESSAGE_TYPE_ASSISTANT:
				log.Printf("🤖 [%s] Gemma: %s", timestamp, msg.Content)
			case mcpv1.MessageType_MESSAGE_TYPE_SYSTEM:
				log.Printf("🔔 [%s] System: %s", timestamp, msg.Content)
			default:
				log.Printf("📨 [%s] %s: %s", timestamp, msg.Type, msg.Content)
			}
			log.Printf("")
			fmt.Print("💬 You: ")
		}
	}()

	// Read user input and send messages
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("💬 You: ")

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())

		// Handle special commands
		switch strings.ToLower(input) {
		case "quit", "exit", "q":
			log.Printf("👋 Ending session...")
			stream.CloseSend()
			<-done
			return nil

		case "help", "h":
			printHelp()
			fmt.Print("💬 You: ")
			continue

		case "clear", "cls":
			clearScreen()
			fmt.Print("💬 You: ")
			continue

		case "status":
			log.Printf("📊 Session ID: %s", sessionID)
			log.Printf("📊 Model: %s", *model)
			log.Printf("📊 Connected to: %s", *serverAddr)
			fmt.Print("💬 You: ")
			continue

		case "":
			// Empty input, just continue
			fmt.Print("💬 You: ")
			continue
		}

		// Create and send message
		msg := &mcpv1.ChatMessage{
			MessageId: generateMessageID(),
			SessionId: sessionID,
			Content:   input,
			Type:      mcpv1.MessageType_MESSAGE_TYPE_USER,
			Timestamp: timestamppb.New(time.Now()),
		}

		err := stream.Send(msg)
		if err != nil {
			log.Printf("❌ Failed to send message: %v", err)
			break
		}

		// Don't print "You: " immediately, let the response handler do it
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Input error: %v", err)
	}

	// Close the send direction and wait for completion
	stream.CloseSend()
	<-done

	return nil
}

// generateMessageID creates a simple message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// printHelp displays available commands
func printHelp() {
	log.Printf("📚 Available commands:")
	log.Printf("   help, h     - Show this help")
	log.Printf("   quit, exit  - End the session")
	log.Printf("   clear, cls  - Clear screen")
	log.Printf("   status      - Show session status")
	log.Printf("   [message]   - Send message to Gemma")
}

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print("\033[2J\033[1;1H")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
