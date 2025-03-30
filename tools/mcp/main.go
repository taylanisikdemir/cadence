package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime/debug"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {

	// Create MCP server
	s := server.NewMCPServer(
		"Cadence MCP",
		"0.0.1",
		server.WithLogging(),
	)

	// Add tool handlers
	s.AddTool(mcp.NewTool("domain_rr",
		mcp.WithDescription("Check if a cadence domain is resilient to regional outages"),
		mcp.WithString("domain",
			mcp.Required(),
			mcp.Description("Name of the cadence domain to check"),
		),
		mcp.WithString("grpc_endpoint",
			mcp.DefaultString("localhost:7833"),
			mcp.Description("gRPC endpoint of the cadence domain"),
		),
		mcp.WithString("environment",
			mcp.DefaultString("development"),
			mcp.Description("Environment of the cadence domain"),
		),
	), domainRRHandler)

	debugLog("Cadence MCP started")

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		debugLog("Server error: %v\n", err)
	}

	debugLog("Cadence MCP stopped")
}

func domainRRHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	defer func() {
		// recover from panic
		if r := recover(); r != nil {
			// include the stack trace
			debugLog("Panic: %v\n", r)
			debugLog("Stack trace: %s\n", string(debug.Stack()))
		}
	}()

	domain, ok := request.Params.Arguments["domain"].(string)
	if !ok {
		return nil, errors.New("domain must be a string")
	}

	endpoint, ok := request.Params.Arguments["grpc_endpoint"].(string)
	if !ok {
		endpoint = "localhost:7833"
	}

	environment, ok := request.Params.Arguments["environment"].(string)
	if !ok {
		environment = "development"
	}

	// run cadence CLI to check if it's a global domain or not
	cmd := exec.Command("cadence",
		"--transport", "grpc",
		"--address", endpoint,
		"--env", environment,
		"--domain", domain,
		"domain", "describe")
	output, err := cmd.Output()
	if err != nil {
		debugLog("Error checking domain resilience: %v, %v\n", err, string(output))
		return mcp.NewToolResultText("Error checking domain resilience: " + err.Error() + "\n" + string(output)), nil
	}

	// parse the output of the cadence CLI
	// if it contains "IsGlobal(XDC)Domain: true" then it's a global domain
	// otherwise it's not
	if strings.Contains(string(output), "IsGlobal(XDC)Domain: true") {
		return mcp.NewToolResultText("Yes, this domain is resilient to regional outages"), nil
	}

	return mcp.NewToolResultText("No, this domain is not resilient to regional outages"), nil
}

func debugLog(format string, args ...interface{}) {
	// get the path of the binary
	binaryPath, err := os.Executable()
	if err != nil {
		fmt.Println("Failed to get executable path:", err)
		return
	}
	logFile, err := os.OpenFile(path.Join(path.Dir(binaryPath), "cadence_mcp.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return
	}
	defer logFile.Close()

	logFile.WriteString(fmt.Sprintf(format, args...))
	logFile.WriteString("\n")
}
