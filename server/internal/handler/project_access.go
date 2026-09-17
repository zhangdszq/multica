package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
	"github.com/multica-ai/multica/server/pkg/protocol"
)

// projectVisibilitySQL is composed only from constant column names and SQL
// placeholders. The principal is always a bound parameter.
func projectVisibilitySQL(projectColumn, workspaceColumn, userParam string) string {
	return fmt.Sprintf(`NOT EXISTS (SELECT 1 FROM project pa WHERE pa.id = %s
        AND pa.workspace_id = %s AND pa.access_restricted
        AND pa.created_by IS DISTINCT FROM %s::uuid
        AND NOT (%s::uuid = ANY(pa.allowed_user_ids)))`, projectColumn, workspaceColumn, userParam, userParam)
}

func projectPrincipal(r *http.Request) string {
	if id := requestUserID(r); id != "" {
		return id
	}
	return "00000000-0000-0000-0000-000000000000"
}

func (h *Handler) RequireIssueProjectAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := h.loadIssueForUser(w, r, chi.URLParam(r, "id")); !ok {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) canReadProjectIssue(r *http.Request, issueID, workspaceID pgtype.UUID) bool {
	if !issueID.Valid {
		return true
	}
	issue, err := h.Queries.GetIssueInWorkspace(r.Context(), db.GetIssueInWorkspaceParams{ID: issueID, WorkspaceID: workspaceID})
	if err != nil {
		return false
	}
	if !issue.ProjectID.Valid {
		return true
	}
	p, err := h.Queries.GetProjectInWorkspace(r.Context(), db.GetProjectInWorkspaceParams{ID: issue.ProjectID, WorkspaceID: workspaceID})
	return err == nil && projectAllowsUser(p, requestUserID(r))
}

func projectAllowsUser(p db.Project, userID string) bool {
	if !p.AccessRestricted {
		return true
	}
	if userID == "" {
		return false
	}
	if p.CreatedBy.Valid && uuidToString(p.CreatedBy) == userID {
		return true
	}
	for _, id := range p.AllowedUserIds {
		if uuidToString(id) == userID {
			return true
		}
	}
	return false
}

// Project responses are only emitted after the caller's access has been
// checked. AccessAllowed remains part of the response contract for compatible
// clients, but a false response must never cross the HTTP boundary.
func (h *Handler) projectForRequest(r *http.Request, p db.Project) ProjectResponse {
	resp := projectToResponse(p)
	resp.AccessAllowed = projectAllowsUser(p, requestUserID(r))
	resp.CanManageAccess = p.CreatedBy.Valid && uuidToString(p.CreatedBy) == requestUserID(r)
	resp.AllowedUserIDs = []string{}
	if resp.CanManageAccess {
		for _, id := range p.AllowedUserIds {
			resp.AllowedUserIDs = append(resp.AllowedUserIDs, uuidToString(id))
		}
	}
	if !resp.AccessAllowed {
		resp.Description = nil
		resp.LeadType, resp.LeadID = nil, nil
		resp.StartDate, resp.DueDate = nil, nil
		resp.IssueCount, resp.DoneCount, resp.ResourceCount = 0, 0, 0
	}
	return resp
}

func (h *Handler) requireProjectAccess(w http.ResponseWriter, r *http.Request, projectID pgtype.UUID) bool {
	if !projectID.Valid {
		return true
	}
	wsID, ok := parseUUIDOrBadRequest(w, h.resolveWorkspaceID(r), "workspace_id")
	if !ok {
		return false
	}
	p, err := h.Queries.GetProjectInWorkspace(r.Context(), db.GetProjectInWorkspaceParams{ID: projectID, WorkspaceID: wsID})
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return false
	}
	if !projectAllowsUser(p, requestUserID(r)) {
		writeError(w, http.StatusNotFound, "resource not found")
		return false
	}
	return true
}

func (h *Handler) UpdateProjectAccess(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Actor-Source") == "task_token" {
		writeError(w, http.StatusForbidden, "project access must be managed by a member")
		return
	}
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, ok := parseUUIDOrBadRequest(w, chi.URLParam(r, "id"), "project_id")
	if !ok {
		return
	}
	wsID, ok := parseUUIDOrBadRequest(w, h.resolveWorkspaceID(r), "workspace_id")
	if !ok {
		return
	}
	p, err := h.Queries.GetProjectInWorkspace(r.Context(), db.GetProjectInWorkspaceParams{ID: id, WorkspaceID: wsID})
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if !projectAllowsUser(p, userID) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if !p.CreatedBy.Valid || uuidToString(p.CreatedBy) != userID {
		writeError(w, http.StatusForbidden, "only the project creator can manage access")
		return
	}
	var req struct {
		Restricted *bool    `json:"access_restricted"`
		UserIDs    []string `json:"allowed_user_ids"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&req); err != nil || req.Restricted == nil {
		writeError(w, http.StatusBadRequest, "access_restricted and a valid member list are required")
		return
	}
	if len(req.UserIDs) > 1000 {
		writeError(w, http.StatusBadRequest, "too many members")
		return
	}
	ids := []pgtype.UUID{}
	seen := map[string]bool{}
	for _, value := range req.UserIDs {
		memberID, ok := parseUUIDOrBadRequest(w, value, "allowed_user_ids")
		if !ok {
			return
		}
		if _, err := h.Queries.GetMemberByUserAndWorkspace(r.Context(), db.GetMemberByUserAndWorkspaceParams{UserID: memberID, WorkspaceID: wsID}); err != nil {
			writeError(w, http.StatusBadRequest, "selected members must belong to this workspace")
			return
		}
		key := uuidToString(memberID)
		if !seen[key] {
			ids = append(ids, memberID)
			seen[key] = true
		}
	}
	p, err = h.Queries.UpdateProjectAccess(r.Context(), db.UpdateProjectAccessParams{
		ID: id, WorkspaceID: wsID, AccessRestricted: *req.Restricted, AllowedUserIds: ids, CreatedBy: p.CreatedBy,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update project access")
		return
	}
	resp := h.projectForRequest(r, p)
	workspaceID := uuidToString(wsID)
	// Recipients that retain access receive the normal full update. Recipients
	// that lost access receive only the opaque access_changed invalidation; the
	// realtime authorizer routes the two events to mutually exclusive audiences.
	h.publish(protocol.EventProjectUpdated, workspaceID, "member", userID, map[string]any{"project": projectEventResponse(resp)})
	h.publish(protocol.EventProjectAccessChanged, workspaceID, "member", userID, map[string]any{"project_id": resp.ID})
	writeJSON(w, http.StatusOK, resp)
}

// Per-request permissions cannot be broadcast as if every recipient were the
// actor. Clients reload their own access policy from the project endpoint.
func projectEventResponse(resp ProjectResponse) map[string]any {
	encoded, _ := json.Marshal(resp)
	payload := map[string]any{}
	_ = json.Unmarshal(encoded, &payload)
	delete(payload, "access_allowed")
	delete(payload, "can_manage_access")
	delete(payload, "allowed_user_ids")
	return payload
}
