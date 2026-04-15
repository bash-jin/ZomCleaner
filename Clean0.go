package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ── Data Structures ──────────────────────────────────────────────────────────

type Process struct {
	PID, PPID int
	Name      string
}

type SystemStats struct {
	MemTotal     string
	MemAvailable string
	Cached       string
	FragScore    string
}

type LogEntry struct {
	Time   string
	Action string
	Status string
	Class  string // For CSS coloring (success vs error)
}

type PageData struct {
	Zombies []Process
	Stats   SystemStats
	Logs    []LogEntry
	IsRoot  bool
}

// Global state for logs
var logs = []LogEntry{}

// ── Logic Functions ──────────────────────────────────────────────────────────

func addLog(action, status, class string) {
	entry := LogEntry{
		Time:   time.Now().Format("15:04:05"),
		Action: action,
		Status: status,
		Class:  class,
	}
	// Prepend to show newest first
	logs = append([]LogEntry{entry}, logs...)
	if len(logs) > 8 {
		logs = logs[:8]
	}
}

func getStats() SystemStats {
	data, _ := os.ReadFile("/proc/meminfo")
	stats := SystemStats{FragScore: "Optimal"}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 { continue }
		val := parts[1] + " KB" // Standardize units
		switch parts[0] {
		case "MemTotal:":     stats.MemTotal = val
		case "MemAvailable:": stats.MemAvailable = val
		case "Cached:":       stats.Cached = val
		}
	}
	// Check buddyinfo for memory holes
	buddy, _ := os.ReadFile("/proc/buddyinfo")
	if strings.Contains(string(buddy), " 0 0 0 ") {
		stats.FragScore = "Fragmented"
	}
	return stats
}

func scanZombies() []Process {
	entries, _ := os.ReadDir("/proc")
	var list []Process
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil { continue }
		statusPath := fmt.Sprintf("/proc/%d/status", pid)
		data, err := os.ReadFile(statusPath)
		if err != nil { continue }
		
		content := string(data)
		if strings.Contains(content, "State:\tZ (zombie)") {
			name := "unknown"
			ppid := 0
			lines := strings.Split(content, "\n")
			for _, l := range lines {
				if strings.HasPrefix(l, "Name:") { name = strings.TrimSpace(l[5:]) }
				if strings.HasPrefix(l, "PPid:") { ppid, _ = strconv.Atoi(strings.TrimSpace(l[5:])) }
			}
			list = append(list, Process{PID: pid, PPID: ppid, Name: name})
		}
	}
	return list
}

// ── Server ───────────────────────────────────────────────────────────────────

func main() {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Zombies: scanZombies(),
			Stats:   getStats(),
			Logs:    logs,
			IsRoot:  os.Geteuid() == 0,
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/reap", func(w http.ResponseWriter, r *http.Request) {
		z := scanZombies()
		if len(z) == 0 {
			addLog("Reaper", "No zombies found to reap", "warn")
		} else {
			for _, p := range z {
				if p.PPID > 1 {
					syscall.Kill(p.PPID, syscall.SIGCHLD)
				}
			}
			addLog("Reaper", fmt.Sprintf("Signals sent to %d parents", len(z)), "success")
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/clear-cache", func(w http.ResponseWriter, r *http.Request) {
		if os.Geteuid() != 0 {
			addLog("Cache", "FAILED: Root required", "error")
		} else {
			syscall.Sync()
			err := os.WriteFile("/proc/sys/vm/drop_caches", []byte("3\n"), 0644)
			if err != nil {
				addLog("Cache", "Error writing to proc", "error")
			} else {
				addLog("Cache", "PageCache and Inodes dropped", "success")
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/compact", func(w http.ResponseWriter, r *http.Request) {
		if os.Geteuid() != 0 {
			addLog("Defrag", "FAILED: Root required", "error")
		} else {
			err := os.WriteFile("/proc/sys/vm/compact_memory", []byte("1\n"), 0644)
			if err != nil {
				addLog("Defrag", "Kernel rejected compaction", "error")
			} else {
				addLog("Defrag", "Memory compaction triggered", "success")
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	fmt.Println("🧟 ZomCleaner Pro | http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// ── UI ───────────────────────────────────────────────────────────────────────

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>ZomCleaner Pro</title>
    <style>
        :root { --bg: #0d1117; --card: #161b22; --border: #30363d; --text: #c9d1d9; --blue: #58a6ff; --red: #f85149; --green: #3fb950; --warn: #d29922; }
        body { font-family: monospace; background: var(--bg); color: var(--text); padding: 20px; }
        .grid { display: grid; grid-template-columns: 350px 1fr; gap: 20px; }
        .card { background: var(--card); border: 1px solid var(--border); padding: 15px; border-radius: 6px; }
        .btn { display: block; width: 100%; padding: 10px; margin: 8px 0; cursor: pointer; background: #21262d; color: white; border: 1px solid var(--border); border-radius: 4px; font-weight: bold; }
        .btn:hover { background: #30363d; }
        .log-item { font-size: 12px; padding: 5px 0; border-bottom: 1px solid #222; }
        .success { color: var(--green); } .error { color: var(--red); } .warn { color: var(--warn); }
        table { width: 100%; border-collapse: collapse; margin-top: 10px; }
        th { text-align: left; color: #8b949e; border-bottom: 1px solid var(--border); padding: 8px; }
        td { padding: 8px; border-bottom: 1px solid #222; }
    </style>
</head>
<body>
    <h1>🧟 ZomCleaner Pro <small style="font-size: 14px; color: {{if .IsRoot}}var(--green){{else}}var(--red){{end}}">[{{if .IsRoot}}ROOT{{else}}USER{{end}}]</small></h1>
    <div class="grid">
        <div>
            <div class="card">
                <h3>System Health</h3>
                Avail: <b class="success">{{.Stats.MemAvailable}}</b><br>
                Cache: <b class="success">{{.Stats.Cached}}</b><br>
                Fragmentation: <b class="success">{{.Stats.FragScore}}</b>
            </div>
            <div class="card" style="margin-top: 20px;">
                <h3>Controls</h3>
                <form action="/reap" method="POST"><button class="btn">💀 REAP ZOMBIES</button></form>
                <form action="/clear-cache" method="POST"><button class="btn">🧹 FLUSH CACHE</button></form>
                <form action="/compact" method="POST"><button class="btn">💎 DEEP DEFRAG</button></form>
            </div>
        </div>
        <div class="card">
            <h3>Execution Logs</h3>
            {{range .Logs}}
            <div class="log-item">
                <span style="color:#8b949e">[{{.Time}}]</span> <b>{{.Action}}</b>: <span class="{{.Class}}">{{.Status}}</span>
            </div>
            {{else}}
            <div style="color:#555">Waiting for input...</div>
            {{end}}
        </div>
    </div>
    <div class="card" style="margin-top: 20px;">
        <h3>Zombie Processes</h3>
        <table>
            <thead><tr><th>PID</th><th>Process Name</th><th>Parent PID</th></tr></thead>
            <tbody>
                {{range .Zombies}}
                <tr><td>{{.PID}}</td><td class="error">{{.Name}}</td><td>{{.PPID}}</td></tr>
                {{else}}
                <tr><td colspan="3" style="text-align:center; padding:30px; color:var(--green)">No Zombies Found</td></tr>
                {{end}}
            </tbody>
        </table>
    </div>
</body>
</html>
`
