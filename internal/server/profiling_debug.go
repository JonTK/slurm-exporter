package server

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/performance"
	"github.com/sirupsen/logrus"
)

// ProfilingDebugHandler handles profiling debug endpoints
type ProfilingDebugHandler struct {
	profiler *performance.Profiler
	manager  *collector.ProfiledCollectorManager
	logger   *logrus.Entry
}

// NewProfilingDebugHandler creates a new profiling debug handler
func NewProfilingDebugHandler(
	profiler *performance.Profiler,
	manager *collector.ProfiledCollectorManager,
	logger *logrus.Entry,
) *ProfilingDebugHandler {
	return &ProfilingDebugHandler{
		profiler: profiler,
		manager:  manager,
		logger:   logger,
	}
}

// RegisterRoutes registers profiling debug routes
func (h *ProfilingDebugHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/debug/profiling", h.handleProfilingStatus).Methods("GET")
	router.HandleFunc("/debug/profiling/list", h.handleListProfiles).Methods("GET")
	router.HandleFunc("/debug/profiling/profile/{id}", h.handleGetProfile).Methods("GET")
	router.HandleFunc("/debug/profiling/collector/{name}/enable", h.handleEnableProfiling).Methods("POST")
	router.HandleFunc("/debug/profiling/collector/{name}/disable", h.handleDisableProfiling).Methods("POST")
	router.HandleFunc("/debug/profiling/download/{id}", h.handleDownloadProfile).Methods("GET")
	router.HandleFunc("/debug/profiling/stats", h.handleStats).Methods("GET")
}

// handleProfilingStatus shows profiling status
func (h *ProfilingDebugHandler) handleProfilingStatus(w http.ResponseWriter, r *http.Request) {
	acceptJSON := r.Header.Get("Accept") == "application/json"

	stats := h.manager.GetStats()

	if acceptJSON {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			h.logger.WithError(err).Error("Failed to encode profiling stats")
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// HTML response
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Profiling Status - SLURM Exporter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .enabled { color: green; }
        .disabled { color: red; }
        button { margin: 5px; padding: 5px 10px; }
        .actions { white-space: nowrap; }
    </style>
</head>
<body>
    <h1>Profiling Status</h1>
    
    <h2>Collectors</h2>
    <table>
        <tr>
            <th>Collector</th>
            <th>Collector Enabled</th>
            <th>Profiling Enabled</th>
            <th>Actions</th>
        </tr>
        {{range $name, $stats := .collectors}}
        <tr>
            <td>{{$name}}</td>
            <td class="{{if $stats.collector_enabled}}enabled{{else}}disabled{{end}}">
                {{if $stats.collector_enabled}}Yes{{else}}No{{end}}
            </td>
            <td class="{{if $stats.profiling_enabled}}enabled{{else}}disabled{{end}}">
                {{if $stats.profiling_enabled}}Yes{{else}}No{{end}}
            </td>
            <td class="actions">
                {{if $stats.profiling_enabled}}
                <button onclick="toggleProfiling('{{$name}}', false)">Disable Profiling</button>
                {{else}}
                <button onclick="toggleProfiling('{{$name}}', true)">Enable Profiling</button>
                {{end}}
                <button onclick="viewProfiles('{{$name}}')">View Profiles</button>
            </td>
        </tr>
        {{end}}
    </table>

    <h2>Profiler Configuration</h2>
    <pre>{{.profiler_stats}}</pre>

    <script>
    function toggleProfiling(collector, enable) {
        fetch('/debug/profiling/collector/' + collector + '/' + (enable ? 'enable' : 'disable'), {
            method: 'POST'
        }).then(() => location.reload());
    }
    
    function viewProfiles(collector) {
        window.location.href = '/debug/profiling/list?collector=' + collector;
    }
    </script>
</body>
</html>
`

	t, err := template.New("profiling").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert stats for template
	data := map[string]interface{}{
		"collectors":     stats["collectors"],
		"profiler_stats": fmt.Sprintf("%+v", stats["profiler_stats"]),
	}

	w.Header().Set("Content-Type", "text/html")
	_ = t.Execute(w, data)
}

// handleListProfiles lists available profiles
func (h *ProfilingDebugHandler) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	collectorFilter := r.URL.Query().Get("collector")
	acceptJSON := r.Header.Get("Accept") == "application/json"

	var profiles []*performance.ProfileMetadata
	var err error

	if collectorFilter != "" {
		profiles, err = h.manager.GetCollectorProfiles(collectorFilter)
	} else {
		allProfiles, err := h.manager.GetAllProfiles()
		if err == nil {
			for _, collectorProfiles := range allProfiles {
				profiles = append(profiles, collectorProfiles...)
			}
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if acceptJSON {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profiles); err != nil {
			h.logger.WithError(err).Error("Failed to encode profile list")
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// HTML response
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Profiles - SLURM Exporter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .actions { white-space: nowrap; }
        button { margin: 2px; }
    </style>
</head>
<body>
    <h1>Profiles {{if .collector}}for {{.collector}}{{end}}</h1>
    
    <a href="/debug/profiling">← Back to Status</a>
    
    <table>
        <tr>
            <th>ID</th>
            <th>Collector</th>
            <th>Start Time</th>
            <th>Duration</th>
            <th>Size</th>
            <th>Actions</th>
        </tr>
        {{range .profiles}}
        <tr>
            <td>{{.ID}}</td>
            <td>{{.CollectorName}}</td>
            <td>{{.StartTime.Format "2006-01-02 15:04:05"}}</td>
            <td>{{.Duration}}</td>
            <td>{{.Size}} bytes</td>
            <td class="actions">
                <button onclick="viewProfile('{{.ID}}')">View</button>
                <button onclick="downloadProfile('{{.ID}}')">Download</button>
            </td>
        </tr>
        {{end}}
    </table>

    <script>
    function viewProfile(id) {
        window.location.href = '/debug/profiling/profile/' + id;
    }
    
    function downloadProfile(id) {
        window.location.href = '/debug/profiling/download/' + id;
    }
    </script>
</body>
</html>
`

	t, err := template.New("profiles").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"profiles":  profiles,
		"collector": collectorFilter,
	}

	w.Header().Set("Content-Type", "text/html")
	_ = t.Execute(w, data)
}

