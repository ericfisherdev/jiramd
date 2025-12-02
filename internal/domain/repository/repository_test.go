package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/esfisher/jiramd/internal/domain"
	"github.com/esfisher/jiramd/internal/domain/repository"
)

// txKey is used as a context key for transaction state in tests.
type txKey struct{}

// TestJiraRepositoryInterface verifies that the JiraRepository interface
// can be satisfied by a mock implementation and that the interface compiles.
func TestJiraRepositoryInterface(t *testing.T) {
	var _ repository.JiraRepository = (*mockJiraRepository)(nil)

	ctx := context.Background()
	mock := &mockJiraRepository{}

	// Test FetchTicket
	ticket, err := mock.FetchTicket(ctx, "JMD-1")
	if err != nil {
		t.Errorf("FetchTicket failed: %v", err)
	}
	if ticket == nil {
		t.Error("FetchTicket returned nil ticket")
	}

	// Test FetchTicketsModifiedSince
	tickets, err := mock.FetchTicketsModifiedSince(ctx, "JMD", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Errorf("FetchTicketsModifiedSince failed: %v", err)
	}
	if tickets == nil {
		t.Error("FetchTicketsModifiedSince returned nil slice")
	}

	// Test FetchAllTickets
	allTickets, err := mock.FetchAllTickets(ctx, "JMD")
	if err != nil {
		t.Errorf("FetchAllTickets failed: %v", err)
	}
	if allTickets == nil {
		t.Error("FetchAllTickets returned nil slice")
	}

	// Test UpdateTicket
	if updatedTicket, err := mock.UpdateTicket(ctx, ticket); err != nil {
		t.Errorf("UpdateTicket failed: %v", err)
	} else if updatedTicket == nil {
		t.Error("UpdateTicket returned nil ticket")
	}

	// Test FetchComments
	comments, err := mock.FetchComments(ctx, "JMD-1")
	if err != nil {
		t.Errorf("FetchComments failed: %v", err)
	}
	if comments == nil {
		t.Error("FetchComments returned nil slice")
	}

	// Test AddComment
	comment := &domain.Comment{
		TicketKey: "JMD-1",
		Author:    "test",
		Body:      "test comment",
	}
	createdComment, err := mock.AddComment(ctx, "JMD-1", comment)
	if err != nil {
		t.Errorf("AddComment failed: %v", err)
	}
	if createdComment == nil {
		t.Error("AddComment returned nil comment")
	}

	// Test FetchProject
	project, err := mock.FetchProject(ctx, "JMD")
	if err != nil {
		t.Errorf("FetchProject failed: %v", err)
	}
	if project == nil {
		t.Error("FetchProject returned nil project")
	}

	// Test FetchProjects
	projects, err := mock.FetchProjects(ctx)
	if err != nil {
		t.Errorf("FetchProjects failed: %v", err)
	}
	if projects == nil {
		t.Error("FetchProjects returned nil slice")
	}
}

// TestMarkdownRepositoryInterface verifies that the MarkdownRepository interface
// can be satisfied by a mock implementation and that the interface compiles.
func TestMarkdownRepositoryInterface(t *testing.T) {
	var _ repository.MarkdownRepository = (*mockMarkdownRepository)(nil)

	ctx := context.Background()
	mock := &mockMarkdownRepository{}

	// Test ReadTicket
	ticket, err := mock.ReadTicket(ctx, "tickets/JMD-1.md")
	if err != nil {
		t.Errorf("ReadTicket failed: %v", err)
	}
	if ticket == nil {
		t.Error("ReadTicket returned nil ticket")
	}

	// Test WriteTicket
	if err := mock.WriteTicket(ctx, "tickets/JMD-1.md", ticket); err != nil {
		t.Errorf("WriteTicket failed: %v", err)
	}

	// Test ReadComments
	comments, err := mock.ReadComments(ctx, "tickets/JMD-1.md")
	if err != nil {
		t.Errorf("ReadComments failed: %v", err)
	}
	if comments == nil {
		t.Error("ReadComments returned nil slice")
	}

	// Test WriteComments
	if err := mock.WriteComments(ctx, "tickets/JMD-1.md", comments); err != nil {
		t.Errorf("WriteComments failed: %v", err)
	}

	// Test ListTicketFiles
	files, err := mock.ListTicketFiles(ctx, "tickets")
	if err != nil {
		t.Errorf("ListTicketFiles failed: %v", err)
	}
	if files == nil {
		t.Error("ListTicketFiles returned nil slice")
	}

	// Test GenerateIndex
	tickets := []*domain.Ticket{ticket}
	if err := mock.GenerateIndex(ctx, "tickets/index.md", tickets); err != nil {
		t.Errorf("GenerateIndex failed: %v", err)
	}

	// Test ValidateTemplate
	if err := mock.ValidateTemplate(ctx, "templates/ticket.md.tmpl"); err != nil {
		t.Errorf("ValidateTemplate failed: %v", err)
	}
}

