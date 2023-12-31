package renderer

const markdownTemplate = `{{- if .Version -}}
## {{ .Version }}{{ with .Date }} - {{ . }}{{ end }}

{{ end }}

{{- with .Notes -}}
{{ . }}

{{ end }}


{{- with .Sections -}}


{{- with .breaking -}}
### ⚠️️ Breaking changes ⚠️
{{- range . }}
- {{ . }}
{{- end }}

{{ end }}


{{- with .security -}}
### 🛡️ Security notices
{{- range . }}
- {{ . }}
{{- end }}

{{ end }}


{{- with .enhancement -}}
### 🚀 Enhancements
{{- range . }}
- {{ . }}
{{- end }}

{{ end }}


{{- with .bugfix -}}
### 🐞 Bug fixes
{{- range . }}
- {{ . }}
{{- end }}

{{ end }}


{{- with .dependency -}}
### ⛓️ Dependencies
{{- range . }}
- {{ . }}
{{- end }}

{{ end }}


{{- end }} {{- /* end .Sections */ -}}
`
