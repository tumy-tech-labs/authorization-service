{{- define "authorization-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "authorization-service.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "authorization-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{- define "authorization-service.labels" -}}
helm.sh/chart: {{ include "authorization-service.chart" . }}
app.kubernetes.io/name: {{ include "authorization-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "authorization-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "authorization-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
