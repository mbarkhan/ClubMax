package service

import (
	"context"
	"strings"
	"time"

	"github.com/clubmax/clubmax/internal/domain"
	"github.com/clubmax/clubmax/internal/repository"
	"github.com/google/uuid"
)

type TicketService struct {
	repo     repository.TicketRepository
	messages *MessageService
	now      func() time.Time
}

func NewTicketService(r repository.TicketRepository, m *MessageService) *TicketService {
	return &TicketService{repo: r, messages: m, now: time.Now}
}

func validTicketCategory(v domain.TicketCategory) bool {
	return v == domain.TicketCategoryPayment || v == domain.TicketCategoryAccount || v == domain.TicketCategoryGym || v == domain.TicketCategoryTechnical || v == domain.TicketCategoryOther
}
func (s *TicketService) Create(ctx context.Context, actor *domain.User, t *domain.SupportTicket, body string) (*domain.SupportTicket, error) {
	if actor.Role != domain.RoleCoach && actor.Role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	t.Subject, body = strings.TrimSpace(t.Subject), strings.TrimSpace(body)
	if !validTicketCategory(t.Category) || t.Subject == "" || len([]rune(t.Subject)) > 200 || body == "" || len([]rune(body)) > 5000 {
		return nil, domain.ErrInvalid
	}
	now := s.now()
	t.ID, t.OwnerID, t.Status = uuid.NewString(), actor.ID, domain.TicketWaitingAdmin
	t.CreatedAt, t.UpdatedAt, t.LastMessageAt = now, now, now
	first := &domain.SupportTicketMessage{ID: uuid.NewString(), TicketID: t.ID, SenderID: actor.ID, Body: body, CreatedAt: now}
	if err := s.repo.CreateTicket(ctx, t, first); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TicketService) UnreadCount(ctx context.Context, actor *domain.User) (int64, error) {
	return s.repo.CountUnreadTickets(ctx, actor.ID, actor.Role)
}

func (s *TicketService) List(ctx context.Context, actor *domain.User, filter repository.TicketFilter) ([]domain.SupportTicket, error) {
	filter.ActorID, filter.ActorRole = actor.ID, actor.Role
	if filter.Status != nil && *filter.Status != domain.TicketWaitingAdmin && *filter.Status != domain.TicketWaitingMember && *filter.Status != domain.TicketClosed {
		return nil, domain.ErrInvalid
	}
	if filter.Category != nil && !validTicketCategory(*filter.Category) {
		return nil, domain.ErrInvalid
	}
	return s.repo.ListTickets(ctx, filter)
}

func (s *TicketService) Detail(ctx context.Context, actor *domain.User, id string) (*domain.SupportTicketDetail, error) {
	t, err := s.repo.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor.Role != domain.RoleAdmin && t.OwnerID != actor.ID {
		return nil, domain.ErrForbidden
	}
	items, err := s.repo.ListTicketMessages(ctx, id)
	if err != nil {
		return nil, err
	}
	if err = s.repo.MarkTicketRead(ctx, id, actor.Role); err != nil {
		return nil, err
	}
	t.Unread = false
	return &domain.SupportTicketDetail{Ticket: *t, Messages: items}, nil
}

func (s *TicketService) Reply(ctx context.Context, actor *domain.User, id, body string) (*domain.SupportTicketMessage, error) {
	t, err := s.repo.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor.Role != domain.RoleAdmin && t.OwnerID != actor.ID {
		return nil, domain.ErrForbidden
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleCoach && actor.Role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	body = strings.TrimSpace(body)
	if t.Status == domain.TicketClosed {
		return nil, domain.ErrInvalidState
	}
	if body == "" || len([]rune(body)) > 5000 {
		return nil, domain.ErrInvalid
	}
	now := s.now()
	m := &domain.SupportTicketMessage{ID: uuid.NewString(), TicketID: id, SenderID: actor.ID, SenderName: strings.TrimSpace(actor.FirstName + " " + actor.LastName), SenderRole: actor.Role, Body: body, CreatedAt: now}
	status := domain.TicketWaitingAdmin
	var assigned *string
	if actor.Role == domain.RoleAdmin {
		status, assigned = domain.TicketWaitingMember, &actor.ID
	}
	if err = s.repo.AddTicketReply(ctx, m, status, assigned); err != nil {
		return nil, err
	}
	if actor.Role == domain.RoleAdmin && s.messages != nil {
		recipient := t.OwnerID
		_, _ = s.messages.Send(ctx, actor.ID, &domain.Message{Audience: domain.AudienceDirect, RecipientID: &recipient, Subject: "پاسخ پشتیبانی به تیکت " + t.Subject, Body: "پشتیبانی کلاب‌ماکس به تیکت شما پاسخ داد."})
	}
	return m, nil
}

func (s *TicketService) SetStatus(ctx context.Context, actor *domain.User, id string, status domain.TicketStatus) error {
	t, err := s.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}
	if actor.Role != domain.RoleAdmin && t.OwnerID != actor.ID {
		return domain.ErrForbidden
	}
	if status != domain.TicketClosed && status != domain.TicketWaitingAdmin {
		return domain.ErrInvalidState
	}
	if status == domain.TicketWaitingAdmin && t.Status != domain.TicketClosed {
		return domain.ErrInvalidState
	}
	if status == domain.TicketClosed && t.Status == domain.TicketClosed {
		return domain.ErrInvalidState
	}
	now := s.now()
	var closed *time.Time
	if status == domain.TicketClosed {
		closed = &now
	}
	var admin *string
	if actor.Role == domain.RoleAdmin {
		admin = &actor.ID
		if status == domain.TicketWaitingAdmin {
			status = domain.TicketWaitingMember
		}
	}
	return s.repo.SetTicketStatus(ctx, id, status, admin, closed)
}
