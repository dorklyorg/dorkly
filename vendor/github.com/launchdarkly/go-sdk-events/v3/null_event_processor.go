package ldevents

import (
	"encoding/json"
	"time"
)

type nullEventProcessor struct{}

// NewNullEventProcessor creates a no-op implementation of EventProcessor.
func NewNullEventProcessor() EventProcessor {
	return nullEventProcessor{}
}

func (n nullEventProcessor) RecordEvaluation(EvaluationData) {}

func (n nullEventProcessor) RecordIdentifyEvent(IdentifyEventData) {}

func (n nullEventProcessor) RecordCustomEvent(CustomEventData) {}

func (n nullEventProcessor) RecordMigrationOpEvent(MigrationOpEventData) {}

func (n nullEventProcessor) RecordRawEvent(json.RawMessage) {}

func (n nullEventProcessor) Flush() {}

func (n nullEventProcessor) FlushBlocking(time.Duration) bool { return true }

func (n nullEventProcessor) Close() error {
	return nil
}
