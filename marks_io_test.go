package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupMarksTest(t *testing.T, marks []FileMarks) {
	t.Helper()
	t.Chdir(t.TempDir())
	os.MkdirAll(".bfr", 0755)
	data, _ := json.MarshalIndent(marks, "", "  ")
	os.WriteFile(filepath.Join(".bfr", "bfr.json"), data, 0644)
}

func TestInitOrUpdateMarks_NoWriteWhenUnchanged(t *testing.T) {
	marks := []FileMarks{
		{
			Path:     "a.go",
			FileName: "a.go",
			Commit:   "abc123",
			Reviewers: map[string][]Segment{
				"Test": {mkSeg(1, 3, StateUnreviewed)},
			},
			ImportanceSegments: []ImportanceSegment{mkISeg(1, 3, ImportanceMedium)},
		},
	}
	setupMarksTest(t, marks)

	// Create the file so it can be read
	os.WriteFile("a.go", []byte("line1\nline2\nline3"), 0644)

	entries := []fileEntry{{relPath: "a.go", name: "a.go"}}
	result, changed, err := initOrUpdateMarks(entries, "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("expected changed=false when commit and files are the same")
	}
	if len(result) != 1 {
		t.Errorf("expected 1 mark, got %d", len(result))
	}
}

func TestInitOrUpdateMarks_WritesWhenNewFile(t *testing.T) {
	marks := []FileMarks{
		{
			Path:     "a.go",
			FileName: "a.go",
			Commit:   "abc123",
			Reviewers: map[string][]Segment{
				"Test": {mkSeg(1, 3, StateUnreviewed)},
			},
			ImportanceSegments: []ImportanceSegment{mkISeg(1, 3, ImportanceMedium)},
		},
	}
	setupMarksTest(t, marks)

	os.WriteFile("a.go", []byte("line1\nline2\nline3"), 0644)
	os.WriteFile("b.go", []byte("line1\nline2"), 0644)

	entries := []fileEntry{
		{relPath: "a.go", name: "a.go"},
		{relPath: "b.go", name: "b.go"},
	}
	result, changed, err := initOrUpdateMarks(entries, "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("expected changed=true when new file added")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 marks, got %d", len(result))
	}
}

func TestInitOrUpdateMarks_WritesWhenCommitDiffers(t *testing.T) {
	marks := []FileMarks{
		{
			Path:     "a.go",
			FileName: "a.go",
			Commit:   "old123",
			Reviewers: map[string][]Segment{
				"Test": {mkSeg(1, 3, StateUnreviewed)},
			},
			ImportanceSegments: []ImportanceSegment{mkISeg(1, 3, ImportanceMedium)},
		},
	}
	setupMarksTest(t, marks)

	os.WriteFile("a.go", []byte("line1\nline2\nline3"), 0644)

	entries := []fileEntry{{relPath: "a.go", name: "a.go"}}
	_, changed, err := initOrUpdateMarks(entries, "new456")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("expected changed=true when commit differs")
	}
}

func TestDeleteCommentsByID_Single(t *testing.T) {
	marks := []FileMarks{
		{
			Path: "a.go",
			Comments: []Comment{
				{ID: "aaa", Text: "first"},
				{ID: "bbb", Text: "second"},
			},
		},
	}
	count := deleteCommentsByID(marks, []string{"aaa"})
	if count != 1 {
		t.Errorf("expected 1 deleted, got %d", count)
	}
	if len(marks[0].Comments) != 1 {
		t.Errorf("expected 1 comment remaining, got %d", len(marks[0].Comments))
	}
	if marks[0].Comments[0].ID != "bbb" {
		t.Errorf("wrong comment remaining: %s", marks[0].Comments[0].ID)
	}
}

func TestDeleteCommentsByID_Multiple(t *testing.T) {
	marks := []FileMarks{
		{
			Path: "a.go",
			Comments: []Comment{
				{ID: "aaa", Text: "first"},
				{ID: "bbb", Text: "second"},
				{ID: "ccc", Text: "third"},
			},
		},
	}
	count := deleteCommentsByID(marks, []string{"aaa", "ccc"})
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}
	if len(marks[0].Comments) != 1 || marks[0].Comments[0].ID != "bbb" {
		t.Errorf("expected only bbb remaining, got %v", marks[0].Comments)
	}
}

func TestDeleteCommentsByID_NotFound(t *testing.T) {
	marks := []FileMarks{
		{
			Path:     "a.go",
			Comments: []Comment{{ID: "aaa", Text: "first"}},
		},
	}
	count := deleteCommentsByID(marks, []string{"zzz"})
	if count != 0 {
		t.Errorf("expected 0 deleted, got %d", count)
	}
	if len(marks[0].Comments) != 1 {
		t.Errorf("should not have removed anything")
	}
}

func TestDeleteCommentsByID_AcrossFiles(t *testing.T) {
	marks := []FileMarks{
		{Path: "a.go", Comments: []Comment{{ID: "aaa", Text: "in a"}}},
		{Path: "b.go", Comments: []Comment{{ID: "bbb", Text: "in b"}}},
	}
	count := deleteCommentsByID(marks, []string{"aaa", "bbb"})
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}
	if len(marks[0].Comments) != 0 || len(marks[1].Comments) != 0 {
		t.Error("all comments should be removed")
	}
}
