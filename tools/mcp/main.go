package main

import (
	"fmt"

	"github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
)

type DomainRRArguments struct {
	Domain      string `json:"domain" jsonschema:"required,description=The domain to check"`
	Environment string `json:"environment" jsonschema:"required,description=The environment of domain"`
}
type DomainRRResponse struct {
	Resilient bool `json:"resilient" jsonschema:"required,description=Whether the domain is resilient to regional outages"`
}

func main() {
	// Create an HTTP transport
	transport := http.NewHTTPTransport("/mcp")
	port := 9696
	transport.WithAddr(fmt.Sprintf(":%d", port))

	// Create server with the HTTP transport
	server := mcp_golang.NewServer(transport)

	// err := server.RegisterTool("cadence.rr", "Tell user if their domain is resilient to regional outages", func(arguments DomainRRArguments) (*mcp_golang.ToolResponse, error) {
	// 	fmt.Printf("cadence.rr called with arguments: %+v\n", arguments)

	// 	// run cadence CLI to check if it's a global domain or not
	// 	// e.g. cadence --env arguments.Environment --proxy_region dca --domain arguments.Domain domain describe
	// 	cmd := exec.Command("cadence", "--env", arguments.Environment, "--proxy_region", "dca", "--domain", arguments.Domain, "domain", "describe")
	// 	output, err := cmd.Output()
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	// parse the output of the cadence CLI
	// 	// if it contains "IsGlobal(XDC)Domain: true" then it's a global domain
	// 	// otherwise it's not
	// 	if strings.Contains(string(output), "IsGlobal(XDC)Domain: true") {
	// 		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("Yes, this domain is resilient to regional outages")), nil
	// 	}

	// 	// return the result
	// 	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("No, this domain is not resilient to regional outages")), nil
	// })
	// if err != nil {
	// 	panic(err)
	// }

	fmt.Printf("Server starting on port %d\n", port)
	err := server.Serve()
	if err != nil {
		panic(err)
	}
}
