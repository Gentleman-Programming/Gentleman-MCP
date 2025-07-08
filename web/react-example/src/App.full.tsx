import React, { useState } from 'react';
import MCPChatBox from './components/MCPChatBox';
import './App.css';

interface AppState {
  tenantId: string;
  agentId: string;
  model: string;
  showSettings: boolean;
  currentSession: any;
  errors: string[];
}

function App() {
  const [state, setState] = useState<AppState>({
    tenantId: 'react-demo',
    agentId: 'user-' + Math.random().toString(36).substr(2, 9),
    model: 'gemma3:4b',
    showSettings: false,
    currentSession: null,
    errors: [],
  });

  const updateState = (updates: Partial<AppState>) => {
    setState(prev => ({ ...prev, ...updates }));
  };

  const handleError = (error: string) => {
    setState(prev => ({
      ...prev,
      errors: [...prev.errors.slice(-4), error], // Keep last 5 errors
    }));

    // Auto-clear error after 5 seconds
    setTimeout(() => {
      setState(prev => ({
        ...prev,
        errors: prev.errors.filter(e => e !== error),
      }));
    }, 5000);
  };

  const handleSessionChange = (session: any) => {
    updateState({ currentSession: session });
  };

  const clearAllErrors = () => {
    updateState({ errors: [] });
  };

  return (
    <div className="App min-h-screen bg-gray-100">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <h1 className="text-2xl font-bold text-gray-900">
                🚀 Gentleman MCP Gateway
              </h1>
              <span className="bg-green-100 text-green-800 text-xs font-medium px-2.5 py-0.5 rounded">
                Sprint 3: React + gRPC-Web
              </span>
            </div>

            <div className="flex items-center space-x-4">
              {state.currentSession && (
                <div className="text-sm text-gray-600">
                  ✅ Connected • {state.currentSession.model}
                </div>
              )}
              <button
                onClick={() => updateState({ showSettings: !state.showSettings })}
                className="px-3 py-1 bg-gray-500 text-white rounded hover:bg-gray-600 text-sm"
              >
                ⚙️ Settings
              </button>
            </div>
          </div>

          {/* Settings Panel */}
          {state.showSettings && (
            <div className="mt-4 p-4 bg-gray-50 rounded-lg border">
              <h3 className="text-lg font-medium mb-3">Connection Settings</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Tenant ID
                  </label>
                  <input
                    type="text"
                    value={state.tenantId}
                    onChange={(e) => updateState({ tenantId: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="your-tenant-id"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Agent ID
                  </label>
                  <input
                    type="text"
                    value={state.agentId}
                    onChange={(e) => updateState({ agentId: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="user-123"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Model
                  </label>
                  <select
                    value={state.model}
                    onChange={(e) => updateState({ model: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="gemma3:4b">Gemma 3 4B</option>
                    <option value="gemma3:8b">Gemma 3 8B</option>
                    <option value="gemma3:27b">Gemma 3 27B</option>
                  </select>
                </div>
              </div>
              <div className="mt-3 flex items-center space-x-2">
                <button
                  onClick={() => updateState({
                    agentId: 'user-' + Math.random().toString(36).substr(2, 9)
                  })}
                  className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 text-sm"
                >
                  🎲 Random Agent ID
                </button>
                <button
                  onClick={() => updateState({ showSettings: false })}
                  className="px-3 py-1 bg-gray-500 text-white rounded hover:bg-gray-600 text-sm"
                >
                  ✅ Done
                </button>
              </div>
            </div>
          )}

          {/* Error Messages */}
          {state.errors.length > 0 && (
            <div className="mt-4">
              {state.errors.map((error, index) => (
                <div
                  key={index}
                  className="mb-2 p-3 bg-red-100 border border-red-200 rounded text-red-700 flex items-center justify-between"
                >
                  <span>❌ {error}</span>
                  <button
                    onClick={clearAllErrors}
                    className="text-red-500 hover:text-red-700 ml-2"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Chat Interface */}
          <div className="lg:col-span-3">
            <div className="bg-white rounded-lg shadow-sm border h-[600px]">
              <MCPChatBox
                tenantId={state.tenantId}
                agentId={state.agentId}
                model={state.model}
                onError={handleError}
                onSessionChange={handleSessionChange}
                className="h-full"
              />
            </div>
          </div>

          {/* Sidebar */}
          <div className="lg:col-span-1 space-y-6">
            {/* Connection Info */}
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <h3 className="text-lg font-medium mb-3">Connection Info</h3>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Status:</span>
                  <span className={state.currentSession ? 'text-green-600' : 'text-red-600'}>
                    {state.currentSession ? '✅ Connected' : '❌ Disconnected'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Tenant:</span>
                  <span className="font-mono text-xs">{state.tenantId}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Agent:</span>
                  <span className="font-mono text-xs">{state.agentId}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Model:</span>
                  <span className="font-mono text-xs">{state.model}</span>
                </div>
                {state.currentSession && (
                  <>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Session:</span>
                      <span className="font-mono text-xs">
                        {state.currentSession.sessionId?.slice(0, 8)}...
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Expires:</span>
                      <span className="text-xs">
                        {new Date(state.currentSession.expiresAt).toLocaleTimeString()}
                      </span>
                    </div>
                  </>
                )}
              </div>
            </div>

            {/* Sprint 3 Features */}
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <h3 className="text-lg font-medium mb-3">🚀 Sprint 3 Features</h3>
              <div className="space-y-2 text-sm">
                <div className="flex items-center space-x-2">
                  <span className="text-green-600">✅</span>
                  <span>gRPC-Web proxy server</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-600">✅</span>
                  <span>React hooks integration</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-600">✅</span>
                  <span>TypeScript support</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-600">✅</span>
                  <span>CORS handling</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-600">✅</span>
                  <span>Real-time chat UI</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-600">✅</span>
                  <span>Session management</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-600">✅</span>
                  <span>Auto-reconnection</span>
                </div>
              </div>
            </div>

            {/* Technical Info */}
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <h3 className="text-lg font-medium mb-3">🔧 Technical Stack</h3>
              <div className="space-y-1 text-xs text-gray-600">
                <div>• React 18 + TypeScript</div>
                <div>• gRPC-Web client</div>
                <div>• TailwindCSS styling</div>
                <div>• Custom React hooks</div>
                <div>• Protobuf code generation</div>
                <div>• JWT authentication</div>
                <div>• Session persistence</div>
                <div>• Error boundary handling</div>
              </div>
            </div>

            {/* Quick Actions */}
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <h3 className="text-lg font-medium mb-3">⚡ Quick Actions</h3>
              <div className="space-y-2">
                <button
                  onClick={() => window.location.reload()}
                  className="w-full px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded text-sm text-left"
                >
                  🔄 Reload App
                </button>
                <button
                  onClick={() => console.log('Current state:', state)}
                  className="w-full px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded text-sm text-left"
                >
                  📊 Log State
                </button>
                <button
                  onClick={() => updateState({ errors: [] })}
                  className="w-full px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded text-sm text-left"
                >
                  🧹 Clear Errors
                </button>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-white border-t mt-8">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="text-center text-sm text-gray-600">
            🚀 Gentleman MCP Gateway • Sprint 3: gRPC-Web + React Integration
            <br />
            <span className="text-xs">
              Built with React, TypeScript, gRPC-Web, and ❤️
            </span>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;
