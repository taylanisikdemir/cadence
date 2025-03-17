## Overview

This folder contains an [MCP server](https://modelcontextprotocol.io/introduction) which exposes useful Cadence tools to Cursor.

## How to install
This should be integrated to Cursors inside devpod seamlessly. For now, follow manual steps below:

1. Build the server executable
```
mkdir -p .bin && go build -o .bin/cadence_mcp tools/mcp/main.go
```


2. Update .cursor/mcp.json with following entry:
```
{
"mcpServers": {
  "cadence-mcp-server": {
      "command": ".bin/cadence_mcp",
      "args": [],
      "env": {}
    }
  }
}
```

## Usage

1. Is my domain resilient to regional outages?
TODO: Explain what regional resiliency is and insert doc links.

For now, it will tell you "Yes" if the domain is global, and "No" otherwise.
```
% cadence --env prod11 --proxy_region dca --domain cadence-system domain describe
Name: cadence-system
IsGlobal(XDC)Domain: false
....
```
