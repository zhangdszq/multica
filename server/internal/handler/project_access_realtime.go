package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// AuthorizeProjectMessage rechecks access at delivery time, including existing
// task subscriptions. Revoking access must not leave an open websocket as a
// permanent bypass. Database failures deny delivery.
func (h *Handler) AuthorizeProjectMessage(userID, workspaceID, scopeType, scopeID string, message []byte) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	projects := map[string]bool{}
	issues := map[string]bool{}
	tasks := map[string]bool{}
	if scopeType == "task" {
		tasks[scopeID] = true
	}
	var body any
	if json.Unmarshal(message, &body) != nil {
		return false
	}
	// Deletion notifications carry only an opaque ID; the deleted row can no
	// longer authorize delivery. Allow these cache invalidations, never content.
	if event, ok := body.(map[string]any); ok {
		kind, _ := event["type"].(string)
		payload, _ := event["payload"].(map[string]any)
		key := ""
		if kind == "project:deleted" {
			key = "project_id"
		}
		if kind == "issue:deleted" {
			key = "issue_id"
		}
		if key != "" && len(payload) == 1 && scopeType == "workspace" {
			if id, ok := payload[key].(string); ok {
				_, err := util.ParseUUID(id)
				return err == nil
			}
		}
	}
	var walk func(any, string)
	walk = func(value any, parent string) {
		switch value := value.(type) {
		case map[string]any:
			for key, child := range value {
				if id, ok := child.(string); ok && id != "" {
					switch key {
					case "project_id":
						projects[id] = true
					case "issue_id":
						issues[id] = true
					case "task_id":
						tasks[id] = true
					case "id":
						if parent == "project" {
							projects[id] = true
						}
						if parent == "issue" {
							issues[id] = true
						}
					}
				}
				walk(child, key)
			}
		case []any:
			for _, child := range value {
				walk(child, parent)
			}
		}
	}
	walk(body, "")
	for value := range tasks {
		id, err := util.ParseUUID(value)
		if err != nil {
			return false
		}
		task, err := h.Queries.GetAgentTask(ctx, id)
		if err != nil {
			return false
		}
		if task.IssueID.Valid {
			issues[uuidToString(task.IssueID)] = true
		}
	}
	for value := range issues {
		id, err := util.ParseUUID(value)
		if err != nil {
			return false
		}
		issue, err := h.Queries.GetIssue(ctx, id)
		if err != nil {
			return false
		}
		if issue.ProjectID.Valid {
			projects[uuidToString(issue.ProjectID)] = true
		}
	}
	uid, err := util.ParseUUID(userID)
	if len(projects) > 0 && err != nil {
		return false
	}
	for value := range projects {
		id, err := util.ParseUUID(value)
		if err != nil {
			return false
		}
		var wsID pgtype.UUID
		// User-scoped inbox events may refer to another workspace. Resolve
		// the project's workspace, then check current membership there too.
		if err := h.DB.QueryRow(ctx, "SELECT workspace_id FROM project WHERE id = $1", id).Scan(&wsID); err != nil {
			return false
		}
		p, err := h.Queries.GetProjectInWorkspace(ctx, db.GetProjectInWorkspaceParams{ID: id, WorkspaceID: wsID})
		if err != nil || !projectAllowsUser(p, userID) {
			return false
		}
		if _, err := h.Queries.GetMemberByUserAndWorkspace(ctx, db.GetMemberByUserAndWorkspaceParams{UserID: uid, WorkspaceID: wsID}); err != nil {
			return false
		}
	}
	return true
}
