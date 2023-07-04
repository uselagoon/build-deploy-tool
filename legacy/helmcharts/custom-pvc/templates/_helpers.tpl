{{/* vim: set filetype=mustache: */}}
{{/*

{{/*
  Create chart name and version as used by the chart label.
  */}}
  {{- define "custom-pvc.chart" -}}
  {{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
  {{- end -}}


{{- define "custom-pvc.name" -}}
{{- printf "lcv-%s" .Values | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}

{{- define "custom-pvc.labels" -}}
helm.sh/chart: {{ template "custom-pvc.chart" .Context }}
{{ include "custom-pvc.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Context.Release.Service }}
{{ include "custom-pvc.lagoonLabels" .Context }}
{{- end -}}

{{/*
Selector labels
*/}}

{{- define "custom-pvc.selectorLabels" -}}
app.kubernetes.io/name: {{ include "custom-pvc.name" . }}
app.kubernetes.io/instance: {{ .Context.Release.Name }}
{{- end -}}

{{/*
Lagoon Labels
*/}}
{{- define "custom-pvc.lagoonLabels" -}}
lagoon.sh/service: {{ .Values.Release.Name }}
lagoon.sh/service-type: {{ .Values.Chart.Name }}
lagoon.sh/project: {{ .Values.project }}
lagoon.sh/environment: {{ .Values.environment }}
lagoon.sh/environmentType: {{ .Values.environmentType }}
lagoon.sh/buildType: {{ .Values.buildType }}
{{- end -}}


{{/*
Annotations
*/}}
{{- define "custom-pvc.annotations" -}}
lagoon.sh/version: {{ .Values.lagoonVersion | quote }}
{{- if .Values.branch }}
lagoon.sh/branch: {{ .Values.branch | quote }}
{{- end }}
{{- if .Values.prNumber }}
lagoon.sh/prNumber: {{ .Values.prNumber | quote }}
lagoon.sh/prHeadBranch: {{ .Values.prHeadBranch | quote }}
lagoon.sh/prBaseBranch: {{ .Values.prBaseBranch | quote }}
{{- end }}
{{- end -}}
