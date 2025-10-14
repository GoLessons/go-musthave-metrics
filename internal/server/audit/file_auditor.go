package audit

import (
	"context"
	"github.com/goccy/go-json"
	"os"
)

type FileAuditor struct {
	filePath string
}

func NewFileAuditor(filePath string) *FileAuditor {
	return &FileAuditor{filePath: filePath}
}

func (a *FileAuditor) Journal(ctx context.Context, item *JournalItem) bool {
	if a.filePath == "" || item == nil {
		return false
	}

	data, err := json.Marshal(item)
	if err != nil {
		return false
	}

	f, err := os.OpenFile(a.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return false
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
		}
	}(f)

	if ctx.Err() != nil {
		return false
	}
	_, err = f.Write(append(data, '\n'))

	return err == nil
}
