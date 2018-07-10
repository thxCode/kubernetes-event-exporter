package pipes

import (
	"fmt"
	"strings"
	"time"

	apiCoreV1 "k8s.io/api/core/v1"
	apisMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

const (
	printHead          = "S   LST_S   FST_S   CNT   TYPE            KIND                NS                  OBJ                                     SUB_OBJ                                 SRC                                               RES                                     MSG"
	printBodyFormatter = "%-4s%-8s%-8s%-6d%-16s%-20s%-20s%-40.40s%-40.40s%-50.50s%-40.40s%-s"
	printBreak         = '\n'
	inState            = "✓"
	outState           = "✗"
)

func showHead(bodies ...string) string {
	builder := strings.Builder{}

	builder.WriteString(printHead)
	for _, body := range bodies {
		builder.WriteByte(printBreak)
		builder.WriteString(body)
	}

	return builder.String()
}

func formatSource(es *apiCoreV1.EventSource) string {
	if es != nil {
		if len(es.Host) != 0 {
			return fmt.Sprintf("%s (%s)", es.Component, es.Host)
		}

		return es.Component
	}

	return ""
}

func formatTimestamp(m *apisMetaV1.Time) string {
	if m != nil {
		if !m.IsZero() {
			return duration.ShortHumanDuration(time.Since(m.Time))
		}
	}

	return ""
}

func printEvent(state string, event *apiCoreV1.Event) string {
	involvedObject := event.InvolvedObject

	return fmt.Sprintf(printBodyFormatter,
		state,
		formatTimestamp(&event.LastTimestamp),
		formatTimestamp(&event.FirstTimestamp),
		event.Count,
		event.Type,
		involvedObject.Kind,
		involvedObject.Namespace,
		involvedObject.Name,
		involvedObject.FieldPath,
		formatSource(&event.Source),
		event.Reason,
		event.Message,
	)
}

func printEventList(eventList *apiCoreV1.EventList) string {
	builder := strings.Builder{}

	headerLine := true
	for _, event := range eventList.Items {
		if headerLine {
			headerLine = false
		} else {
			builder.WriteByte(printBreak)
		}
		builder.WriteString(printEvent(inState, event.DeepCopy()))
	}

	return builder.String()
}
