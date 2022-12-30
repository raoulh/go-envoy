package logger

import (
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/raoulh/go-envoy/internal/config"
)

// FilterFormatter formats logs into text using logrus.TextFormatter
type FilterFormatter struct {
	formatter logrus.Formatter

	useConfig       bool
	defaultLogLevel log.Level
}

// NewFilterFormatter returns a new formatter
func NewFilterFormatter() *FilterFormatter {
	return &FilterFormatter{
		formatter: &logrus.TextFormatter{
			DisableTimestamp: true,
			QuoteEmptyFields: true,
		},
	}
}

// NewCustomFormatter returns a new formatter
func NewCustomFormatter(useconfig bool, defloglevel log.Level) *FilterFormatter {
	return &FilterFormatter{
		formatter: &logrus.TextFormatter{
			DisableTimestamp: true,
			QuoteEmptyFields: true,
		},
		useConfig:       useconfig,
		defaultLogLevel: defloglevel,
	}
}

// Format renders a single log entry
func (f *FilterFormatter) Format(entry *log.Entry) ([]byte, error) {
	defaultLogLevel, err := logrus.ParseLevel(config.Config.String("log.default"))
	if err != nil {
		defaultLogLevel = logrus.InfoLevel
	}

	//get current log domain
	currentDom := "default"
	if val, ok := entry.Data["domain"]; ok {
		if dom, ok := val.(string); ok {
			currentDom = dom
		}
	}

	wantedLevel := defaultLogLevel
	if f.useConfig {
		//check if there is a specific log level for this domain
		lvl := config.Config.String("log." + currentDom)
		if lvl != "" {
			l, err := logrus.ParseLevel(lvl)
			if err == nil {
				wantedLevel = l
			}
		}
	} else {
		wantedLevel = f.defaultLogLevel
	}

	//	fmt.Printf("currentDom=%s  lvl=%s  entry.Level=%v(%d)  wantedLevel=%v(%d)  displayed=%v\n",
	//		currentDom, lvl, entry.Level, entry.Level, wantedLevel, wantedLevel, entry.Level <= wantedLevel)

	if entry.Level <= wantedLevel {
		return f.formatter.Format(entry)
	}

	return nil, nil
}
