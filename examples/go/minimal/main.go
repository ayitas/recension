package main

import (
	"os"

	recension "github.com/ayitas/recension/sdk/go"
)

type student struct {
	Username string
	FullName string
	GPA      float64
}

func findStudent(username string) student {
	directory := map[string]student{
		"alice":   {Username: "alice", FullName: "Alice Anderson", GPA: 3.9},
		"bob":     {Username: "bob", FullName: "Bob Brown", GPA: 3.2},
		"charlie": {Username: "charlie", FullName: "Charlie Clark", GPA: 3.5},
	}
	return directory[username]
}

func main() {
	recension.Workflow("students", func(username string) {
		recension.StartTimer("find_student")
		s := findStudent(username)
		recension.StopTimer("find_student")
		recension.Assume("username", s.Username)
		recension.Check("fullname", s.FullName)
		recension.Check("gpa", s.GPA)
	}, recension.WithTestcases([]string{"alice", "bob", "charlie"}))

	os.Exit(recension.Run())
}
