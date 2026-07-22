package repo

import (
	"context"
	"testing"
	"time"

	stModel "donetick.com/core/internal/subtask/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSubtaskCompletionGuardAndChoreScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&stModel.SubTask{}); err != nil {
		t.Fatal(err)
	}

	parentID := 10
	completed := time.Now()
	rows := []stModel.SubTask{
		{ID: parentID, ChoreID: 1, Name: "parent"},
		{ID: 11, ChoreID: 1, Name: "open child", ParentId: &parentID},
		{ID: 12, ChoreID: 1, Name: "done child", ParentId: &parentID, CompletedAt: &completed},
		{ID: 13, ChoreID: 2, Name: "other chore"},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}

	repository := NewSubTasksRepository(db)
	hasIncomplete, err := repository.HasIncompleteChildren(context.Background(), 1, parentID)
	if err != nil {
		t.Fatal(err)
	}
	if !hasIncomplete {
		t.Fatal("expected incomplete child to block parent completion")
	}

	if err := repository.UpdateSubTaskStatus(context.Background(), 7, 1, 13, &completed); err != nil {
		t.Fatal(err)
	}
	var other stModel.SubTask
	if err := db.First(&other, 13).Error; err != nil {
		t.Fatal(err)
	}
	if other.CompletedAt != nil {
		t.Fatal("subtask from another chore must not be updated")
	}
}
