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
	EndedAt    *time.Time
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

type GetRecordingForOpts struct {
	FromCamera int64
	QueryStart time.Time
	QueryEnd   time.Time
}

type RecordingService interface {
	CreateRecording(opts *CreateRecordingOpts) (*Recording, error)
	SetEndAt(endTime time.Time, id int) (*Recording, error)
	GetRecordingFor(opts *GetRecordingForOpts) (*[]Recording, error)
	ListRecordings() ([]*Recording, error)
	DeleteRecording(id int) error
}

func RecordingFromSQLC(raw database.Recording) *Recording {
	var endedAtPtr *time.Time

	if raw.EndedAt.Valid {
		t := raw.EndedAt.Time
		endedAtPtr = &t
	}

	return &Recording{
		RecordID:   int(raw.RecordID),
		FromCamera: int(raw.FromCamera),
		FileName:   raw.FileName,
		StartedAt:  raw.StartedAt,
		EndedAt:    endedAtPtr, // Safe pointer containing your real DB date!
	}
}
