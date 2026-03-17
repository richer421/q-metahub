package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/richer421/q-metahub/app/metadata"
	"github.com/richer421/q-metahub/conf"
	openmodeloam "github.com/richer421/q-metahub/pkg/openModel/oam"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Run() error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "q-metahub",
		Version: "1.0.0",
	}, nil)

	s.registerTools(server)
	return server.Run(context.Background(), &mcp.StdioTransport{})
}

func (s *Server) registerTools(server *mcp.Server) {
	type readLogsArgs struct {
		Lines int `json:"lines,omitempty" jsonschema:"Number of lines to read (default 100)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_logs",
		Description: "Read last N lines from log file",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args readLogsArgs) (*mcp.CallToolResult, any, error) {
		lines := args.Lines
		if lines <= 0 {
			lines = 100
		}
		result, err := s.handleReadLogs(lines)
		if err != nil {
			return s.errorResult(err), nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	type getDeployPlanArgs struct {
		DeployPlanID int64 `json:"deploy_plan_id"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_open_model_deploy_plan",
		Description: "Get open model deploy plan by deploy plan ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in getDeployPlanArgs) (*mcp.CallToolResult, any, error) {
		res, err := metadata.NewApp(ctx).GetDeployPlan(in.DeployPlanID)
		if err != nil {
			return s.errorResult(err), nil, nil
		}
		out, err := s.jsonResult(openmodeloam.ToOpenModelDeployPlan(res))
		return out, nil, err
	})
}

func (s *Server) jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil
}

func (s *Server) errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(`{"error":%q}`, err.Error())}},
	}
}

func (s *Server) handleReadLogs(lines int) (string, error) {
	logPath := conf.C.Log.File.Path
	if logPath == "" {
		logPath = "logs/app.log"
	}

	content, err := readLastLines(logPath, lines)
	if err != nil {
		return "", err
	}
	return content, nil
}

func readLastLines(path string, n int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	return strings.Join(lines, "\n"), scanner.Err()
}
