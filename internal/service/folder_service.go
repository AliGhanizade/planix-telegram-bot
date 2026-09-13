package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/AliGhanizade/planix-telegram-bot/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// maxFolderName is the folder name length limit.
const maxFolderName = 80

// FolderService organizes tasks into folders and subfolders.
type FolderService struct {
	db      *gorm.DB
	folders *repository.FolderRepository
	tasks   *repository.TaskRepository
	log     *zap.Logger
}

// NewFolderService builds the folder service.
func NewFolderService(db *gorm.DB, log *zap.Logger) *FolderService {
	return &FolderService{db: db, folders: repository.NewFolder(db), tasks: repository.NewTask(db), log: log}
}

// FolderView is a folder with its computed kind and task count.
type FolderView struct {
	Folder   domain.Folder
	Kind     string // own | shared
	SubCount int64
}

// List returns the folders of a user: owned ones plus delegation
// folders that arrived from others.
func (s *FolderService) List(ctx context.Context, userID uuid.UUID) ([]FolderView, error) {
	owned, err := s.folders.ListOwned(ctx, userID)
	if err != nil {
		return nil, err
	}
	shared, err := s.folders.ListShared(ctx, userID)
	if err != nil {
		return nil, err
	}
	views := make([]FolderView, 0, len(owned)+len(shared))
	for _, f := range owned {
		views = append(views, FolderView{Folder: f, Kind: "own"})
	}
	for _, f := range shared {
		views = append(views, FolderView{Folder: f, Kind: "shared"})
	}
	return views, nil
}

// Create makes a folder, optionally inside a parent folder.
func (s *FolderService) Create(ctx context.Context, ownerID uuid.UUID, name string, parentID *uuid.UUID) (*domain.Folder, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("folder name is required")
	}
	if len([]rune(name)) > maxFolderName {
		return nil, fmt.Errorf("folder name is too long")
	}
	if parentID != nil {
		parent, err := s.folders.GetByID(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("parent folder not found")
		}
		if parent.OwnerID != ownerID {
			return nil, fmt.Errorf("parent folder is not yours")
		}
	}
	f := &domain.Folder{OwnerID: ownerID, Name: name, ParentID: parentID}
	if err := s.folders.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

// Delete removes a folder the user owns.
func (s *FolderService) Delete(ctx context.Context, userID, folderID uuid.UUID) error {
	f, err := s.folders.GetByID(ctx, folderID)
	if err != nil {
		return err
	}
	if f.OwnerID != userID {
		return fmt.Errorf("folder is not yours")
	}
	return s.folders.Delete(ctx, folderID)
}

// GetForUser returns a folder the user may see.
func (s *FolderService) GetForUser(ctx context.Context, userID, folderID uuid.UUID) (*domain.Folder, error) {
	f, err := s.folders.GetByID(ctx, folderID)
	if err != nil {
		return nil, err
	}
	if f.OwnerID != userID {
		shared := false
		views, _ := s.List(ctx, userID)
		for _, v := range views {
			if v.Folder.ID == folderID {
				shared = true
				break
			}
		}
		if !shared {
			return nil, fmt.Errorf("folder not found")
		}
	}
	return f, nil
}

// TasksInFolder lists the tasks of a folder visible to the user.
func (s *FolderService) TasksInFolder(ctx context.Context, userID, folderID uuid.UUID) ([]domain.Task, error) {
	if _, err := s.GetForUser(ctx, userID, folderID); err != nil {
		return nil, err
	}
	return s.folders.TasksInFolder(ctx, folderID, userID)
}

// Link puts a visible task into a folder the user owns.
func (s *FolderService) Link(ctx context.Context, userID, taskID, folderID uuid.UUID) error {
	f, err := s.folders.GetByID(ctx, folderID)
	if err != nil {
		return err
	}
	if f.OwnerID != userID {
		return fmt.Errorf("you can only organize tasks into your own folders")
	}
	if _, err := s.tasks.GetByID(ctx, taskID); err != nil {
		return fmt.Errorf("task not found")
	}
	return s.folders.Link(ctx, taskID, folderID)
}

// Unlink removes a task from a folder.
func (s *FolderService) Unlink(ctx context.Context, userID, taskID, folderID uuid.UUID) error {
	f, err := s.folders.GetByID(ctx, folderID)
	if err != nil {
		return err
	}
	if f.OwnerID != userID {
		return fmt.Errorf("folder is not yours")
	}
	return s.folders.Unlink(ctx, taskID, folderID)
}

// LinksOfTask returns the ids of visible folders the task belongs to.
func (s *FolderService) LinksOfTask(ctx context.Context, userID, taskID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := s.folders.LinksForTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	views, err := s.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	visible := map[uuid.UUID]bool{}
	for _, v := range views {
		visible[v.Folder.ID] = true
	}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if visible[id] {
			out = append(out, id)
		}
	}
	return out, nil
}

// FoldersOfTask returns the folders of a task the user may see.
func (s *FolderService) FoldersOfTask(ctx context.Context, userID, taskID uuid.UUID) ([]FolderView, error) {
	ids, err := s.folders.LinksForTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	views, err := s.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	idSet := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	out := make([]FolderView, 0)
	for _, v := range views {
		if idSet[v.Folder.ID] {
			out = append(out, v)
		}
	}
	return out, nil
}
