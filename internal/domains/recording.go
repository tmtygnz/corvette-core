package domains

import (
	"corvette/internal/database"
	"time"
)

type Recording struct {
	RecordID   int
	FromCamera int
	FileName   string
	StartedAt  time.Time
	Duration   int // max 5 mins
}

type CreateRecordingOpts struct {
	//govalid:required
	//govalid:numeric
	FromCamera int

	//govalid:required
	FileName string

	//govalid:required
	StartedAt time.Time
}

type RecordingService interface {
	CreateRecording() (*Recording, error)
	SetDuration(duration int, id int) (*Recording, error)
	ListRecordings() ([]*Recording, error)
	DeleteRecording(id int) error
}

func RecordingFromSQLC(raw database.Recording) *Recording {
	return &Recording{
		RecordID:   int(raw.RecordID),
		FromCamera: int(raw.FromCamera),
		FileName:   raw.FileName,
		StartedAt:  raw.StartedAt,
		Duration:   int(raw.Duration),
	}
}
