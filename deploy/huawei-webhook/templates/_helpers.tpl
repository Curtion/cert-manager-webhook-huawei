{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "huawei-webhook.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "huawei-webhook.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "huawei-webhook.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "huawei-webhook.selfSignedIssuer" -}}
{{ printf "%s-selfsign" (include "huawei-webhook.fullname" .) }}
{{- end -}}

{{- define "huawei-webhook.rootCAIssuer" -}}
{{ printf "%s-ca" (include "huawei-webhook.fullname" .) }}
{{- end -}}

{{- define "huawei-webhook.rootCACertificate" -}}
{{ printf "%s-ca" (include "huawei-webhook.fullname" .) }}
{{- end -}}

{{- define "huawei-webhook.servingCertificate" -}}
{{ printf "%s-webhook-tls" (include "huawei-webhook.fullname" .) }}
{{- end -}}
