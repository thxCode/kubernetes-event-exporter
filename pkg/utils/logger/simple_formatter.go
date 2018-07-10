package logger

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37

	LogKubernetesHostKey = "log-k8s-host"
	LogScopeKey          = "log-scope"
	indentSpace          = "           "
	indentPoint          = "    └────> "
	breakChar            = '\n'
)

var (
	baseTimestamp = time.Now()
)

type simpleFormatter struct {
	// Whether the logger's out is to a terminal
	isTerminal bool

	sync.Once
}

func (f *simpleFormatter) init(entry *logrus.Entry) {
	if entry.Logger != nil {
		if out := entry.Logger.Out; out != nil {
			switch v := out.(type) {
			case *os.File:
				f.isTerminal = terminal.IsTerminal(int(v.Fd()))
			default:
				f.isTerminal = false
			}
		}
	}
}

func (f *simpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	f.Do(func() { f.init(entry) })

	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		if k != LogKubernetesHostKey && k != LogScopeKey && k != logrus.ErrorKey {
			keys = append(keys, k)
		}
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	if f.isTerminal {
		fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m[%04d]", levelColor, strings.ToUpper(entry.Level.String())[0:4], int(entry.Time.Sub(baseTimestamp)/time.Second))
	} else {
		fmt.Fprintf(b, "%s[%04d]", strings.ToUpper(entry.Level.String())[0:4], int(entry.Time.Sub(baseTimestamp)/time.Second))
	}

	if logK8sHostData := entry.Data[LogKubernetesHostKey]; logK8sHostData != nil {
		if f.isTerminal {
			fmt.Fprintf(b, " \x1b[3m%-44.44s\x1b[0m", logK8sHostData)
		} else {
			fmt.Fprintf(b, " %-44.44s", logK8sHostData)
		}
	} else {
		fmt.Fprintf(b, " %-44s", " ")
	}

	if logScopeData := entry.Data[LogScopeKey]; logScopeData != nil {
		if f.isTerminal {
			fmt.Fprintf(b, " \x1b[1m%-15.15s\x1b[0m", logScopeData)
		} else {
			fmt.Fprintf(b, " %-15.15s", logScopeData)
		}
	} else {
		if f.isTerminal {
			fmt.Fprintf(b, " \x1b[1m%-15s\x1b[0m", "MAIN")
		} else {
			fmt.Fprintf(b, " %-15s", "MAIN")
		}
	}

	newline := false
	if len(keys) != 0 {
		newline = true
		for _, k := range keys {
			v := entry.Data[k]
			if f.isTerminal {
				fmt.Fprintf(b, " \x1b[1;%dm%s\x1b[0m=", levelColor, k)
			} else {
				fmt.Fprintf(b, " %s=", k)
			}

			stringVal, ok := v.(string)
			if !ok {
				stringVal = fmt.Sprint(v)
			}
			b.WriteString(fmt.Sprintf("%q", stringVal))
		}
	}

	if newline {
		headerLine := true
		sc := bufio.NewScanner(strings.NewReader(entry.Message))
		for sc.Scan() {
			b.WriteByte(breakChar)
			if headerLine {
				headerLine = false
				b.WriteString(indentPoint)
			} else {
				b.WriteString(indentSpace)
			}
			b.WriteString(sc.Text())
		}
	} else {
		fmt.Fprintf(b, " -> %s", entry.Message)
	}

	if errData := entry.Data[logrus.ErrorKey]; errData != nil {
		if f.isTerminal {
			fmt.Fprintf(b, " \x1b[1;33m=>\x1b[0m \x1b[%dm%s\x1b[0m", red, errData)
		} else {
			fmt.Fprintf(b, " => %s", errData)
		}
	}

	b.WriteByte(breakChar)
	return b.Bytes(), nil
}

func CreateLogContext(scope, kubernetesHost string) logrus.Fields {
	if len(scope) == 0 {
		scope = "main"
	}

	if len(kubernetesHost) == 0 {
		return logrus.Fields{
			LogScopeKey: scope,
		}
	}

	return logrus.Fields{
		LogKubernetesHostKey: kubernetesHost,
		LogScopeKey:          scope,
	}
}

func NewSimpleFormatter() *simpleFormatter {
	return &simpleFormatter{}
}
