package pipeline

import (
	"encoding/json"
	"fmt"
	"time"
)


type JSONTime time.Time

func (j JSONTime) MarshalJSON() ([]byte, error) {
    stamp := fmt.Sprintf("\"%s\"", time.Time(j).Format(DATE_TIME_LAYOUT))
    return []byte(stamp), nil
}

// EventType is used to distinguish between different types of pipeline logs
type EventType string

const (
    DiagnosticEventType EventType = "diagnostic_event"
    DiagnosticType      EventType = "diagnostic"
)

// LogWrapper is a structure that wraps pipeline logs with type information
type LogWrapper struct {
    Type EventType    `json:"type"`
    Data interface{} `json:"data"`
}

// MarshalJSON implements custom JSON marshalling for Diagnostic
func (d *Diagnostic) MarshalJSON() ([]byte, error) {
    type DiagnosticAlias Diagnostic // Avoid recursive MarshalJSON calls
    
    events := make([]LogWrapper, len(d.Events))
    for i, event := range d.Events {
        switch e := event.(type) {
        case *DiagnosticEvent:
            events[i] = LogWrapper{
                Type: DiagnosticEventType,
                Data: e,
            }
        case *Diagnostic:
            events[i] = LogWrapper{
                Type: DiagnosticType,
                Data: e,
            }
        default:
            return nil, fmt.Errorf("unknown event type: %T", event)
        }
    }
    
    alias := &struct {
        *DiagnosticAlias
        Events []LogWrapper `json:"logs"`
    }{
        DiagnosticAlias: (*DiagnosticAlias)(d),
        Events:          events,
    }
    
    return json.Marshal(alias)
}

// UnmarshalJSON implements custom JSON unmarshalling for Diagnostic
func (d *Diagnostic) UnmarshalJSON(data []byte) error {
    type DiagnosticAlias Diagnostic
    
    temp := &struct {
        *DiagnosticAlias
        Events []LogWrapper `json:"logs"`
    }{
        DiagnosticAlias: (*DiagnosticAlias)(d),
    }
    
    if err := json.Unmarshal(data, temp); err != nil {
        return err
    }
    
    d.Events = make([]pipelineLog, len(temp.Events))
    for i, wrapper := range temp.Events {
        switch wrapper.Type {
        case DiagnosticEventType:
            var event DiagnosticEvent
            eventData, err := json.Marshal(wrapper.Data)
            if err != nil {
                return err
            }
            if err := json.Unmarshal(eventData, &event); err != nil {
                return err
            }
            d.Events[i] = &event
            
        case DiagnosticType:
            var diagnostic Diagnostic
            diagData, err := json.Marshal(wrapper.Data)
            if err != nil {
                return err
            }
            if err := json.Unmarshal(diagData, &diagnostic); err != nil {
                return err
            }
            d.Events[i] = &diagnostic
            
        default:
            return fmt.Errorf("unknown event type: %s", wrapper.Type)
        }
    }
    
    return nil
}


