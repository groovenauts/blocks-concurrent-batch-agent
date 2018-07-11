package model

import (
	"time"
)

func (m *InstanceGroupOperation) AppendLog(msg string) {
	m.Logs = append(m.Logs, CloudAsyncOperationLog{CreatedAt: time.Now(), Message: msg})
}

func (m *PipelineBaseOperation) AppendLog(msg string) {
	m.Logs = append(m.Logs, CloudAsyncOperationLog{CreatedAt: time.Now(), Message: msg})
}
