package live

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"text/template"
	"time"
)

func init() {
	Template = template.Must(template.New("status").Parse(TemplateSource))
	data = &Data{
		LiveTargetsMap: make(map[string]int),
		LiveTargets:    make([]string, 0),
	}
}

type Kind uint8

const (
	KindPlan   = Kind(1)
	KindStep   = Kind(2)
	KindJob    = Kind(3)
	KindTarget = Kind(4)
)

type Fragment struct {
	Kind
	ID   string
	Line string
}

func UpdateTarget(id string) Fragment {
	return Fragment{Kind: KindTarget, ID: id}
}

func UpdatePlan() Fragment {
	return Fragment{Kind: KindJob}
}

func UpdateStep() Fragment {
	return Fragment{Kind: KindStep}
}
func Fprintf(frag Fragment, w io.Writer, format string, args ...any) {
	frag.Line = fmt.Sprintf(format, args...)

	mutex.Lock()
	defer mutex.Unlock()

	switch frag.Kind {
	case KindTarget:
		data.LiveTargets[data.LiveTargetsMap[frag.ID]] = frag.Line
	}
}

func SetHosts(hosts []string) {
	mutex.Lock()
	defer mutex.Unlock()
	clear(data.LiveTargetsMap)
	data.LiveTargets = make([]string, len(hosts))

	for i, h := range hosts {
		data.LiveTargets[i] = fmt.Sprintf("[%s] waiting", h)
		data.LiveTargetsMap[h] = i
	}
}

func UIPlanStarted(name string, runID string, started time.Time) {
	mutex.Lock()
	defer mutex.Unlock()

	data.RunID = runID
	data.PlanName = name
	data.RunStartedAt = started.Format(time.RFC3339)
}

func Render() string {
	mutex.RLock()
	defer mutex.RUnlock()

	var buf bytes.Buffer
	err := Template.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("cannot update the live view: %v", err).Error()
	}

	return buf.String()
}

var Template *template.Template
var data *Data
var mutex sync.RWMutex

const TemplateSource = `
Plan: {{ .PlanName }}
Run ID: {{ .RunID }}
Started: {{ .RunStartedAt }}

{{ .PlanStatus }} Step {{ .StepCurrent }}/{{ .StepTotal }}:
  Job: {{ .Job }}
  Targets: {{ .Targets }}
  Hosts: {{ .StepHosts }}
  Status: {{ .StepStatus }}
  Started: {{ .StepStartedAt }}
{{ range .LiveTargets }}
{{ . }}
{{- end }}
`

type Data struct {
	PlanName       string
	PlanStatus     string
	RunID          string
	RunStartedAt   string
	StepCurrent    string
	StepTotal      string
	Job            string
	Targets        string
	StepHosts      string
	StepStatus     string
	StepStartedAt  string
	LiveTargetsMap map[string]int
	LiveTargets    []string
}
