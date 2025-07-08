import { useEffect, useState } from "react";
import { useGentlemanMCP } from "./hooks/useGentlemanMCP";

function App() {
  const [message, setMessage] = useState("");
  const [autoConnected, setAutoConnected] = useState(false);

  // Initialize the MCP hook
  const mcp = useGentlemanMCP({
    tenantId: "demo-tenant",
    agentId: "demo-user",
    model: "gemma3:4b",
    serverUrl: "http://localhost:8080",
  });

  // Auto-connect on component mount
  useEffect(() => {
    if (!autoConnected && !mcp.session && !mcp.isLoading) {
      setAutoConnected(true);
      mcp.register().catch((error) => {
        console.error("Auto-connect failed:", error);
      });
    }
  }, [mcp, autoConnected]);

  const handleSendMessage = async () => {
    if (!message.trim()) return;

    const userMsg = message;
    setMessage("");

    try {
      if (!mcp.session) {
        await mcp.register();
      }
      await mcp.sendMessage(userMsg);
    } catch (error) {
      console.error("Failed to send message:", error);
    }
  };

  const handleConnect = async () => {
    try {
      await mcp.register();
    } catch (error) {
      console.error("Connection failed:", error);
    }
  };

  const handleDisconnect = () => {
    mcp.disconnect();
    setAutoConnected(false);
  };

  const formatMessageContent = (content: string) => {
    // Add some basic formatting for better readability
    return content
      .replace(/\n/g, "\n")
      .replace(/✅/g, "✅")
      .replace(/❌/g, "❌")
      .replace(/🤖/g, "🤖")
      .replace(/💬/g, "💬");
  };

  return (
    <div
      style={{
        minHeight: "100vh",
        backgroundColor: "#f5f5f5",
        fontFamily: "Arial, sans-serif",
      }}
    >
      {/* Header */}
      <header
        style={{
          backgroundColor: "white",
          padding: "1rem",
          borderBottom: "1px solid #e5e5e5",
        }}
      >
        <div style={{ maxWidth: "800px", margin: "0 auto" }}>
          <h1 style={{ margin: 0, color: "#333" }}>🚀 Gentleman MCP Gateway</h1>
          <p
            style={{
              margin: "0.5rem 0 0 0",
              color: "#666",
              fontSize: "0.9rem",
            }}
          >
            Sprint 3: React + gRPC-Web Integration
          </p>
          <div style={{ marginTop: "0.5rem" }}>
            <span
              style={{
                padding: "0.25rem 0.5rem",
                borderRadius: "4px",
                fontSize: "0.8rem",
                backgroundColor: mcp.isConnected ? "#d4edda" : "#f8d7da",
                color: mcp.isConnected ? "#155724" : "#721c24",
              }}
            >
              {mcp.isConnected ? "✅ Connected" : "❌ Disconnected"}
            </span>
            {mcp.session?.model && (
              <span
                style={{
                  marginLeft: "0.5rem",
                  padding: "0.25rem 0.5rem",
                  borderRadius: "4px",
                  fontSize: "0.8rem",
                  backgroundColor: "#d1ecf1",
                  color: "#0c5460",
                }}
              >
                Model: {mcp.session.model}
              </span>
            )}
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main
        style={{ maxWidth: "800px", margin: "0 auto", padding: "2rem 1rem" }}
      >
        <div
          style={{
            backgroundColor: "white",
            borderRadius: "8px",
            padding: "1.5rem",
            boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
          }}
        >
          <h2 style={{ marginTop: 0 }}>Chat Interface</h2>

          {/* Error Display */}
          {mcp.error && (
            <div
              style={{
                backgroundColor: "#f8d7da",
                color: "#721c24",
                padding: "0.75rem",
                borderRadius: "4px",
                marginBottom: "1rem",
                border: "1px solid #f5c6cb",
              }}
            >
              <strong>Error:</strong> {mcp.error}
            </div>
          )}

          {/* Response Area */}
          <div
            style={{
              backgroundColor: "#f8f9fa",
              border: "1px solid #e9ecef",
              borderRadius: "4px",
              padding: "1rem",
              minHeight: "400px",
              maxHeight: "500px",
              marginBottom: "1rem",
              fontSize: "0.9rem",
              overflowY: "auto",
              fontFamily: "monospace",
            }}
          >
            {mcp.messages.length === 0 ? (
              <div
                style={{
                  textAlign: "center",
                  color: "#6c757d",
                  paddingTop: "3rem",
                }}
              >
                <div style={{ fontSize: "2rem", marginBottom: "0.5rem" }}>
                  💬
                </div>
                <p>
                  {mcp.isLoading
                    ? "Connecting to Gemma 3..."
                    : mcp.isConnected
                      ? "Ready to chat with Gemma 3!"
                      : "Connect to start chatting"}
                </p>
              </div>
            ) : (
              <div>
                {mcp.messages.map((msg) => (
                  <div key={msg.messageId} style={{ marginBottom: "1rem" }}>
                    <div
                      style={{
                        fontSize: "0.8rem",
                        color: "#6c757d",
                        marginBottom: "0.25rem",
                      }}
                    >
                      <span
                        style={{
                          fontWeight: "bold",
                          color:
                            msg.type === "USER"
                              ? "#007bff"
                              : msg.type === "ASSISTANT"
                                ? "#28a745"
                                : "#6c757d",
                        }}
                      >
                        {msg.type === "USER"
                          ? "💬 You"
                          : msg.type === "ASSISTANT"
                            ? "🤖 Gemma"
                            : "ℹ️ System"}
                      </span>
                      <span style={{ marginLeft: "0.5rem" }}>
                        {msg.timestamp.toLocaleTimeString()}
                      </span>
                    </div>
                    <div
                      style={{
                        padding: "0.5rem",
                        borderRadius: "4px",
                        backgroundColor:
                          msg.type === "USER"
                            ? "#e3f2fd"
                            : msg.type === "ASSISTANT"
                              ? "#e8f5e8"
                              : "#fff3cd",
                        whiteSpace: "pre-wrap",
                        wordBreak: "break-word",
                      }}
                    >
                      {formatMessageContent(msg.content)}
                    </div>
                  </div>
                ))}
                {mcp.isLoading && (
                  <div
                    style={{
                      textAlign: "center",
                      color: "#6c757d",
                      fontStyle: "italic",
                    }}
                  >
                    🤖 Gemma is thinking...
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Controls */}
          <div style={{ marginBottom: "1rem" }}>
            {!mcp.isConnected ? (
              <button
                type="button"
                onClick={handleConnect}
                disabled={mcp.isLoading}
                style={{
                  backgroundColor: "#007bff",
                  color: "white",
                  border: "none",
                  padding: "0.5rem 1rem",
                  borderRadius: "4px",
                  cursor: mcp.isLoading ? "not-allowed" : "pointer",
                  opacity: mcp.isLoading ? 0.6 : 1,
                  marginRight: "0.5rem",
                }}
              >
                {mcp.isLoading ? "🔄 Connecting..." : "🔌 Connect to Gateway"}
              </button>
            ) : (
              <button
                type="button"
                onClick={handleDisconnect}
                style={{
                  backgroundColor: "#dc3545",
                  color: "white",
                  border: "none",
                  padding: "0.5rem 1rem",
                  borderRadius: "4px",
                  cursor: "pointer",
                  marginRight: "0.5rem",
                }}
              >
                🔌 Disconnect
              </button>
            )}

            <button
              type="button"
              onClick={mcp.clearMessages}
              disabled={mcp.messages.length === 0}
              style={{
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                padding: "0.5rem 1rem",
                borderRadius: "4px",
                cursor: mcp.messages.length === 0 ? "not-allowed" : "pointer",
                opacity: mcp.messages.length === 0 ? 0.6 : 1,
              }}
            >
              🗑️ Clear Chat
            </button>
          </div>

          {/* Message Input */}
          <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
            <input
              type="text"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder={
                mcp.isConnected
                  ? "Type your message to Gemma 3..."
                  : "Connect first to chat"
              }
              disabled={!mcp.isConnected || mcp.isLoading}
              onKeyDown={(e) => e.key === "Enter" && handleSendMessage()}
              style={{
                flex: 1,
                padding: "0.5rem",
                border: "1px solid #ced4da",
                borderRadius: "4px",
                fontSize: "1rem",
              }}
            />
            <button
              type="button"
              onClick={handleSendMessage}
              disabled={!mcp.isConnected || !message.trim() || mcp.isLoading}
              style={{
                backgroundColor: "#28a745",
                color: "white",
                border: "none",
                padding: "0.5rem 1rem",
                borderRadius: "4px",
                cursor:
                  !mcp.isConnected || !message.trim() || mcp.isLoading
                    ? "not-allowed"
                    : "pointer",
                opacity:
                  !mcp.isConnected || !message.trim() || mcp.isLoading
                    ? 0.6
                    : 1,
              }}
            >
              {mcp.isLoading ? "🔄" : "📤"} Send
            </button>
          </div>

          {/* Quick Messages */}
          <div style={{ display: "flex", flexWrap: "wrap", gap: "0.5rem" }}>
            {[
              "Hello Gemma! How are you today?",
              "Can you explain React hooks?",
              "Tell me a programming joke",
              "What is gRPC and how does it work?",
              "Explain the difference between REST and GraphQL",
            ].map((msg) => (
              <button
                type="button"
                key={msg}
                onClick={() => setMessage(msg)}
                disabled={!mcp.isConnected}
                style={{
                  padding: "0.25rem 0.5rem",
                  fontSize: "0.8rem",
                  border: "1px solid #ced4da",
                  borderRadius: "4px",
                  backgroundColor: mcp.isConnected ? "#f8f9fa" : "#e9ecef",
                  cursor: mcp.isConnected ? "pointer" : "not-allowed",
                  opacity: mcp.isConnected ? 1 : 0.6,
                  maxWidth: "200px",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
                title={msg}
              >
                {msg}
              </button>
            ))}
          </div>
        </div>

        {/* Status & Diagnostics */}
        <div
          style={{
            backgroundColor: "white",
            borderRadius: "8px",
            padding: "1rem",
            marginTop: "1rem",
            boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
          }}
        >
          <h3 style={{ marginTop: 0 }}>🔧 Connection Status</h3>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))",
              gap: "0.5rem",
              fontSize: "0.9rem",
            }}
          >
            <div>{mcp.isConnected ? "✅" : "❌"} gRPC-Web Gateway</div>
            <div>
              {mcp.session ? "✅" : "❌"} Session:{" "}
              {mcp.session?.sessionId.slice(0, 12) || "None"}...
            </div>
            <div>🤖 Model: {mcp.session?.model || "Not selected"}</div>
            <div>📊 Messages: {mcp.messages.length}</div>
          </div>

          {mcp.session?.tenantId &&
            mcp.session.agentId &&
            mcp.session.expiresAt && (
              <div
                style={{
                  marginTop: "0.5rem",
                  padding: "0.5rem",
                  backgroundColor: "#f8f9fa",
                  borderRadius: "4px",
                  fontSize: "0.8rem",
                  color: "#6c757d",
                }}
              >
                <div>🏢 Tenant: {mcp.session.tenantId}</div>
                <div>👤 Agent: {mcp.session.agentId}</div>
                <div>⏰ Expires: {mcp.session.expiresAt.toLocaleString()}</div>
              </div>
            )}

          <p
            style={{
              fontSize: "0.8rem",
              color: "#6c757d",
              marginBottom: 0,
              marginTop: "0.5rem",
            }}
          >
            🚀 Real gRPC-Web integration with Gentleman MCP Gateway → Backend Go
            Server → Ollama → Gemma 3
          </p>
        </div>
      </main>

      {/* Footer */}
      <footer
        style={{
          backgroundColor: "white",
          borderTop: "1px solid #e5e5e5",
          padding: "1rem",
          textAlign: "center",
          color: "#6c757d",
          fontSize: "0.8rem",
        }}
      >
        🚀 Gentleman MCP Gateway • Sprint 3: React + gRPC-Web + Gemma 3
      </footer>
    </div>
  );
}

export default App;