// TestStateRepositoryInterface verifies that the StateRepository interface
// can be satisfied by a mock implementation and that the interface compiles.
func TestStateRepositoryInterface(t *testing.T) {
	var _ repository.StateRepository = (*mockStateRepository)(nil)

	ctx := context.Background()
	mock := &mockStateRepository{}

	// Test SaveTicketState
	ticketState := &repository.TicketSyncState{
		TicketKey:         "JMD-1",
		LastSynced:        time.Now(),
		LastModifiedLocal: time.Now(),
		LastModifiedJira:  time.Now(),
		IsDirty:           false,
	}
	if err := mock.SaveTicketState(ctx, ticketState); err != nil {
		t.Errorf("SaveTicketState failed: %v", err)
	}

	// Test GetTicketState
	retrievedState, err := mock.GetTicketState(ctx, "JMD-1")
	if err != nil {
		t.Errorf("GetTicketState failed: %v", err)
	}
	if retrievedState == nil {
		t.Error("GetTicketState returned nil state")
	}

	// Test GetTicketsModifiedSince
	modifiedTickets, err := mock.GetTicketsModifiedSince(ctx, time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Errorf("GetTicketsModifiedSince failed: %v", err)
	}
	if modifiedTickets == nil {
		t.Error("GetTicketsModifiedSince returned nil slice")
	}

	// Test GetDirtyTickets
	dirtyTickets, err := mock.GetDirtyTickets(ctx)
	if err != nil {
		t.Errorf("GetDirtyTickets failed: %v", err)
	}
	if dirtyTickets == nil {
		t.Error("GetDirtyTickets returned nil slice")
	}

	// Test GetConflictedTickets
	conflictedTickets, err := mock.GetConflictedTickets(ctx)
	if err != nil {
		t.Errorf("GetConflictedTickets failed: %v", err)
	}
	if conflictedTickets == nil {
		t.Error("GetConflictedTickets returned nil slice")
	}

	// Test DeleteTicketState
	if err := mock.DeleteTicketState(ctx, "JMD-1"); err != nil {
		t.Errorf("DeleteTicketState failed: %v", err)
	}

	// Test SaveProjectState
	projectState := &repository.ProjectSyncState{
		ProjectKey:          "JMD",
		LastFullSync:        time.Now(),
		LastIncrementalSync: time.Now(),
		TicketCount:         10,
	}
	if err := mock.SaveProjectState(ctx, projectState); err != nil {
		t.Errorf("SaveProjectState failed: %v", err)
	}

	// Test GetProjectState
	retrievedProjectState, err := mock.GetProjectState(ctx, "JMD")
	if err != nil {
		t.Errorf("GetProjectState failed: %v", err)
	}
	if retrievedProjectState == nil {
		t.Error("GetProjectState returned nil state")
	}

	// Test GetAllProjectStates
	allProjectStates, err := mock.GetAllProjectStates(ctx)
	if err != nil {
		t.Errorf("GetAllProjectStates failed: %v", err)
	}
	if allProjectStates == nil {
		t.Error("GetAllProjectStates returned nil slice")
	}

	// Test DeleteProjectState
	if err := mock.DeleteProjectState(ctx, "JMD"); err != nil {
		t.Errorf("DeleteProjectState failed: %v", err)
	}

	// Test transaction methods
	txCtx, err := mock.BeginTransaction(ctx)
	if err != nil {
		t.Errorf("BeginTransaction failed: %v", err)
	}
	if txCtx == nil {
		t.Error("BeginTransaction returned nil context")
	}

	if err := mock.Commit(txCtx); err != nil {
		t.Errorf("Commit failed: %v", err)
	}

	txCtx2, err := mock.BeginTransaction(ctx)
	if err != nil {
		t.Errorf("BeginTransaction (2) failed: %v", err)
	}
	if err := mock.Rollback(txCtx2); err != nil {
		t.Errorf("Rollback failed: %v", err)
	}
}

// TestTicketSyncStateStruct verifies the TicketSyncState struct compiles.
func TestTicketSyncStateStruct(t *testing.T) {
	now := time.Now()
	state := repository.TicketSyncState{
		TicketKey:         "JMD-1",
		LastSynced:        now,
		LastModifiedLocal: now,
		LastModifiedJira:  now,
		IsDirty:           true,
		ConflictDetected:  false,
	}

	if state.TicketKey != "JMD-1" {
		t.Errorf("TicketKey: got %s, want JMD-1", state.TicketKey)
	}
	if !state.IsDirty {
		t.Error("IsDirty: got false, want true")
	}
	if state.ConflictDetected {
		t.Error("ConflictDetected: got true, want false")
	}
}

