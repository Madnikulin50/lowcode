package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func initMail(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("send_mail",
		mcp.WithDescription("Send an email notification"),
		mcp.WithString("to", mcp.Description("Comma-separated recipient email addresses"), mcp.Required()),
		mcp.WithString("subject", mcp.Description("Email subject"), mcp.Required()),
		mcp.WithString("body", mcp.Description("Email body content (plain text or HTML)"), mcp.Required()),
		mcp.WithString("cc", mcp.Description("Comma-separated CC email addresses")),
		mcp.WithString("replyTo", mcp.Description("Reply-To address")),
		mcp.WithString("contentType", mcp.Description("'plain' or 'html' (default: plain)")),
		mcp.WithString("attachments", mcp.Description("Comma-separated remote attachment URLs")),
	), handleSendMail)
}

func handleSendMail(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	to := splitTrim(getString(args, "to"))
	if len(to) == 0 {
		return textResult("Error: 'to' is required"), nil
	}

	subject := getString(args, "subject")
	body := getString(args, "body")
	cc := splitTrim(getString(args, "cc"))
	replyTo := getString(args, "replyTo")
	contentType := getString(args, "contentType")
	attachments := splitTrim(getString(args, "attachments"))

	if contentType == "" {
		contentType = "plain"
	}

	n := &types.EmailNotification{
		To:                to,
		Cc:                cc,
		ReplyTo:           replyTo,
		Subject:           subject,
		RemoteAttachments: attachments,
	}

	switch contentType {
	case "html":
		n.ContentHTML = body
	default:
		n.ContentPlain = body
	}

	if err := service.DefaultNotification.SendEmail(ctx, n); err != nil {
		return errorResult(fmt.Errorf("failed to send email: %w", err)), nil
	}

	return jsonResult(map[string]interface{}{
		"sent":    true,
		"to":      to,
		"subject": subject,
	}), nil
}

func splitTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
