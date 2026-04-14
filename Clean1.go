package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ---------------- DATA STRUCTURES ----------------

type Process struct {
	PID  int
	Name string
}

type SystemStats struct {
	MemTotal     string
	MemAvailable string
}

type LogEntry struct {
	Time   string
	Action string
	Status string
	Class  string
}

type PageData struct {
	Processes []Process
	Stats     SystemStats
	Logs      []LogEntry
	OS        string
}

var logs []LogEntry

// ---------------- LOGGING ----------------

func addLog(action, status, class string) {
	entry := LogEntry{
		Time:   time.Now().Format("15:04:05"),
		Action: action,
		Status: status,
		Class:  class,
	}
	logs = append([]LogEntry{entry}, logs...)
	if len(logs) > 10 {
		logs = logs[:10]
	}
}

// ---------------- GET PROCESSES ----------------

func getProcesses() []Process {
	cmd := exec.Command("powershell", "-Command",
		"Get-Process | Select-Object Id,ProcessName | ConvertTo-Json")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var raw interface{}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil
	}

	var processes []Process

	switch data := raw.(type) {
	case []interface{}:
		for _, item := range data {
			m := item.(map[string]interface{})
			processes = append(processes, Process{
				PID:  int(m["Id"].(float64)),
				Name: m["ProcessName"].(string),
			})
		}
	case map[string]interface{}:
		processes = append(processes, Process{
			PID:  int(data["Id"].(float64)),
			Name: data["ProcessName"].(string),
		})
	}
	return processes
}

// ---------------- MEMORY STATS ----------------

func getSystemStats() SystemStats {
	cmd := exec.Command("powershell", "-Command",
		"Get-CimInstance Win32_OperatingSystem | Select-Object TotalVisibleMemorySize,FreePhysicalMemory | ConvertTo-Json")

	out, err := cmd.Output()
	if err != nil {
		return SystemStats{"N/A", "N/A"}
	}

	var mem map[string]interface{}
	if err := json.Unmarshal(out, &mem); err != nil {
		return SystemStats{"N/A", "N/A"}
	}

	total := int(mem["TotalVisibleMemorySize"].(float64)) / 1024
	free := int(mem["FreePhysicalMemory"].(float64)) / 1024

	return SystemStats{
		MemTotal:     fmt.Sprintf("%d MB", total),
		MemAvailable: fmt.Sprintf("%d MB", free),
	}
}

// ---------------- CLEANING FUNCTIONS ----------------

func cleanProcesses() {
	cmd := exec.Command("powershell", "-Command",
		"Get-Process | Where-Object {$_.Responding -eq $false} | Stop-Process -Force")
	err := cmd.Run()
	if err != nil {
		addLog("Process Cleaner", "Failed or requires admin rights", "error")
	} else {
		addLog("Process Cleaner", "Unresponsive processes terminated", "success")
	}
}

func clearTempFiles() {
	cmd := exec.Command("powershell", "-Command", `
		try {
			$paths = @(
				"$env:TEMP\*",
				"$env:LOCALAPPDATA\Temp\*"
			)

			foreach ($path in $paths) {
				if (Test-Path $path) {
					Remove-Item -Path $path -Recurse -Force -ErrorAction SilentlyContinue
				}
			}
			Write-Output "SUCCESS"
		} catch {
			Write-Output "FAILED"
		}
	`)

	output, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(output))

	if err != nil || result != "SUCCESS" {
		addLog("Temp Cleaner", "Failed to clear temporary files", "error")
	} else {
		addLog("Temp Cleaner", "Temporary files cleared successfully", "success")
	}
}

func main() {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Processes: getProcesses(),
			Stats:     getSystemStats(),
			Logs:      logs,
			OS:        runtime.GOOS,
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/clean", func(w http.ResponseWriter, r *http.Request) {
		cleanProcesses()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/clear-temp", func(w http.ResponseWriter, r *http.Request) {
		clearTempFiles()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	fmt.Println("ZomCleaner running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// ---------------- HTML TEMPLATE ----------------

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
<title>ZomCleaner</title>
<style>
body { font-family: Arial; background:#0d1117; color:#c9d1d9; padding:20px; }
.card { background:#161b22; padding:15px; margin:15px 0; border-radius:8px; }
button { padding:10px; margin:5px; background:#238636; color:white; border:none; cursor:pointer; }
table { width:100%; border-collapse:collapse; }
th, td { padding:8px; border-bottom:1px solid #30363d; }
</style>
</head>
<body>

<h1>ZomCleaner ({{.OS}})</h1>

<div class="card">
<h3>System Memory</h3>
<p>Total: {{.Stats.MemTotal}}</p>
<p>Available: {{.Stats.MemAvailable}}</p>
</div>

<div class="card">
<h3>Controls</h3>
<form action="/clean" method="POST">
<button>Clean Background Processes</button>
</form>
<form action="/clear-temp" method="POST">
<button>Clear Temporary Files</button>
</form>
</div>

<div class="card">
<h3>Running Processes</h3>
<table>
<tr><th>PID</th><th>Name</th></tr>
{{range .Processes}}
<tr><td>{{.PID}}</td><td>{{.Name}}</td></tr>
{{else}}
<tr><td colspan="2">No processes found</td></tr>
{{end}}
</table>
</div>

<div class="card">
<h3>Logs</h3>
{{range .Logs}}
<p>[{{.Time}}] {{.Action}} - {{.Status}}</p>
{{else}}
<p>No actions performed yet.</p>
{{end}}
</div>

</body>
</html>
`
