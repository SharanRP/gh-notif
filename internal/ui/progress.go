package ui

// ProgressUpdateMsg is a message for updating progress
type ProgressUpdateMsg struct {
	Completed int
	Total     int
	Errors    int
}
