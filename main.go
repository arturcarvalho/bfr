package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Set via: go build -ldflags "-X main.BuildTime=$(date +%Y-%m-%dT%H:%M)"
var BuildTime string

func main() {
	showComments := flag.Bool("comments", false, "output comments as JSON")
	deleteComment := flag.String("delete-comment", "", "delete comments by comma-separated IDs")
	flag.Parse()

	if *deleteComment != "" {
		ids := strings.Split(*deleteComment, ",")
		marks, err := loadMarks()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		count := deleteCommentsByID(marks, ids)
		if count > 0 {
			if err := saveMarks(marks); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		fmt.Printf("Deleted %d comment(s).\n", count)
		return
	}

	if *showComments {
		printComments()
		return
	}

	m, err := newModel(detectDirtyPaths())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type commentOutput struct {
	ID        string `json:"id"`
	File      string `json:"file"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Text      string `json:"text"`
	Author    string `json:"author"`
	CreatedAt string `json:"createdAt"`
}

func printComments() {
	marks, err := loadMarks()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var out []commentOutput
	for _, fm := range marks {
		for _, c := range fm.Comments {
			out = append(out, commentOutput{
				ID:        c.ID,
				File:      fm.Path,
				StartLine: c.StartLine,
				EndLine:   c.EndLine,
				Text:      c.Text,
				Author:    c.Author,
				CreatedAt: c.CreatedAt,
			})
		}
	}
	if out == nil {
		out = []commentOutput{}
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
}
