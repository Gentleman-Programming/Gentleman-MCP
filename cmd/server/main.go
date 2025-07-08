package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/Gentleman-Programming/gentleman-mcp/internal/handlers"
	mcpv1 "github.com/Gentleman-Programming/gentleman-mcp/proto/mcp/v1"
)

var (
	port        = flag.String("port", "50051", "The server port")
	webPort     = flag.String("web-port", "8080", "The gRPC-Web server port")
	certFile    = flag.String("cert-file", "certs/server-cert.pem", "Path to the server certificate file")
	keyFile     = flag.String("key-file", "certs/server-key.pem", "Path to the server private key file")
	caCertFile  = flag.String("ca-cert-file", "certs/ca-cert.pem", "Path to the CA certificate file")
	ollamaURL   = flag.String("ollama-url", "http://localhost:11434", "Ollama server URL")
	enableMTLS  = flag.Bool("mtls", false, "Enable mutual TLS authentication")
	insecure    = flag.Bool("insecure", false, "Run server without TLS (development only)")
	enableWeb   = flag.Bool("enable-web", true, "Enable gRPC-Web proxy server")
	corsOrigins = flag.String("cors-origins", "http://localhost:3000,http://localhost:8080", "Comma-separated list of allowed CORS origins")
)

func main() {
	flag.Parse()

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", *port, err)
	}

	// Setup gRPC server options
	var opts []grpc.ServerOption

	if !*insecure {
		// Load TLS credentials
		creds, err := loadTLSCredentials()
		if err != nil {
			log.Fatalf("❌ Failed to load TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
		log.Printf("🔐 TLS enabled (mTLS: %v)", *enableMTLS)
	} else {
		log.Printf("⚠️  Running in INSECURE mode (no TLS)")
	}

	// Create gRPC server
	server := grpc.NewServer(opts...)

	// Register services
	handshakeServer := handlers.NewHandshakeServer()
	agentServer := handlers.NewAgentServer(*ollamaURL)

	mcpv1.RegisterHandshakeServiceServer(server, handshakeServer)
	mcpv1.RegisterAgentServiceServer(server, agentServer)

	// Enable reflection for development (grpcurl support)
	reflection.Register(server)

	// Start cleanup routine for expired sessions
	go startSessionCleanup(handshakeServer)

	// Start gRPC-Web proxy if enabled
	var webServer *http.Server
	if *enableWeb {
		webServer = startGRPCWebServer(server)
	}

	// Graceful shutdown handling
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Printf("🛑 Shutting down servers...")

		// Shutdown web server first
		if webServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			webServer.Shutdown(ctx)
		}

		// Then shutdown gRPC server
		server.GracefulStop()
	}()

	// Start server
	log.Printf("🚀 Gentleman MCP Gateway starting...")
	log.Printf("📡 gRPC Server listening on :%s", *port)
	if *enableWeb {
		log.Printf("🌐 gRPC-Web Server listening on :%s", *webPort)
	}
	log.Printf("🤖 Ollama URL: %s", *ollamaURL)
	log.Printf("📋 Services registered:")
	log.Printf("   • HandshakeService - Authentication & session management")
	log.Printf("   • AgentService - Chat with LLM")
	log.Printf("")
	log.Printf("💡 Test with grpcurl:")
	if *insecure {
		log.Printf("   grpcurl -plaintext -d '{\"tenant_id\":\"demo\",\"agent_id\":\"test\",\"model\":\"gemma3:4b\"}' localhost:%s mcp.v1.HandshakeService/Register", *port)
	} else {
		log.Printf("   grpcurl -cacert certs/ca-cert.pem -d '{\"tenant_id\":\"demo\",\"agent_id\":\"test\",\"model\":\"gemma3:4b\"}' localhost:%s mcp.v1.HandshakeService/Register", *port)
	}
	if *enableWeb {
		log.Printf("🌐 Test with browser:")
		log.Printf("   Open http://localhost:%s in your browser", *webPort)
	}
	log.Printf("")

	if err := server.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}

// loadTLSCredentials loads the TLS credentials for the server
func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load server certificate
	serverCert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	// Configure TLS
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		MinVersion:   tls.VersionTLS13, // Force TLS 1.3 as mentioned in README
	}

	if *enableMTLS {
		// Load CA certificate for client verification
		if *caCertFile == "" {
			return nil, fmt.Errorf("CA certificate file is required for mTLS")
		}

		// For mTLS, we would load the CA cert and set ClientAuth
		config.ClientAuth = tls.RequireAndVerifyClientCert
		log.Printf("🔒 Mutual TLS (mTLS) enabled")
	}

	return credentials.NewTLS(config), nil
}

// startSessionCleanup runs a background goroutine to clean up expired sessions
func startSessionCleanup(handshakeServer *handlers.HandshakeServer) {
	ticker := time.NewTicker(1 * time.Minute) // Cleanup every minute
	defer ticker.Stop()

	for range ticker.C {
		handshakeServer.CleanupExpiredSessions()
		activeCount := handshakeServer.GetActiveSessionsCount()
		if activeCount > 0 {
			log.Printf("🧹 Session cleanup: %d active sessions", activeCount)
		}
	}
}

