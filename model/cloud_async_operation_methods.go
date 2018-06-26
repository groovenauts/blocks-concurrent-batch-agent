package model

import (
	"time"
)

func (m *CloudAsyncOperation) AppendLog(msg string) {
	m.Logs = append(m.Logs, CloudAsyncOperationLog{CreatedAt: time.Now(), Message: msg})
}
