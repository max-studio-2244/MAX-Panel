# MAX Panel - CodeSandbox Version

This is a CodeSandbox-compatible version of the MAX Panel game server management system.

## What's Different

- **Mock Docker Operations**: All server management is simulated since CodeSandbox doesn't support Docker
- **Simplified Dependencies**: Removed Docker and crypto dependencies
- **Demo Data**: Shows sample servers and data for demonstration

## Files for CodeSandbox

- `codesandbox-main.go` - Main application with mock handlers
- `package.json` - Node.js package file for CodeSandbox
- `sandbox.config.json` - CodeSandbox configuration
- `go.mod` - Simplified Go modules

## How to Use

1. Upload all files to CodeSandbox
2. Select "Go" template
3. The app will start on port 8080
4. Access the web interface

## Features Working

- ✅ Web interface
- ✅ Mock server listing
- ✅ Mock server creation
- ✅ Mock authentication
- ✅ Admin panel (mock)
- ❌ Real Docker containers
- ❌ File management
- ❌ Real server console

## Demo Login

- Username: `demo`
- Password: `any password`

The panel will show sample Minecraft and CS:GO servers for demonstration.