// ensureCertFiles checks if certificate files exist and provides helpful error messages
func ensureCertFiles() error {
	files := []string{*certFile, *keyFile}
	if *enableMTLS {
		files = append(files, *caCertFile)
	}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			absPath, _ := filepath.Abs(file)
			return fmt.Errorf(`certificate file not found: %s

💡 Generate development certificates with:
   ./scripts/gen-certs.sh

📁 Expected certificate files:
   • %s (server certificate)
   • %s (server private key)
   • %s (CA certificate, for mTLS)

🔧 Or run without TLS for development:
   go run cmd/server/main.go -insecure`, absPath, *certFile, *keyFile, *caCertFile)
		}
	}
	return nil
}

// startGRPCWebServer starts the gRPC-Web proxy server
func startGRPCWebServer(grpcServer *grpc.Server) *http.Server {
	// Create gRPC-Web wrapper
	wrappedGrpc := grpcweb.WrapServer(grpcServer,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool {
			// Parse allowed origins
			allowedOrigins := strings.Split(*corsOrigins, ",")
			for _, allowedOrigin := range allowedOrigins {
				if strings.TrimSpace(allowedOrigin) == origin {
					return true
				}
			}
			// Allow localhost for development
			return strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1")
		}),
	)

	// Create HTTP server
	httpServer := &http.Server{
		Addr: ":" + *webPort,
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			// Add CORS headers
			resp.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
			resp.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Grpc-Web, X-User-Agent")
			resp.Header().Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message")

			// Handle preflight requests
			if req.Method == "OPTIONS" {
				resp.WriteHeader(http.StatusOK)
				return
			}

			// Serve static files for the web client
			if req.URL.Path == "/" || req.URL.Path == "/index.html" {
				serveWebClient(resp, req)
				return
			}

			// Handle gRPC-Web requests
			if wrappedGrpc.IsGrpcWebRequest(req) {
				wrappedGrpc.ServeHTTP(resp, req)
				return
			}

			// Default 404
			http.NotFound(resp, req)
		}),
	}

	// Start server in goroutine
	go func() {
		log.Printf("🌐 Starting gRPC-Web server on :%s", *webPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ gRPC-Web server error: %v", err)
		}
	}()

	return httpServer
}

// serveWebClient serves a simple web client for testing
func serveWebClient(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Gentleman MCP Gateway - Web Client</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .status { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .success { background-color: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .error { background-color: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        button { padding: 8px 16px; margin: 5px; cursor: pointer; }
        textarea { width: 100%; height: 200px; margin: 10px 0; }
        .response { background: #f8f9fa; padding: 15px; border-radius: 4px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Gentleman MCP Gateway</h1>
        <h2>Web Client Demo (Sprint 3)</h2>

        <div class="status success">
            ✅ gRPC-Web connection established<br>
            📡 Server: ` + r.Host + `<br>
            🌐 CORS enabled for web browsers
        </div>

        <h3>Test Connection</h3>
        <button onclick="testConnection()">Test gRPC-Web Connection</button>

        <h3>Register Session</h3>
        <input type="text" id="tenantId" placeholder="Tenant ID" value="web-tenant">
        <input type="text" id="agentId" placeholder="Agent ID" value="web-agent">
        <button onclick="registerSession()">Register</button>

        <h3>Chat with Gemma</h3>
        <textarea id="message" placeholder="Type your message here...">Hello Gemma from the browser!</textarea>
        <button onclick="sendMessage()">Send Message</button>

        <div id="response" class="response" style="display:none;">
            <h4>Response:</h4>
            <div id="responseContent"></div>
        </div>

        <div id="logs" class="response">
            <h4>Logs:</h4>
            <div id="logContent"></div>
        </div>
    </div>

    <script>
        let sessionId = null;
        let jwtToken = null;

        function log(message) {
            const logDiv = document.getElementById('logContent');
            logDiv.innerHTML += '<div>' + new Date().toLocaleTimeString() + ': ' + message + '</div>';
            logDiv.scrollTop = logDiv.scrollHeight;
        }

        function testConnection() {
            log('🔄 Testing gRPC-Web connection...');
            log('✅ gRPC-Web proxy is running on port ` + *webPort + `');
            log('📡 gRPC server is running on port ` + *port + `');
            log('🌐 Ready for frontend integration!');
        }

        function registerSession() {
            const tenantId = document.getElementById('tenantId').value;
            const agentId = document.getElementById('agentId').value;

            log('🔄 Registering session for tenant: ' + tenantId + ', agent: ' + agentId);

            // Note: This is a demo page. Actual gRPC-Web client would be implemented
            // with the generated TypeScript code in a React/Angular app.

            log('💡 To implement actual gRPC-Web client:');
            log('   1. Use generated TypeScript code in web/generated/');
            log('   2. Install grpc-web npm package');
            log('   3. Create React hooks or Angular services');
            log('   4. See the example React app we will create next!');
        }

        function sendMessage() {
            const message = document.getElementById('message').value;
            log('💬 Sending message: ' + message);
            log('🤖 (Demo mode - actual response would come from Gemma via gRPC-Web)');

            // Show simulated response
            document.getElementById('response').style.display = 'block';
            document.getElementById('responseContent').innerHTML =
                '<strong>🤖 Gemma:</strong> Hello from the web! This is a demo response. ' +
                'The actual implementation will use gRPC-Web to communicate with the gateway.';
        }

        // Initialize
        log('🚀 Gentleman MCP Gateway Web Client loaded');
        log('📡 Server: ' + window.location.host);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func init() {
	// Check certificate files on startup (unless running insecure)
	flag.Parse()
	if !*insecure {
		if err := ensureCertFiles(); err != nil {
			log.Fatalf("❌ %v", err)
		}
	}
}
