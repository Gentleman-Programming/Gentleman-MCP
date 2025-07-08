import React, { useState, useEffect, useRef } from "react";
import useGentlemanMCP from "../hooks/useGentlemanMCP";

interface MCPChatBoxProps {
  tenantId: string;
  agentId: string;
  model?: string;
  className?: string;
  onError?: (error: string) => void;
  onSessionChange?: (session: any) => void;
}

const MCPChatBox: React.FC<MCPChatBoxProps> = ({
  tenantId,
  agentId,
  model = "gemma3:4b",
  className = "",
  onError,
  onSessionChange,
}) => {
  const [inputMessage, setInputMessage] = useState("");
  const [autoConnect, setAutoConnect] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const mcp = useGentlemanMCP({
    tenantId,
    agentId,
    model,
  });

  // Scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [mcp.messages]);

  // Notify parent of session changes
  useEffect(() => {
    onSessionChange?.(mcp.session);
  }, [mcp.session, onSessionChange]);

  // Notify parent of errors
  useEffect(() => {
    if (mcp.error) {
      onError?.(mcp.error);
    }
  }, [mcp.error, onError]);

  // Auto-connect if enabled
  useEffect(() => {
    if (autoConnect && !mcp.session && !mcp.isLoading) {
      mcp.register();
    }
  }, [autoConnect, mcp.session, mcp.isLoading, mcp.register]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!inputMessage.trim()) return;

    if (!mcp.session) {
      await mcp.register();
    }

    if (mcp.session) {
      await mcp.sendMessage(inputMessage);
      setInputMessage("");
    }
  };

  const formatTimestamp = (timestamp: Date) => {
    return timestamp.toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  };

  const getMessageIcon = (type: string) => {
    switch (type) {
      case "USER":
        return "👤";
      case "ASSISTANT":
        return "🤖";
      case "SYSTEM":
        return "🔔";
      default:
        return "💬";
    }
  };

  const getMessageClass = (type: string) => {
    switch (type) {
      case "USER":
        return "bg-blue-100 border-blue-200 text-blue-900";
      case "ASSISTANT":
        return "bg-green-100 border-green-200 text-green-900";
      case "SYSTEM":
        return "bg-gray-100 border-gray-200 text-gray-700";
      default:
        return "bg-white border-gray-200";
    }
  };

  return (
    <div className={`flex flex-col h-full ${className}`}>
      {/* Header */}
      <div className="bg-gray-50 border-b border-gray-200 p-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">
              🚀 Gentleman MCP Gateway
            </h2>
            <p className="text-sm text-gray-600">
              {mcp.session && mcp.session.sessionId ? (
                <>
                  ✅ Connected • Session: {mcp.session.sessionId.slice(0, 8)}...
                  <br />
                  🤖 Model: {mcp.session.model} • Tenant: {mcp.session.tenantId}
                </>
              ) : mcp.isLoading ? (
                "🔄 Connecting..."
              ) : (
                "❌ Not connected"
              )}
            </p>
          </div>

          <div className="flex gap-2">
            {(!mcp.session || !mcp.session.sessionId) && (
              <button
                onClick={mcp.register}
                disabled={mcp.isLoading}
                className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50 text-sm"
              >
                {mcp.isLoading ? "🔄 Connecting..." : "🔌 Connect"}
              </button>
            )}

            {mcp.session && mcp.session.sessionId && (
              <button
                onClick={mcp.disconnect}
                className="px-3 py-1 bg-red-500 text-white rounded hover:bg-red-600 text-sm"
              >
                🔌 Disconnect
              </button>
            )}

            <button
              onClick={mcp.clearMessages}
              className="px-3 py-1 bg-gray-500 text-white rounded hover:bg-gray-600 text-sm"
            >
              🧹 Clear
            </button>
          </div>
        </div>

        {/* Auto-connect toggle */}
        <div className="mt-2">
          <label className="flex items-center text-sm">
            <input
              type="checkbox"
              checked={autoConnect}
              onChange={(e) => setAutoConnect(e.target.checked)}
              className="mr-2"
            />
            Auto-connect on mount
          </label>
        </div>

        {/* Error display */}
        {mcp.error && (
          <div className="mt-2 p-2 bg-red-100 border border-red-200 rounded text-red-700 text-sm">
            ❌ {mcp.error}
          </div>
        )}
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {mcp.messages.length === 0 ? (
          <div className="text-center text-gray-500 py-8">
            <div className="text-4xl mb-2">💬</div>
            <p>No messages yet. Start a conversation!</p>
            <p className="text-sm mt-2">
              {!mcp.session || !mcp.session.sessionId
                ? "Connect to the gateway first."
                : "Type a message below."}
            </p>
          </div>
        ) : (
          mcp.messages.map((message) => (
            <div
              key={message.messageId}
              className={`border rounded-lg p-3 ${getMessageClass(message.type)}`}
            >
              <div className="flex items-start gap-2">
                <span className="text-lg">{getMessageIcon(message.type)}</span>
                <div className="flex-1">
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-medium text-sm">
                      {message.type === "USER"
                        ? "You"
                        : message.type === "ASSISTANT"
                          ? "Gemma"
                          : "System"}
                    </span>
                    <span className="text-xs opacity-60">
                      {formatTimestamp(message.timestamp)}
                    </span>
                  </div>
                  <div className="text-sm whitespace-pre-wrap">
                    {message.content}
                  </div>
                </div>
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input form */}
      <form onSubmit={handleSubmit} className="border-t border-gray-200 p-4">
        <div className="flex gap-2">
          <input
            type="text"
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            placeholder={
              mcp.session && mcp.session.sessionId
                ? "Type your message..."
                : "Connect first to send messages"
            }
            disabled={!mcp.session || !mcp.session.sessionId || mcp.isLoading}
            className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
          />
          <button
            type="submit"
            disabled={
              !mcp.session ||
              !mcp.session.sessionId ||
              !inputMessage.trim() ||
              mcp.isLoading
            }
            className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {mcp.isLoading ? "🔄" : "📤"} Send
          </button>
        </div>

        {/* Quick actions */}
        <div className="mt-2 flex flex-wrap gap-1">
          {[
            "Hello Gemma!",
            "How are you today?",
            "Tell me a joke",
            "Explain quantum computing",
            "What's the weather like?",
          ].map((quickMessage) => (
            <button
              key={quickMessage}
              type="button"
              onClick={() => setInputMessage(quickMessage)}
              disabled={!mcp.session || !mcp.session.sessionId}
              className="px-2 py-1 text-xs bg-gray-100 hover:bg-gray-200 rounded border disabled:opacity-50"
            >
              {quickMessage}
            </button>
          ))}
        </div>
      </form>

      {/* Footer info */}
      <div className="bg-gray-50 border-t border-gray-200 px-4 py-2">
        <div className="text-xs text-gray-500 text-center">
          🚀 Gentleman MCP Gateway • Sprint 3: gRPC-Web + React Hooks
          {mcp.session && mcp.session.sessionId && mcp.session.expiresAt && (
            <span className="ml-2">
              • Expires: {mcp.session.expiresAt.toLocaleTimeString()}
            </span>
          )}
        </div>
      </div>
    </div>
  );
};

export default MCPChatBox;
