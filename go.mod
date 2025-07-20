module github.com/jontk/slurm-exporter

go 1.24.4

replace github.com/jontk/slurm-client => ../slurm-client

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