// handleGetProfile shows profile details
func (h *ProfilingDebugHandler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	profile, err := h.profiler.LoadProfile(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Profile not found: %v", err), http.StatusNotFound)
		return
	}

	acceptJSON := r.Header.Get("Accept") == "application/json"
	if acceptJSON {
		w.Header().Set("Content-Type", "application/json")
		// Convert profile to JSON-safe format
		data := map[string]interface{}{
			"collector_name": profile.CollectorName,
			"start_time":     profile.StartTime,
			"end_time":       profile.EndTime,
			"duration":       profile.Duration,
			"phases":         profile.Phases,
			"metadata":       profile.Metadata,
		}
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.WithError(err).Error("Failed to encode profile data")
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// HTML response
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Profile {{.ID}} - SLURM Exporter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .info { margin: 10px 0; }
        .info label { font-weight: bold; }
        pre { background: #f5f5f5; padding: 10px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>Profile Details</h1>
    
    <a href="/debug/profiling/list">← Back to List</a>
    
    <div class="info">
        <label>ID:</label> {{.ID}}<br>
        <label>Collector:</label> {{.profile.CollectorName}}<br>
        <label>Start Time:</label> {{.profile.StartTime.Format "2006-01-02 15:04:05"}}<br>
        <label>End Time:</label> {{.profile.EndTime.Format "2006-01-02 15:04:05"}}<br>
        <label>Duration:</label> {{.profile.Duration}}<br>
    </div>

    <h2>Phases</h2>
    <table>
        <tr>
            <th>Phase</th>
            <th>Duration</th>
            <th>Allocations</th>
        </tr>
        {{range $name, $phase := .profile.Phases}}
        <tr>
            <td>{{$name}}</td>
            <td>{{$phase.Duration}}</td>
            <td>{{$phase.Allocations}} bytes</td>
        </tr>
        {{end}}
    </table>

    <h2>Metadata</h2>
    <pre>{{.metadata}}</pre>

    <h2>Available Profiles</h2>
    <ul>
        {{if .profile.CPUProfile}}<li>CPU Profile ({{.cpuSize}} bytes)</li>{{end}}
        {{if .profile.HeapProfile}}<li>Heap Profile ({{.heapSize}} bytes)</li>{{end}}
        {{if .profile.GoroutineProfile}}<li>Goroutine Profile ({{.goroutineSize}} bytes)</li>{{end}}
        {{if .profile.BlockProfile}}<li>Block Profile ({{.blockSize}} bytes)</li>{{end}}
        {{if .profile.MutexProfile}}<li>Mutex Profile ({{.mutexSize}} bytes)</li>{{end}}
        {{if .profile.TraceData}}<li>Trace Data ({{.traceSize}} bytes)</li>{{end}}
    </ul>

    <button onclick="downloadProfile('{{.ID}}')">Download All Profiles</button>

    <script>
    function downloadProfile(id) {
        window.location.href = '/debug/profiling/download/' + id;
    }
    </script>
</body>
</html>
`

	t, err := template.New("profile").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"ID":            id,
		"profile":       profile,
		"metadata":      fmt.Sprintf("%+v", profile.Metadata),
		"cpuSize":       0,
		"heapSize":      0,
		"goroutineSize": 0,
		"blockSize":     0,
		"mutexSize":     0,
		"traceSize":     0,
	}

	// Calculate sizes
	if profile.CPUProfile != nil {
		data["cpuSize"] = profile.CPUProfile.Len()
	}
	if profile.HeapProfile != nil {
		data["heapSize"] = profile.HeapProfile.Len()
	}
	if profile.GoroutineProfile != nil {
		data["goroutineSize"] = profile.GoroutineProfile.Len()
	}
	if profile.BlockProfile != nil {
		data["blockSize"] = profile.BlockProfile.Len()
	}
	if profile.MutexProfile != nil {
		data["mutexSize"] = profile.MutexProfile.Len()
	}
	if profile.TraceData != nil {
		data["traceSize"] = profile.TraceData.Len()
	}

	w.Header().Set("Content-Type", "text/html")
	_ = t.Execute(w, data)
}

// handleEnableProfiling enables profiling for a collector
func (h *ProfilingDebugHandler) handleEnableProfiling(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	err := h.manager.SetProfilingEnabled(name, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Profiling enabled"))
}

// handleDisableProfiling disables profiling for a collector
func (h *ProfilingDebugHandler) handleDisableProfiling(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	err := h.manager.SetProfilingEnabled(name, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Profiling disabled"))
}

// handleDownloadProfile downloads a profile as a zip file
func (h *ProfilingDebugHandler) handleDownloadProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	profile, err := h.profiler.LoadProfile(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Profile not found: %v", err), http.StatusNotFound)
		return
	}

	// Create zip file in memory
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add metadata
	metadata := fmt.Sprintf("Profile ID: %s\nCollector: %s\nStart: %s\nEnd: %s\nDuration: %s\n\nMetadata:\n%+v",
		id, profile.CollectorName, profile.StartTime, profile.EndTime, profile.Duration, profile.Metadata)

	metaFile, _ := zipWriter.Create("metadata.txt")
	_, _ = metaFile.Write([]byte(metadata))

	// Add CPU profile
	if profile.CPUProfile != nil && profile.CPUProfile.Len() > 0 {
		cpuFile, _ := zipWriter.Create("cpu.pprof")
		_, _ = io.Copy(cpuFile, profile.CPUProfile)
	}

	// Add heap profile
	if profile.HeapProfile != nil && profile.HeapProfile.Len() > 0 {
		heapFile, _ := zipWriter.Create("heap.pprof")
		_, _ = io.Copy(heapFile, profile.HeapProfile)
	}

	// Add goroutine profile
	if profile.GoroutineProfile != nil && profile.GoroutineProfile.Len() > 0 {
		goroutineFile, _ := zipWriter.Create("goroutine.pprof")
		_, _ = io.Copy(goroutineFile, profile.GoroutineProfile)
	}

	// Add block profile
	if profile.BlockProfile != nil && profile.BlockProfile.Len() > 0 {
		blockFile, _ := zipWriter.Create("block.pprof")
		_, _ = io.Copy(blockFile, profile.BlockProfile)
	}

	// Add mutex profile
	if profile.MutexProfile != nil && profile.MutexProfile.Len() > 0 {
		mutexFile, _ := zipWriter.Create("mutex.pprof")
		_, _ = io.Copy(mutexFile, profile.MutexProfile)
	}

	// Add trace data
	if profile.TraceData != nil && profile.TraceData.Len() > 0 {
		traceFile, _ := zipWriter.Create("trace.out")
		_, _ = io.Copy(traceFile, profile.TraceData)
	}

	// Add phases report
	phasesReport := "Phases Report\n\n"
	for name, phase := range profile.Phases {
		phasesReport += fmt.Sprintf("%s:\n  Duration: %v\n  Allocations: %d bytes\n\n",
			name, phase.Duration, phase.Allocations)
	}
	phasesFile, _ := zipWriter.Create("phases.txt")
	_, _ = phasesFile.Write([]byte(phasesReport))

	_ = zipWriter.Close()

	// Send zip file
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"profile_%s.zip\"", id))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	if _, err := w.Write(buf.Bytes()); err != nil {
		h.logger.WithError(err).Error("Failed to write profile zip")
	}
}

// handleStats shows profiling statistics
func (h *ProfilingDebugHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := h.profiler.GetStats()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.WithError(err).Error("Failed to encode profiler stats")
	}
}
