package realtime

import "testing"

func TestScopeBroadcastCombinesProjectAccessAndExcludedUser(t *testing.T) {
	for _, tc := range []struct {
		name, excluded        string
		authorized, delivered bool
	}{
		{name: "authorized recipient", authorized: true, delivered: true},
		{name: "restricted recipient"},
		{name: "excluded authorized recipient", authorized: true, excluded: "recipient"},
		{name: "different excluded user", authorized: true, excluded: "other", delivered: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hub := NewHub()
			client := &Client{userID: "recipient", workspaceID: "workspace", send: make(chan []byte, 1)}
			hub.rooms[sk(ScopeWorkspace, "workspace")] = map[*Client]bool{client: true}
			hub.SetMessageAuthorizer(func(_, _, _, _ string, _ []byte) bool { return tc.authorized })
			hub.broadcastToScope(ScopeWorkspace, "workspace", []byte(`{"type":"project:updated"}`), "", tc.excluded)
			if delivered := len(client.send) == 1; delivered != tc.delivered {
				t.Fatalf("delivered = %v, want %v", delivered, tc.delivered)
			}
		})
	}
}