// TestProjectSyncStateStruct verifies the ProjectSyncState struct compiles.
func TestProjectSyncStateStruct(t *testing.T) {
	now := time.Now()
	state := repository.ProjectSyncState{
		ProjectKey:          "JMD",
		LastFullSync:        now,
		LastIncrementalSync: now,
		TicketCount:         42,
	}

	if state.ProjectKey != "JMD" {
		t.Errorf("ProjectKey: got %s, want JMD", state.ProjectKey)
	}
	if state.TicketCount != 42 {
		t.Errorf("TicketCount: got %d, want 42", state.TicketCount)
	}
}

// Mock implementations for testing interface contracts

type mockJiraRepository struct{}

func (m *mockJiraRepository) FetchTicket(ctx context.Context, key string) (*domain.Ticket, error) {
	return &domain.Ticket{Key: key, Summary: "Test Ticket"}, nil
}

func (m *mockJiraRepository) FetchTicketsModifiedSince(ctx context.Context, projectKey string, since time.Time) ([]*domain.Ticket, error) {
	return []*domain.Ticket{}, nil
}

func (m *mockJiraRepository) FetchAllTickets(ctx context.Context, projectKey string) ([]*domain.Ticket, error) {
	return []*domain.Ticket{}, nil
}

func (m *mockJiraRepository) UpdateTicket(ctx context.Context, ticket *domain.Ticket) (*domain.Ticket, error) {
	return ticket, nil
}

func (m *mockJiraRepository) FetchComments(ctx context.Context, ticketKey string) ([]*domain.Comment, error) {
	return []*domain.Comment{}, nil
}

func (m *mockJiraRepository) AddComment(ctx context.Context, ticketKey string, comment *domain.Comment) (*domain.Comment, error) {
	comment.ID = "12345"
	return comment, nil
}

func (m *mockJiraRepository) FetchProject(ctx context.Context, projectKey string) (*domain.Project, error) {
	return &domain.Project{Key: projectKey, Name: "Test Project"}, nil
}

func (m *mockJiraRepository) FetchProjects(ctx context.Context) ([]*domain.Project, error) {
	return []*domain.Project{}, nil
}

type mockMarkdownRepository struct{}

func (m *mockMarkdownRepository) ReadTicket(ctx context.Context, filePath string) (*domain.Ticket, error) {
	return &domain.Ticket{Key: "JMD-1", Summary: "Test Ticket"}, nil
}

func (m *mockMarkdownRepository) WriteTicket(ctx context.Context, filePath string, ticket *domain.Ticket) error {
	return nil
}

func (m *mockMarkdownRepository) ReadComments(ctx context.Context, filePath string) ([]*domain.Comment, error) {
	return []*domain.Comment{}, nil
}

func (m *mockMarkdownRepository) WriteComments(ctx context.Context, filePath string, comments []*domain.Comment) error {
	return nil
}

func (m *mockMarkdownRepository) ListTicketFiles(ctx context.Context, directory string) ([]string, error) {
	return []string{}, nil
}

func (m *mockMarkdownRepository) GenerateIndex(ctx context.Context, indexPath string, tickets []*domain.Ticket) error {
	return nil
}

func (m *mockMarkdownRepository) ValidateTemplate(ctx context.Context, templatePath string) error {
	return nil
}

type mockStateRepository struct{}

func (m *mockStateRepository) SaveTicketState(ctx context.Context, state *repository.TicketSyncState) error {
	return nil
}

func (m *mockStateRepository) GetTicketState(ctx context.Context, ticketKey string) (*repository.TicketSyncState, error) {
	return &repository.TicketSyncState{TicketKey: ticketKey}, nil
}

func (m *mockStateRepository) GetTicketsModifiedSince(ctx context.Context, since time.Time) ([]*repository.TicketSyncState, error) {
	return []*repository.TicketSyncState{}, nil
}

func (m *mockStateRepository) GetDirtyTickets(ctx context.Context) ([]*repository.TicketSyncState, error) {
	return []*repository.TicketSyncState{}, nil
}

func (m *mockStateRepository) GetConflictedTickets(ctx context.Context) ([]*repository.TicketSyncState, error) {
	return []*repository.TicketSyncState{}, nil
}

func (m *mockStateRepository) DeleteTicketState(ctx context.Context, ticketKey string) error {
	return nil
}

func (m *mockStateRepository) SaveProjectState(ctx context.Context, state *repository.ProjectSyncState) error {
	return nil
}

func (m *mockStateRepository) GetProjectState(ctx context.Context, projectKey string) (*repository.ProjectSyncState, error) {
	return &repository.ProjectSyncState{ProjectKey: projectKey}, nil
}

func (m *mockStateRepository) GetAllProjectStates(ctx context.Context) ([]*repository.ProjectSyncState, error) {
	return []*repository.ProjectSyncState{}, nil
}

func (m *mockStateRepository) DeleteProjectState(ctx context.Context, projectKey string) error {
	return nil
}

func (m *mockStateRepository) BeginTransaction(ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, txKey{}, true), nil
}

func (m *mockStateRepository) Commit(ctx context.Context) error {
	return nil
}

func (m *mockStateRepository) Rollback(ctx context.Context) error {
	return nil
}
