package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/multica-ai/multica/server/internal/testutil"
)

func TestProjectAccessLifecycle(t *testing.T) {
	project := createProjectPermissionTestProject(t, "ACL protected project")
	member := createProjectPermissionTestMember(t, "member")
	admin := createProjectPermissionTestMember(t, "admin")
	issueID := dbfx.Issue(t, "private issue needle", testutil.Cols{"project_id": project.ID})
	_, err := testPool.Exec(context.Background(), "UPDATE project SET description = 'private description' WHERE id = $1", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	setAccess := func(user string, restricted bool, ids []string, want int) {
		t.Helper()
		req := withURLParam(newRequestAs(user, "PUT", "/api/projects/"+project.ID+"/access", map[string]any{"access_restricted": restricted, "allowed_user_ids": ids}), "id", project.ID)
		testutil.Call(t, testHandler.UpdateProjectAccess, req).Want(want)
	}
	get := func(user string) ProjectResponse {
		t.Helper()
		var result ProjectResponse
		testutil.Call(t, testHandler.GetProject, withURLParam(newRequestAs(user, "GET", "/api/projects/"+project.ID, nil), "id", project.ID)).Want(200).JSON(&result)
		return result
	}
	if !get(member).AccessAllowed {
		t.Fatal("existing unrestricted project must remain readable")
	}
	setAccess(testUserID, true, []string{}, 200)
	denied := get(member)
	if denied.AccessAllowed || denied.Description != nil || denied.CanManageAccess || denied.ResourceCount != 0 || len(denied.AllowedUserIDs) != 0 {
		t.Fatalf("denied project leaked content: %+v", denied)
	}
	if denied.CreatedBy == nil || *denied.CreatedBy != testUserID {
		t.Fatal("missing authorization contact")
	}
	if !get(testUserID).AccessAllowed {
		t.Fatal("creator locked themselves out")
	}
	if get(admin).AccessAllowed {
		t.Fatal("unselected administrator bypassed project access")
	}
	setAccess(member, false, nil, 403)
	setAccess(admin, false, nil, 403)
	testutil.Call(t, testHandler.UpdateProject, withURLParam(newRequestAs(member, "PUT", "/api/projects/"+project.ID, map[string]any{"title": "unauthorized edit"}), "id", project.ID)).Want(403)
	testutil.Call(t, testHandler.ListProjectResources, withURLParam(newRequestAs(member, "GET", "/api/projects/"+project.ID+"/resources", nil), "id", project.ID)).Want(403)
	testutil.Call(t, testHandler.GetIssue, withURLParam(newRequestAs(member, "GET", "/api/issues/"+issueID, nil), "id", issueID)).Want(403)
	for _, path := range []string{"/api/issues", "/api/issues?open_only=true", "/api/issues/search?q=private%20issue%20needle"} {
		handler := testHandler.ListIssues
		if strings.Contains(path, "search") {
			handler = testHandler.SearchIssues
		}
		result := testutil.Call(t, handler, newRequestAs(member, "GET", path, nil)).Want(200)
		if strings.Contains(result.Body.String(), issueID) {
			t.Fatalf("restricted issue leaked at %s: %s", path, result.Body.String())
		}
	}
	table := issueTableRowsRequest{Query: issueTableQuerySpec{Scope: issueTableScope{Kind: "workspace"}}, Page: issueTablePageRequest{Limit: 100}}
	tableResult := testutil.Call(t, testHandler.ListIssueTableRows, newRequestAs(member, "POST", "/api/issues/table/rows", table)).Want(200)
	if strings.Contains(tableResult.Body.String(), issueID) {
		t.Fatal("restricted issue leaked through table rows")
	}
	event, _ := json.Marshal(map[string]any{"type": "issue:updated", "payload": map[string]any{"issue_id": issueID}})
	if testHandler.AuthorizeProjectMessage(member, testWorkspaceID, "workspace", testWorkspaceID, event) {
		t.Fatal("websocket disclosed restricted issue")
	}
	setAccess(testUserID, true, []string{member, member}, 200)
	if !get(member).AccessAllowed || get(member).CanManageAccess {
		t.Fatal("selected member access incorrect")
	}
	testutil.Call(t, testHandler.GetIssue, withURLParam(newRequestAs(member, "GET", "/api/issues/"+issueID, nil), "id", issueID)).Want(200)
	if !testHandler.AuthorizeProjectMessage(member, testWorkspaceID, "workspace", testWorkspaceID, event) {
		t.Fatal("selected member websocket denied")
	}
	setAccess(testUserID, true, nil, 200)
	if get(member).AccessAllowed {
		t.Fatal("revocation did not apply")
	}
	if testHandler.AuthorizeProjectMessage(member, testWorkspaceID, "workspace", testWorkspaceID, event) {
		t.Fatal("websocket access survived revocation")
	}
	setAccess(testUserID, true, []string{"invalid"}, http.StatusBadRequest)
	setAccess(testUserID, true, []string{"00000000-0000-0000-0000-000000000001"}, http.StatusBadRequest)
	setAccess(testUserID, false, nil, 200)
	if !get(member).AccessAllowed {
		t.Fatal("unrestricted access not restored")
	}
}
