# Gentleman MCP Gateway - React Example

![Sprint 3](https://img.shields.io/badge/Sprint-3-green.svg)
![React](https://img.shields.io/badge/React-18-blue.svg)
![TypeScript](https://img.shields.io/badge/TypeScript-5-blue.svg)
![gRPC-Web](https://img.shields.io/badge/gRPC--Web-1.4-orange.svg)

> **React web client for Gentleman MCP Gateway with real-time chat capabilities via gRPC-Web**

## 🚀 Quick Start

### Prerequisites

1. **Backend running**: Make sure the Gentleman MCP Gateway is running with gRPC-Web enabled
2. **Node.js**: Version 16+ required
3. **Ollama**: With Gemma 3 model available

### Setup & Run

```bash
# From the project root
cd gentleman-mcp

# 1. Start the gateway with gRPC-Web
make dev-web

# 2. Install React dependencies and start dev server
make web-dev

# 3. Open browser
open http://localhost:3000
```

## 🎯 Features

### ✅ Implemented (Sprint 3)

- **🔌 gRPC-Web Integration**: Direct browser-to-server communication
- **⚛️ React Hooks**: Custom `useGentlemanMCP` hook for state management
- **💬 Real-time Chat**: Interactive chat interface with Gemma 3
- **🔐 Authentication**: JWT-based session management
- **📱 Responsive UI**: Mobile-friendly design with Tailwind CSS
- **🎨 Modern UX**: Loading states, error handling, auto-scroll
- **⚙️ Configuration**: Easy tenant/agent/model switching
- **🔄 Auto-reconnect**: Session refresh before expiration

### 🔜 Planned (Future Sprints)

- **🌊 Streaming Chat**: Real-time message streaming
- **📁 File Upload**: Multimodal input support
- **🎭 Themes**: Dark mode and custom themes
- **📊 Analytics**: Usage metrics and session history
- **🔌 Plugins**: Extensible component system

## 🏗️ Architecture

```
┌─────────────────────┐
│   React App         │
│   (localhost:3000)  │
├─────────────────────┤
│ • useGentlemanMCP   │
│ • MCPChatBox        │
│ • TypeScript        │
│ • Tailwind CSS     │
└─────────┬───────────┘
          │ gRPC-Web
          ▼
┌─────────────────────┐
│ gRPC-Web Proxy      │
│ (localhost:8080)    │
├─────────────────────┤
│ • CORS handling     │
│ • HTTP/2 to gRPC    │
│ • Static file serve │
└─────────┬───────────┘
          │ gRPC
          ▼
┌─────────────────────┐
│ MCP Gateway         │
│ (localhost:50051)   │
├─────────────────────┤
│ • HandshakeService  │
│ • AgentService      │
│ • Session mgmt      │
└─────────┬───────────┘
          │ HTTP
          ▼
┌─────────────────────┐
│ Ollama + Gemma 3    │
│ (localhost:11434)   │
└─────────────────────┘
```

## 🛠️ Development

### Project Structure

```
web/react-example/
├── public/
│   └── index.html          # App shell with loading screen
├── src/
│   ├── components/
│   │   └── MCPChatBox.tsx  # Main chat component
│   ├── hooks/
│   │   └── useGentlemanMCP.ts # gRPC-Web integration hook
│   ├── App.tsx             # Main app component
│   ├── App.css             # Tailwind + custom styles
│   ├── index.tsx           # React entry point
│   └── index.css           # Base styles
├── package.json            # Dependencies and scripts
└── README.md              # This file
```

### Available Scripts

```bash
# Development
npm start                   # Start dev server
npm test                    # Run tests
npm run build              # Build for production
npm run eject              # Eject from Create React App

# Via Makefile (from project root)
make web-dev               # Start React dev server
make web-build             # Build production
make web-test              # Run tests
make web-install           # Install dependencies
```

### Environment Variables

Create `.env.local` for custom configuration:

```env
# Server URLs
REACT_APP_GRPC_WEB_URL=http://localhost:8080
REACT_APP_GRPC_URL=http://localhost:50051

# Default connection settings
REACT_APP_DEFAULT_TENANT=react-demo
REACT_APP_DEFAULT_MODEL=gemma3:4b

# Feature flags
REACT_APP_ENABLE_STREAMING=true
REACT_APP_ENABLE_DEBUG=true
```

## 🎮 Usage

### Basic Chat

```tsx
import { useGentlemanMCP } from './hooks/useGentlemanMCP';

function MyComponent() {
  const mcp = useGentlemanMCP({
    tenantId: 'my-app',
    agentId: 'user-123',
    model: 'gemma3:4b'
  });

  const handleConnect = async () => {
    await mcp.register();
  };

  const handleSendMessage = async (message: string) => {
    await mcp.sendMessage(message);
  };

  return (
    <div>
      {!mcp.session && (
        <button onClick={handleConnect}>Connect</button>
      )}
      
      {mcp.messages.map(msg => (
        <div key={msg.messageId}>
          <strong>{msg.type}:</strong> {msg.content}
        </div>
      ))}
      
      <button onClick={() => handleSendMessage('Hello!')}>
        Send Message
      </button>
    </div>
  );
}
```

### Advanced Hook Usage

```tsx
const mcp = useGentlemanMCP({
  tenantId: 'advanced-app',
  agentId: 'power-user',
  model: 'gemma3:8b',
  serverUrl: 'https://my-gateway.com'
});

// Auto-connect on mount
useEffect(() => {
  mcp.register();
}, []);

// Handle errors
useEffect(() => {
  if (mcp.error) {
    console.error('MCP Error:', mcp.error);
    // Show notification, retry, etc.
  }
}, [mcp.error]);

// Session monitoring
useEffect(() => {
  if (mcp.session) {
    console.log('Connected:', mcp.session.sessionId);
    // Track analytics, update UI, etc.
  }
}, [mcp.session]);
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
npm test

# Run specific test file
npm test -- --testNamePattern="useGentlemanMCP"

# Run with coverage
npm test -- --coverage --watchAll=false
```

### Integration Testing

```bash
# Start full stack for testing
make dev-web                # Terminal 1
make web-dev               # Terminal 2

# Test scenarios
# 1. Register new session
# 2. Send multiple messages
# 3. Handle disconnection
# 4. Test error scenarios
```

### Manual Testing Checklist

- [ ] **Connection**: Register and authenticate successfully
- [ ] **Chat**: Send messages and receive responses
- [ ] **UI**: Responsive design on mobile/desktop
- [ ] **Errors**: Graceful error handling and recovery
- [ ] **Session**: Auto-refresh before expiration
- [ ] **Settings**: Change tenant/agent/model dynamically
- [ ] **Performance**: Smooth scrolling and interactions

## 🐛 Troubleshooting

### Common Issues

**1. CORS Errors**
```bash
# Check if gRPC-Web proxy is running
curl http://localhost:8080
# Should return HTML page
```

**2. Connection Refused**
```bash
# Verify backend is running
make status
# Check all services are up
```

**3. Module Not Found**
```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install
```

**4. Build Errors**
```bash
# Check TypeScript errors
npm run build
# Fix any type issues
```

### Debug Mode

Enable debug logging in development:

```tsx
// In your component
useEffect(() => {
  if (process.env.NODE_ENV === 'development') {
    console.log('MCP State:', mcp);
  }
}, [mcp]);
```

## 📦 Dependencies

### Core Dependencies

- **React 18**: UI framework
- **TypeScript**: Type safety
- **grpc-web**: Browser gRPC client
- **google-protobuf**: Protocol buffer support

### Development Dependencies

- **@types/***: TypeScript definitions
- **react-scripts**: Build tooling
- **tailwindcss**: CSS framework (via CDN)

### Size Analysis

```bash
# Analyze bundle size
npm run build
npx serve -s build

# Check bundle analyzer
npm install --save-dev webpack-bundle-analyzer
npm run build
npx webpack-bundle-analyzer build/static/js/*.js
```

## 🚀 Deployment

### Production Build

```bash
# Build optimized bundle
make web-build

# Serve locally for testing
npx serve -s build -l 3000
```

### Docker Deployment

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Environment Configuration

```bash
# Production environment variables
REACT_APP_GRPC_WEB_URL=https://api.yourdomain.com
REACT_APP_DEFAULT_TENANT=production
REACT_APP_ENABLE_DEBUG=false
```

## 🤝 Contributing

### Development Workflow

1. **Setup**: Follow Quick Start instructions
2. **Branch**: Create feature branch from `main`
3. **Develop**: Make changes with tests
4. **Test**: Run full test suite
5. **Build**: Verify production build works
6. **PR**: Submit pull request with description

### Code Style

- **TypeScript**: Strict mode enabled
- **ESLint**: Follow React/TypeScript rules
- **Prettier**: Auto-format on save
- **Naming**: Use descriptive component/function names

### Testing Guidelines

- **Unit Tests**: Test hooks and utilities
- **Component Tests**: Test user interactions
- **Integration Tests**: Test full workflows
- **Accessibility**: Test keyboard navigation and screen readers

## 📄 License

This project is part of Gentleman MCP Gateway and follows the same license terms.

## 🔗 Links

- **Main Project**: [Gentleman MCP Gateway](../../README.md)
- **gRPC-Web Docs**: https://grpc.io/docs/platforms/web/
- **React Docs**: https://react.dev/
- **Tailwind CSS**: https://tailwindcss.com/

---

**Sprint 3**: ✅ gRPC-Web + React Hooks Integration Complete!