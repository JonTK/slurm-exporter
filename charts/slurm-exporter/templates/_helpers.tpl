{{/*
Expand the name of the chart.
*/}}
{{- define "slurm-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "slurm-exporter.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "slurm-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "slurm-exporter.labels" -}}
helm.sh/chart: {{ include "slurm-exporter.chart" . }}
{{ include "slurm-exporter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "slurm-exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "slurm-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "slurm-exporter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "slurm-exporter.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the ConfigMap to use
*/}}
{{- define "slurm-exporter.configMapName" -}}
{{- if .Values.configMap.create }}
{{- default (include "slurm-exporter.fullname" .) .Values.configMap.name }}
{{- else }}
{{- .Values.configMap.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the SLURM auth Secret to use
*/}}
{{- define "slurm-exporter.slurmAuthSecretName" -}}
{{- if .Values.secrets.slurmAuth.create }}
{{- default (printf "%s-slurm-auth" (include "slurm-exporter.fullname" .)) .Values.secrets.slurmAuth.name }}
{{- else }}
{{- .Values.secrets.slurmAuth.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the TLS Secret to use
*/}}
{{- define "slurm-exporter.tlsSecretName" -}}
{{- if .Values.secrets.tls.create }}
{{- default (printf "%s-tls" (include "slurm-exporter.fullname" .)) .Values.secrets.tls.name }}
{{- else }}
{{- .Values.secrets.tls.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the basic auth Secret to use
*/}}
{{- define "slurm-exporter.basicAuthSecretName" -}}
{{- if .Values.secrets.basicAuth.create }}
{{- default (printf "%s-basic-auth" (include "slurm-exporter.fullname" .)) .Values.secrets.basicAuth.name }}
{{- else }}
{{- .Values.secrets.basicAuth.name }}
{{- end }}
{{- end }}

{{/*
Create the image name
*/}}
{{- define "slurm-exporter.image" -}}
{{- $registryName := .Values.image.registry -}}
{{- $repositoryName := .Values.image.repository -}}
{{- $tag := .Values.image.tag | toString -}}
{{- $digest := .Values.image.digest -}}
{{- $globalRegistry := .Values.global.imageRegistry -}}
{{- if $globalRegistry }}
    {{- $registryName = $globalRegistry -}}
{{- end -}}
{{- if $registryName }}
    {{- printf "%s/%s" $registryName $repositoryName -}}
{{- else }}
    {{- $repositoryName -}}
{{- end -}}
{{- if $digest }}
    {{- printf "@%s" $digest -}}
{{- else }}
    {{- printf ":%s" $tag -}}
{{- end -}}
{{- end }}

{{/*
Validate values
*/}}
{{- define "slurm-exporter.validateValues" -}}
{{- if and .Values.config.server.tls.enabled (not .Values.secrets.tls.create) (not .Values.secrets.tls.name) }}
{{- fail "TLS is enabled but no TLS secret is configured. Set secrets.tls.create=true or provide secrets.tls.name" }}
{{- end }}
{{- if and .Values.config.slurm.auth.tokenFile (not .Values.secrets.slurmAuth.create) (not .Values.secrets.slurmAuth.name) }}
{{- fail "SLURM token authentication is configured but no auth secret is provided. Set secrets.slurmAuth.create=true or provide secrets.slurmAuth.name" }}
{{- end }}
{{- end }}