package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Gentleman-Programming/gentleman-mcp/internal/ollama"
	mcpv1 "github.com/Gentleman-Programming/gentleman-mcp/proto/mcp/v1"
)

type AgentServer struct {
	mcpv1.UnimplementedAgentServiceServer
	ollamaClient  *ollama.Client
	activeStreams map[string]*StreamSession
	streamsMutex  sync.RWMutex
}

// StreamSession holds information about an active streaming session
type StreamSession struct {
	SessionID    string
	Stream       grpc.BidiStreamingServer[mcpv1.ChatMessage, mcpv1.ChatMessage]
	Model        string
	CreatedAt    time.Time
	LastActivity time.Time
	Context      context.Context
	Cancel       context.CancelFunc
}

func NewAgentServer(ollamaBaseURL string) *AgentServer {
	server := &AgentServer{
		ollamaClient:  ollama.NewClient(ollamaBaseURL),
		activeStreams: make(map[string]*StreamSession),
	}

	// Start cleanup routine for inactive streams
	go server.cleanupInactiveStreams()

	return server
}

func (s *AgentServer) SingleChat(ctx context.Context, req *mcpv1.SingleChatRequest) (*mcpv1.SingleChatResponse, error) {
	// 1. Validar request
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	// 2. Determinar modelo (default: gemma3:4b)
	model := req.Model
	if model == "" {
		model = "gemma3:4b"
	}

	// 3. Llamar a Ollama
	response, err := s.ollamaClient.Generate(ctx, model, req.Content)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate response: %v", err)
	}

	// 4. Retornar respuesta
	return &mcpv1.SingleChatResponse{
		Content:   response,
		Timestamp: timestamppb.New(time.Now()),
	}, nil
}

// Chat implements bidirectional streaming chat with Gemma 3
func (s *AgentServer) Chat(stream grpc.BidiStreamingServer[mcpv1.ChatMessage, mcpv1.ChatMessage]) error {
	// Create context for this stream
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var sessionID string
	var model string = "gemma3:4b" // default model

	log.Printf("🔄 New streaming chat session started")

	// Handle incoming messages
	for {
		// Receive message from client
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Printf("✅ Client closed stream for session: %s", sessionID)
			break
		}
		if err != nil {
			log.Printf("❌ Stream receive error: %v", err)
			return status.Errorf(codes.Internal, "failed to receive message: %v", err)
		}

		// Extract session info from first message
		if sessionID == "" {
			sessionID = msg.SessionId
			if sessionID == "" {
				return status.Error(codes.InvalidArgument, "session_id is required in first message")
			}

			// Register this stream session
			streamSession := &StreamSession{
				SessionID:    sessionID,
				Stream:       stream,
				Model:        model,
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
				Context:      ctx,
				Cancel:       cancel,
			}

			s.streamsMutex.Lock()
			s.activeStreams[sessionID] = streamSession
			s.streamsMutex.Unlock()

			log.Printf("📝 Registered stream session: %s", sessionID)
		}

		// Update last activity
		s.streamsMutex.Lock()
		if session, exists := s.activeStreams[sessionID]; exists {
			session.LastActivity = time.Now()
		}
		s.streamsMutex.Unlock()

		// Validate message
		if msg.Content == "" {
			// Send error message back to client
			errorMsg := &mcpv1.ChatMessage{
				MessageId: generateMessageID(),
				SessionId: sessionID,
				Content:   "Error: message content cannot be empty",
				Type:      mcpv1.MessageType_MESSAGE_TYPE_SYSTEM,
				Timestamp: timestamppb.New(time.Now()),
			}
			if err := stream.Send(errorMsg); err != nil {
				log.Printf("❌ Failed to send error message: %v", err)
			}
			continue
		}

		log.Printf("💬 Received message from session %s: %s", sessionID, msg.Content[:min(50, len(msg.Content))]+"...")

		// Process message asynchronously to not block receiving
		go func(message *mcpv1.ChatMessage) {
			s.processStreamMessage(ctx, stream, message, model)
		}(msg)
	}

	// Cleanup when stream ends
	s.streamsMutex.Lock()
	delete(s.activeStreams, sessionID)
	s.streamsMutex.Unlock()

	log.Printf("🧹 Cleaned up stream session: %s", sessionID)
	return nil
}

// processStreamMessage handles individual message processing
func (s *AgentServer) processStreamMessage(ctx context.Context, stream grpc.BidiStreamingServer[mcpv1.ChatMessage, mcpv1.ChatMessage], msg *mcpv1.ChatMessage, model string) {
	// Generate response from Ollama
	response, err := s.ollamaClient.Generate(ctx, model, msg.Content)
	if err != nil {
		log.Printf("❌ Ollama error for session %s: %v", msg.SessionId, err)

		// Send error response
		errorMsg := &mcpv1.ChatMessage{
			MessageId: generateMessageID(),
			SessionId: msg.SessionId,
			Content:   fmt.Sprintf("Error generating response: %v", err),
			Type:      mcpv1.MessageType_MESSAGE_TYPE_SYSTEM,
			Timestamp: timestamppb.New(time.Now()),
		}

		if err := stream.Send(errorMsg); err != nil {
			log.Printf("❌ Failed to send error response: %v", err)
		}
		return
	}

	// Send response back to client
	responseMsg := &mcpv1.ChatMessage{
		MessageId: generateMessageID(),
		SessionId: msg.SessionId,
		Content:   response,
		Type:      mcpv1.MessageType_MESSAGE_TYPE_ASSISTANT,
		Timestamp: timestamppb.New(time.Now()),
	}

	if err := stream.Send(responseMsg); err != nil {
		log.Printf("❌ Failed to send response for session %s: %v", msg.SessionId, err)
		return
	}

	log.Printf("✅ Sent response to session %s: %s", msg.SessionId, response[:min(50, len(response))]+"...")
}

// cleanupInactiveStreams removes streams that haven't been active for a while
func (s *AgentServer) cleanupInactiveStreams() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		inactiveThreshold := 5 * time.Minute // 5 minutes of inactivity

		s.streamsMutex.Lock()
		for sessionID, session := range s.activeStreams {
			if now.Sub(session.LastActivity) > inactiveThreshold {
				log.Printf("🧹 Cleaning up inactive stream session: %s", sessionID)
				session.Cancel() // Cancel the context
				delete(s.activeStreams, sessionID)
			}
		}
		s.streamsMutex.Unlock()
	}
}

// GetActiveStreamsCount returns the number of active streaming sessions
func (s *AgentServer) GetActiveStreamsCount() int {
	s.streamsMutex.RLock()
	defer s.streamsMutex.RUnlock()
	return len(s.activeStreams)
}

// generateMessageID creates a unique message ID
func generateMessageID() string {
	bytes := make([]byte, 8) // 64 bits
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
