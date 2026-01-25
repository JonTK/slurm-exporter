// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package debug

// debugTemplates contains HTML templates for debug endpoints
const debugTemplates = `
{{define "index"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #007acc; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; }
        .status { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .status.enabled { background-color: #d1ecf1; color: #0c5460; }
        .status.disabled { background-color: #e2e3e5; color: #383d41; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { padding: 8px 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f8f9fa; font-weight: bold; }
        tr:hover { background-color: #f8f9fa; }
        .metric { background-color: #f8f9fa; padding: 10px; margin: 5px 0; border-radius: 4px; }
        .nav { margin-bottom: 20px; }
        .nav a { margin-right: 15px; padding: 8px 15px; background-color: #007acc; color: white; text-decoration: none; border-radius: 4px; }
        .nav a:hover { background-color: #005a99; }
        .json-link { float: right; }
        .json-link a { color: #007acc; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="nav">
            <a href="/debug">Home</a>
            <a href="/debug/collectors">Collectors</a>
            <a href="/debug/tracing">Tracing</a>
            <a href="/debug/patterns">Patterns</a>
            <a href="/debug/scheduler">Scheduler</a>
            <a href="/debug/health">Health</a>
            <a href="/debug/runtime">Runtime</a>
            <a href="/debug/config">Config</a>
            <a href="/debug/cache">Cache</a>
            <div class="json-link">
                <a href="?format=json">JSON</a>
            </div>
        </div>
        
        <h1>{{.Title}}</h1>

        <h2>System Overview</h2>
        <div class="metric">
            <strong>Start Time:</strong> {{.StartTime.Format "2006-01-02 15:04:05 MST"}}<br>
            <strong>Uptime:</strong> {{.Uptime}}<br>
            <strong>Debug Requests:</strong> {{.RequestCount}}
        </div>

        <h2>Available Endpoints</h2>
        <table>
            <tr><th>Endpoint</th><th>Description</th></tr>
            {{range .Endpoints}}
            <tr>
                <td><a href="/debug/{{.}}">/debug/{{.}}</a></td>
                <td>
                    {{if eq . "collectors"}}Collector status and metrics{{end}}
                    {{if eq . "tracing"}}OpenTelemetry tracing information{{end}}
                    {{if eq . "patterns"}}Smart filter learned patterns{{end}}
                    {{if eq . "scheduler"}}Adaptive scheduler state{{end}}
                    {{if eq . "health"}}Health check results{{end}}
                    {{if eq . "runtime"}}Go runtime statistics{{end}}
                    {{if eq . "config"}}Configuration information{{end}}
                    {{if eq . "cache"}}Intelligent cache statistics{{end}}
                </td>
            </tr>
            {{end}}
        </table>

        <h2>Component Status</h2>
        <table>
            <tr><th>Component</th><th>Status</th></tr>
            {{range $name, $available := .Components}}
            <tr>
                <td>{{$name}}</td>
                <td>
                    {{if $available}}
                        <span class="status enabled">Available</span>
                    {{else}}
                        <span class="status disabled">Not Available</span>
                    {{end}}
                </td>
            </tr>
            {{end}}
        </table>
    </div>
</body>
</html>
{{end}}

{{define "collectors"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #007acc; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; }
        .status { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .status.healthy { background-color: #d4edda; color: #155724; }
        .status.unhealthy { background-color: #f8d7da; color: #721c24; }
        .status.degraded { background-color: #fff3cd; color: #856404; }
        .status.enabled { background-color: #d1ecf1; color: #0c5460; }
        .status.disabled { background-color: #e2e3e5; color: #383d41; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { padding: 8px 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f8f9fa; font-weight: bold; }
        tr:hover { background-color: #f8f9fa; }
        .metric { background-color: #f8f9fa; padding: 10px; margin: 5px 0; border-radius: 4px; }
        .nav { margin-bottom: 20px; }
        .nav a { margin-right: 15px; padding: 8px 15px; background-color: #007acc; color: white; text-decoration: none; border-radius: 4px; }
        .nav a:hover { background-color: #005a99; }
        .json-link { float: right; }
        .json-link a { color: #007acc; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="nav">
            <a href="/debug">Home</a>
            <a href="/debug/collectors">Collectors</a>
            <a href="/debug/tracing">Tracing</a>
            <a href="/debug/patterns">Patterns</a>
            <a href="/debug/scheduler">Scheduler</a>
            <a href="/debug/health">Health</a>
            <a href="/debug/runtime">Runtime</a>
            <a href="/debug/config">Config</a>
            <a href="/debug/cache">Cache</a>
            <div class="json-link">
                <a href="?format=json">JSON</a>
            </div>
        </div>
        
        <h1>{{.Title}}</h1>

        <h2>Summary</h2>
        <div class="metric">
            <strong>Total Collectors:</strong> {{.Summary.total}}<br>
            <strong>Enabled:</strong> {{.Summary.enabled}}<br>
            <strong>With Errors:</strong> {{.Summary.errors}}
        </div>

        <h2>Collector States</h2>
        <table>
            <tr>
                <th>Name</th>
                <th>Status</th>
                <th>Last Collection</th>
                <th>Duration</th>
                <th>Consecutive Errors</th>
                <th>Total Collections</th>
                <th>Errors/Total</th>
                <th>Current Interval</th>
            </tr>
            {{range $name, $state := .States}}
            <tr>
                <td>{{$name}}</td>
                <td>
                    {{if $state.Enabled}}
                        {{if eq $state.ConsecutiveErrors 0}}
                            <span class="status healthy">Healthy</span>
                        {{else if lt $state.ConsecutiveErrors 3}}
                            <span class="status degraded">Degraded</span>
                        {{else}}
                            <span class="status unhealthy">Unhealthy</span>
                        {{end}}
                    {{else}}
                        <span class="status disabled">Disabled</span>
                    {{end}}
                </td>
                <td>{{if not $state.LastCollection.IsZero}}{{$state.LastCollection.Format "15:04:05"}}{{else}}Never{{end}}</td>
                <td>{{$state.LastDuration}}</td>
                <td>{{$state.ConsecutiveErrors}}</td>
                <td>{{$state.TotalCollections}}</td>
                <td>{{$state.TotalErrors}}/{{$state.TotalCollections}}</td>
                <td>{{$state.CurrentInterval}}</td>
            </tr>
            {{end}}
        </table>
    </div>
</body>
</html>
{{end}}

{{define "tracing"}}
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<h1>{{.Title}}</h1>
<p>Status: {{if .Enabled}}Enabled{{else}}Disabled{{end}}</p>
{{if .Enabled}}
<h2>Configuration</h2>
<ul>
<li>Sample Rate: {{.Config.SampleRate}}</li>
<li>Endpoint: {{.Config.Endpoint}}</li>
<li>Insecure: {{.Config.Insecure}}</li>
</ul>
{{end}}
</body>
</html>
{{end}}

{{define "patterns"}}
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<h1>{{.Title}}</h1>
<p>Status: {{if .Enabled}}Enabled{{else}}Disabled{{end}}</p>
</body>
</html>
{{end}}

{{define "scheduler"}}
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<h1>{{.Title}}</h1>
<p>Status: {{if .Enabled}}Enabled{{else}}Disabled{{end}}</p>
</body>
</html>
{{end}}

{{define "health"}}
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<h1>{{.Title}}</h1>
<h2>Health Checks</h2>
<table>
<tr><th>Check</th><th>Status</th><th>Message</th></tr>
{{range $name, $check := .Checks}}
<tr>
<td>{{$name}}</td>
<td>{{$check.Status}}</td>
<td>{{$check.Message}}</td>
</tr>
{{end}}
</table>
</body>
</html>
{{end}}

{{define "runtime"}}
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<h1>{{.Title}}</h1>
<h2>Go Runtime</h2>
<table>
{{range $key, $value := .Runtime}}
<tr><td>{{$key}}</td><td>{{$value}}</td></tr>
{{end}}
</table>
<h2>Memory Statistics</h2>
<table>
{{range $key, $value := .Memory}}
<tr><td>{{$key}}</td><td>{{$value}}</td></tr>
{{end}}
</table>
</body>
</html>
{{end}}

{{define "config"}}
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<h1>{{.Title}}</h1>
<h2>Debug Configuration</h2>
<pre>{{range $key, $value := .Debug}}{{$key}}: {{$value}}
{{end}}</pre>
</body>
</html>
{{end}}

{{define "cache"}}
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<h1>{{.Title}}</h1>

<h2>Cache Statistics</h2>
<table>
<tr><th>Metric</th><th>Value</th></tr>
{{range $key, $value := .Stats}}
<tr><td>{{$key}}</td><td>{{$value}}</td></tr>
{{end}}
</table>

<h2>Cache Metrics</h2>
<table>
<tr><th>Metric</th><th>Value</th></tr>
<tr><td>Hit Rate</td><td>{{.Metrics.HitRate}}</td></tr>
<tr><td>Total Entries</td><td>{{.Metrics.EntryCount}}</td></tr>
<tr><td>Total Size</td><td>{{.Metrics.TotalSize}} bytes</td></tr>
<tr><td>Average TTL</td><td>{{.Metrics.AverageTTL}}</td></tr>
<tr><td>TTL Extensions</td><td>{{.Metrics.TTLExtensions}}</td></tr>
<tr><td>TTL Reductions</td><td>{{.Metrics.TTLReductions}}</td></tr>
<tr><td>Memory Pressure</td><td>{{.Metrics.MemoryPressure}}</td></tr>
</table>
</body>
</html>
{{end}}
`
