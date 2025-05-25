package dto

type ViewTaskLogs struct {
	Logs       []string `json:"logs"`
	TotalLines int      `json:"total_lines"`
}

func ToViewTaskLogs(logs []string, totalLines int) *ViewTaskLogs {
	return &ViewTaskLogs{Logs: logs, TotalLines: totalLines}
}

type TaskLogFilter struct {
	From int `json:"from"`
	To   int `json:"to"`
}
