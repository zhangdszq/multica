package migrations

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAllVersionsSurvivesIncompleteRuntimeDirectory(t *testing.T) {
	expected, err := AllVersions()
	if err != nil || len(expected) == 0 {
		t.Fatalf("manifest: %v", err)
	}
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "migrations"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "migrations", "536_correct_historical_project_creators.up.sql"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	actual, err := AllVersions()
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Fatalf("runtime files changed manifest: %v", err)
	}
	found := false
	for _, v := range actual {
		if v == "499_agent_task_issue_snapshot" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing issue snapshot migration")
	}
}
