package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	get := func(user string, want int) ProjectResponse {
		t.Helper()
		var result ProjectResponse
		response := testutil.Call(t, testHandler.GetProject, withURLParam(newRequestAs(user, "GET", "/api/projects/"+project.ID, nil), "id", project.ID)).Want(want)
		if want == http.StatusOK {
			response.JSON(&result)
		}
		return result
	}
	if !get(member, http.StatusOK).AccessAllowed {
		t.Fatal("existing unrestricted project must remain readable")
	}
	testutil.Call(t, testHandler.CreatePin, newRequestAs(member, "POST", "/api/pins", map[string]any{"item_type": "project", "item_id": project.ID})).Want(http.StatusCreated)
	testutil.Call(t, testHandler.CreatePin, newRequestAs(member, "POST", "/api/pins", map[string]any{"item_type": "issue", "item_id": issueID})).Want(http.StatusCreated)
	setAccess(testUserID, true, []string{}, 200)
	get(member, http.StatusNotFound)
	// A validated task token acts for the runtime in its bound workspace, not
	// for its owner's project grants. Member sessions remain restricted.
	runtimeRequest := func(method, path string, body any) *http.Request {
		t.Helper()
		req := newRequestAs(member, method, path, body)
		req.Header.Set("X-Actor-Source", "task_token")
		return req
	}
	runtimeProject := testutil.Call(t, testHandler.GetProject, withURLParam(runtimeRequest("GET", "/api/projects/"+project.ID, nil), "id", project.ID)).Want(http.StatusOK)
	var runtimeProjectResponse ProjectResponse
	runtimeProject.JSON(&runtimeProjectResponse)
	if !runtimeProjectResponse.AccessAllowed || runtimeProjectResponse.CanManageAccess {
		t.Fatal("runtime must read restricted project without gaining access management")
	}
	testutil.Call(t, testHandler.GetIssue, withURLParam(runtimeRequest("GET", "/api/issues/"+issueID, nil), "id", issueID)).Want(http.StatusOK)
	runtimeIssues := testutil.Call(t, testHandler.ListIssues, runtimeRequest("GET", "/api/issues?open_only=true", nil)).Want(http.StatusOK)
	if !strings.Contains(runtimeIssues.Body.String(), issueID) {
		t.Fatal("runtime issue list omitted its restricted issue")
	}
	runtimeProjects := testutil.Call(t, testHandler.ListProjects, runtimeRequest("GET", "/api/projects", nil)).Want(http.StatusOK)
	if !strings.Contains(runtimeProjects.Body.String(), project.ID) {
		t.Fatal("runtime project list omitted the restricted project")
	}
	runtimeSearch := testutil.Call(t, testHandler.SearchIssues, runtimeRequest("GET", "/api/issues/search?q=private%20issue%20needle", nil)).Want(http.StatusOK)
	if !strings.Contains(runtimeSearch.Body.String(), issueID) {
		t.Fatal("runtime issue search omitted the restricted issue")
	}
	setAccess(member, false, nil, http.StatusNotFound)
	testutil.Call(t, testHandler.UpdateProjectAccess, withURLParam(runtimeRequest("PUT", "/api/projects/"+project.ID+"/access", map[string]any{"access_restricted": false}), "id", project.ID)).Want(http.StatusForbidden)
	runtimeID := dbfx.Runtime(t, "restricted project runtime")
	agentID := dbfx.Agent(t, "restricted project agent", runtimeID)
	taskID := dbfx.Task(t, agentID, testutil.Cols{"runtime_id": runtimeID, "issue_id": issueID})
	daemonResponse := httptest.NewRecorder()
	_, _, daemonAllowed := testHandler.requireDaemonTaskAccessWithWorkspace(daemonResponse, newRequestAs(member, "GET", "/api/daemon/tasks/"+taskID, nil), taskID)
	if !daemonAllowed {
		t.Fatalf("runtime owner must access its workspace task despite project grants: %d %s", daemonResponse.Code, daemonResponse.Body.String())
	}
	listResult := testutil.Call(t, testHandler.ListProjects, newRequestAs(member, "GET", "/api/projects", nil)).Want(http.StatusOK)
	if strings.Contains(listResult.Body.String(), project.ID) || strings.Contains(listResult.Body.String(), project.Title) {
		t.Fatal("restricted project remained discoverable in the project list")
	}
	searchResult := testutil.Call(t, testHandler.SearchProjects, newRequestAs(member, "GET", "/api/projects/search?q=ACL%20protected", nil)).Want(http.StatusOK)
	if strings.Contains(searchResult.Body.String(), project.ID) || strings.Contains(searchResult.Body.String(), project.Title) {
		t.Fatal("restricted project remained discoverable in project search")
	}
	pinsResult := testutil.Call(t, testHandler.ListPins, newRequestAs(member, "GET", "/api/pins", nil)).Want(http.StatusOK)
	if strings.Contains(pinsResult.Body.String(), project.ID) || strings.Contains(pinsResult.Body.String(), issueID) {
		t.Fatal("restricted project remained discoverable through saved pins")
	}
	testutil.Call(t, testHandler.CreatePin, newRequestAs(member, "POST", "/api/pins", map[string]any{"item_type": "project", "item_id": project.ID})).Want(http.StatusNotFound)
	testutil.Call(t, testHandler.CreatePin, newRequestAs(member, "POST", "/api/pins", map[string]any{"item_type": "issue", "item_id": issueID})).Want(http.StatusNotFound)
	if !get(testUserID, http.StatusOK).AccessAllowed {
		t.Fatal("creator locked themselves out")
	}
	get(admin, http.StatusNotFound)
	setAccess(member, false, nil, http.StatusNotFound)
	setAccess(admin, false, nil, http.StatusNotFound)
	testutil.Call(t, testHandler.UpdateProject, withURLParam(newRequestAs(member, "PUT", "/api/projects/"+project.ID, map[string]any{"title": "unauthorized edit"}), "id", project.ID)).Want(http.StatusNotFound)
	testutil.Call(t, testHandler.ListProjectResources, withURLParam(newRequestAs(member, "GET", "/api/projects/"+project.ID+"/resources", nil), "id", project.ID)).Want(http.StatusNotFound)
	testutil.Call(t, testHandler.GetIssue, withURLParam(newRequestAs(member, "GET", "/api/issues/"+issueID, nil), "id", issueID)).Want(http.StatusNotFound)
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
	table := issueTableRowsRequest{Query: issueTableQuerySpec{Scope: issueTableScope{Kind: "workspace"}}, Group: issueTableGroupSpec{Kind: "none"}, Page: issueTablePageRequest{Limit: 100}}
	tableResult := testutil.Call(t, testHandler.ListIssueTableRows, newRequestAs(member, "POST", "/api/issues/table/rows", table)).Want(200)
	if strings.Contains(tableResult.Body.String(), issueID) {
		t.Fatal("restricted issue leaked through table rows")
	}
	event, _ := json.Marshal(map[string]any{"type": "issue:updated", "payload": map[string]any{"issue_id": issueID}})
	if testHandler.AuthorizeProjectMessage(member, testWorkspaceID, "workspace", testWorkspaceID, event) {
		t.Fatal("websocket disclosed restricted issue")
	}
	accessChanged, _ := json.Marshal(map[string]any{"type": "project:access_changed", "payload": map[string]any{"project_id": project.ID}})
	if !testHandler.AuthorizeProjectMessage(member, testWorkspaceID, "workspace", testWorkspaceID, accessChanged) {
		t.Fatal("revoked member did not receive opaque access invalidation")
	}
	if testHandler.AuthorizeProjectMessage(testUserID, testWorkspaceID, "workspace", testWorkspaceID, accessChanged) {
		t.Fatal("authorized creator received revoked-access invalidation")
	}
	setAccess(testUserID, true, []string{member, member}, 200)
	if !get(member, http.StatusOK).AccessAllowed || get(member, http.StatusOK).CanManageAccess {
		t.Fatal("selected member access incorrect")
	}
	setAccess(member, false, nil, http.StatusForbidden)
	testutil.Call(t, testHandler.GetIssue, withURLParam(newRequestAs(member, "GET", "/api/issues/"+issueID, nil), "id", issueID)).Want(200)
	if !testHandler.AuthorizeProjectMessage(member, testWorkspaceID, "workspace", testWorkspaceID, event) {
		t.Fatal("selected member websocket denied")
	}
	if testHandler.AuthorizeProjectMessage(member, testWorkspaceID, "workspace", testWorkspaceID, accessChanged) {
		t.Fatal("selected member received revoked-access invalidation")
	}
	setAccess(testUserID, true, nil, 200)
	get(member, http.StatusNotFound)
	if testHandler.AuthorizeProjectMessage(member, testWorkspaceID, "workspace", testWorkspaceID, event) {
		t.Fatal("websocket access survived revocation")
	}
	setAccess(testUserID, true, []string{"invalid"}, http.StatusBadRequest)
	setAccess(testUserID, true, []string{"00000000-0000-0000-0000-000000000001"}, http.StatusBadRequest)
	setAccess(testUserID, false, nil, 200)
	if !get(member, http.StatusOK).AccessAllowed {
		t.Fatal("unrestricted access not restored")
	}
	deleted, _ := json.Marshal(map[string]any{"type": "issue:deleted", "payload": map[string]any{"issue_id": "00000000-0000-0000-0000-000000000001"}})
	if !testHandler.AuthorizeProjectMessage(member, testWorkspaceID, "workspace", testWorkspaceID, deleted) {
		t.Fatal("opaque deletion invalidation must survive removal of its row")
	}
}
