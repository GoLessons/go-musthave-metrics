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

func (a *FileAuditor) Journal(ctx context.Context, item *JournalItem) (ok bool) {
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
	defer func() {
		if cerr := f.Close(); cerr != nil && ok {
			ok = false
		}
	}()

	if ctx.Err() != nil {
		return false
	}

	_, err = f.Write(append(data, '\n'))
	ok = err == nil
	return ok
}
