package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *BaseModel) BeforeCreate(*gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type User struct {
	BaseModel
	TelegramID   int64      `gorm:"uniqueIndex;not null" json:"telegram_id"`
	Username     string     `gorm:"index" json:"username"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	LanguageCode string     `gorm:"default:fa" json:"language_code"`
	Timezone     string     `gorm:"default:Asia/Tehran" json:"timezone"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	DailyReport  bool       `gorm:"default:true" json:"daily_report"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
}

type Planner struct {
	BaseModel
	OwnerID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"owner_id"`
	TargetUserID *uuid.UUID `gorm:"type:uuid;index" json:"target_user_id"`
	Title        string     `gorm:"size:160;not null" json:"title"`
	PlanDate     time.Time  `gorm:"type:date;not null;index" json:"plan_date"`
	Notes        string     `gorm:"type:text" json:"notes"`
	Status       string     `gorm:"size:24;default:active;index" json:"status"`
}

type Task struct {
	BaseModel
	PlannerID        *uuid.UUID `gorm:"type:uuid;index" json:"planner_id"`
	OwnerID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"owner_id"`
	AssigneeID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"assignee_id"`
	Title            string     `gorm:"size:240;not null" json:"title"`
	Description      string     `gorm:"type:text" json:"description"`
	DueAt            *time.Time `gorm:"index" json:"due_at"`
	Priority         string     `gorm:"size:16;default:normal;index" json:"priority"`
	Status           string     `gorm:"size:24;default:pending;index" json:"status"`
	RequiresEvidence bool       `gorm:"default:false" json:"requires_evidence"`
	CompletedAt      *time.Time `json:"completed_at"`
	RemindedAt       *time.Time `json:"reminded_at"`
}

type TaskEvidence struct {
	BaseModel
	TaskID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"task_id"`
	SubmittedByID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"submitted_by_id"`
	Kind           string     `gorm:"size:20;not null" json:"kind"`
	Text           string     `gorm:"type:text" json:"text"`
	TelegramFileID string     `gorm:"size:255" json:"telegram_file_id"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
}
type ActivityLog struct {
	BaseModel
	UserID     *uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	EntityType string     `gorm:"size:50;not null;index" json:"entity_type"`
	EntityID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"entity_id"`
	Action     string     `gorm:"size:80;not null;index" json:"action"`
	Metadata   string     `gorm:"type:jsonb" json:"metadata"`
	OccurredAt time.Time  `gorm:"not null;index" json:"occurred_at"`
}
type BotSession struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	State     string    `gorm:"size:80;not null" json:"state"`
	Data      string    `gorm:"type:jsonb" json:"data"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
}
