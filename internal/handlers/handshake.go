package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	mcpv1 "github.com/Gentleman-Programming/gentleman-mcp/proto/mcp/v1"
)

type HandshakeServer struct {
	mcpv1.UnimplementedHandshakeServiceServer
	// In-memory storage for demo purposes
	// In production, this would be Redis or a database
	sessions map[string]*SessionInfo
	tokens   map[string]*TokenInfo
}

type SessionInfo struct {
	SessionID string
	TenantID  string
	AgentID   string
	Model     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type TokenInfo struct {
	TenantID  string
	AgentID   string
	SessionID string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewHandshakeServer() *HandshakeServer {
	return &HandshakeServer{
		sessions: make(map[string]*SessionInfo),
		tokens:   make(map[string]*TokenInfo),
	}
}

// Register creates a new session and returns a JWT token
func (s *HandshakeServer) Register(ctx context.Context, req *mcpv1.RegisterRequest) (*mcpv1.RegisterResponse, error) {
	// Validate request
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}
	if req.Model == "" {
		req.Model = "gemma3:4b" // default model
	}

	// Generate session ID
	sessionID, err := generateRandomID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate session ID: %v", err)
	}

	// Generate JWT token (simplified for demo)
	jwtToken, err := generateRandomID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate JWT token: %v", err)
	}

	// Set expiration (5 minutes as mentioned in the README)
	now := time.Now()
	expiresAt := now.Add(5 * time.Minute)

	// Store session info
	sessionInfo := &SessionInfo{
		SessionID: sessionID,
		TenantID:  req.TenantId,
		AgentID:   req.AgentId,
		Model:     req.Model,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	s.sessions[sessionID] = sessionInfo

	// Store token info
	tokenInfo := &TokenInfo{
		TenantID:  req.TenantId,
		AgentID:   req.AgentId,
		SessionID: sessionID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	s.tokens[jwtToken] = tokenInfo

	return &mcpv1.RegisterResponse{
		SessionId: sessionID,
		JwtToken:  jwtToken,
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

// Authenticate validates a JWT token and returns session info
func (s *HandshakeServer) Authenticate(ctx context.Context, req *mcpv1.AuthRequest) (*mcpv1.AuthResponse, error) {
	if req.JwtToken == "" {
		return nil, status.Error(codes.InvalidArgument, "jwt_token is required")
	}

	// Look up token
	tokenInfo, exists := s.tokens[req.JwtToken]
	if !exists {
		return &mcpv1.AuthResponse{
			Valid: false,
		}, nil
	}

	// Check if token is expired
	if time.Now().After(tokenInfo.ExpiresAt) {
		// Clean up expired token
		delete(s.tokens, req.JwtToken)
		delete(s.sessions, tokenInfo.SessionID)

		return &mcpv1.AuthResponse{
			Valid: false,
		}, nil
	}

	return &mcpv1.AuthResponse{
		Valid:    true,
		TenantId: tokenInfo.TenantID,
		AgentId:  tokenInfo.AgentID,
	}, nil
}

// GetSessionInfo returns session information by session ID (helper method)
func (s *HandshakeServer) GetSessionInfo(sessionID string) (*SessionInfo, bool) {
	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		delete(s.sessions, sessionID)
		return nil, false
	}

	return session, true
}

// generateRandomID creates a cryptographically secure random ID
func generateRandomID() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CleanupExpiredSessions removes expired sessions and tokens (should be called periodically)
func (s *HandshakeServer) CleanupExpiredSessions() {
	now := time.Now()

	// Clean up expired sessions
	for sessionID, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, sessionID)
		}
	}

	// Clean up expired tokens
	for token, tokenInfo := range s.tokens {
		if now.After(tokenInfo.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}

// GetActiveSessionsCount returns the number of active sessions (for monitoring)
func (s *HandshakeServer) GetActiveSessionsCount() int {
	s.CleanupExpiredSessions()
	return len(s.sessions)
}
