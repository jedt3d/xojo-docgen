package main

import "testing"

func TestExcludeFolderSubtrees(t *testing.T) {
	items := []ManifestItem{
		{ItemType: "Class", Name: "DependencyClass", ID: 4, ParentID: 3},
		{ItemType: "Class", Name: "App", ID: 1, ParentID: 0},
		{ItemType: "Folder", Name: "Nested", ID: 3, ParentID: 2},
		{ItemType: "Folder", Name: "dependencies", ID: 2, ParentID: 0},
		{ItemType: "Class", Name: "Feature", ID: 5, ParentID: 0},
	}

	filtered, excludedCount := excludeFolderSubtrees(items, []string{"Dependencies"})

	if excludedCount != 3 {
		t.Fatalf("excludedCount = %d, want 3", excludedCount)
	}
	if len(filtered) != 2 {
		t.Fatalf("len(filtered) = %d, want 2", len(filtered))
	}
	if filtered[0].Name != "App" || filtered[1].Name != "Feature" {
		t.Fatalf("unexpected filtered items: %#v", filtered)
	}
}

func TestExcludeFolderSubtreesWithoutNamesKeepsItems(t *testing.T) {
	items := []ManifestItem{{ItemType: "Class", Name: "App", ID: 1}}

	filtered, excludedCount := excludeFolderSubtrees(items, nil)

	if excludedCount != 0 || len(filtered) != 1 || filtered[0].Name != "App" {
		t.Fatalf("unexpected result: count=%d items=%#v", excludedCount, filtered)
	}
}
