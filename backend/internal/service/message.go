package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/clubmax/clubmax/internal/domain"
	"github.com/clubmax/clubmax/internal/repository"
	"strings"
	"time"
)

type MessageService struct {
	repo  repository.MessageRepository
	users repository.UserRepository
	now   func() time.Time
}

func NewMessageService(r repository.MessageRepository, u repository.UserRepository) *MessageService {
	return &MessageService{r, u, time.Now}
}
func (s *MessageService) Send(ctx context.Context, senderID string, m *domain.Message) (*domain.Message, error) {
	sender, e := s.users.GetByID(ctx, senderID)
	if e != nil {
		return nil, e
	}
	m.Subject = strings.TrimSpace(m.Subject)
	m.Body = strings.TrimSpace(m.Body)
	if m.Subject == "" || m.Body == "" || len(m.Subject) > 200 {
		return nil, domain.ErrInvalid
	}
	var ids []string
	switch m.Audience {
	case domain.AudienceAll:
		if sender.Role != domain.RoleAdmin {
			return nil, domain.ErrForbidden
		}
		ids, e = s.repo.AllUserIDs(ctx, senderID)
	case domain.AudienceGymMembers:
		if m.GymID == nil {
			return nil, domain.ErrInvalid
		}
		if sender.Role != domain.RoleAdmin {
			var ok bool
			ok, e = s.repo.CanAccessGym(ctx, senderID, *m.GymID)
			if e == nil && !ok {
				return nil, domain.ErrMessageRecipient
			}
		}
		if e == nil {
			ids, e = s.repo.GymRecipientIDs(ctx, *m.GymID, senderID)
		}
	case domain.AudienceDirect:
		if m.RecipientID == nil {
			return nil, domain.ErrInvalid
		}
		if sender.Role != domain.RoleAdmin {
			var ok bool
			ok, e = s.repo.ShareGym(ctx, senderID, *m.RecipientID)
			if e == nil && !ok {
				return nil, domain.ErrMessageRecipient
			}
		}
		ids = []string{*m.RecipientID}
	default:
		return nil, domain.ErrInvalid
	}
	if e != nil {
		return nil, e
	}
	if len(ids) == 0 {
		return nil, domain.ErrMessageRecipient
	}
	m.ID = uuid.NewString()
	m.SenderID = senderID
	m.CreatedAt = s.now()
	recipients := make([]domain.MessageRecipient, 0, len(ids))
	for _, id := range ids {
		recipients = append(recipients, domain.MessageRecipient{MessageID: m.ID, UserID: id})
	}
	if e = s.repo.Create(ctx, m, recipients); e != nil {
		return nil, e
	}
	return m, nil
}
func (s *MessageService) Inbox(ctx context.Context, user string, limit, offset int) ([]domain.Message, error) {
	return s.repo.ListForUser(ctx, user, limit, offset)
}
func (s *MessageService) MarkRead(ctx context.Context, user, message string) error {
	return s.repo.MarkRead(ctx, message, user)
}
