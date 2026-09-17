package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/clubmax/clubmax/internal/domain"
)

type paymentRepositoryStub struct {
	order         *domain.PaymentOrder
	completeError error
	reopened      int
	failed        int
}

func (s *paymentRepositoryStub) CreateRegistrationPaymentOrder(context.Context, string, string, int, int, time.Time, time.Time) (*domain.PaymentOrder, error) {
	return nil, errors.New("unexpected call")
}
func (s *paymentRepositoryStub) CreateMonthlyPaymentOrder(context.Context, string, string, int, int, time.Time, time.Time) (*domain.PaymentOrder, error) {
	return nil, errors.New("unexpected call")
}
func (s *paymentRepositoryStub) SetPaymentOrderTrack(context.Context, string, int64, time.Time) error {
	return errors.New("unexpected call")
}
func (s *paymentRepositoryStub) ReopenPaymentOrder(_ context.Context, _ string, _ time.Time) error {
	s.reopened++
	s.order.Status = domain.PaymentOrderPending
	return nil
}
func (s *paymentRepositoryStub) FailPaymentOrder(_ context.Context, _ string, _ string, _ time.Time) error {
	s.failed++
	s.order.Status = domain.PaymentOrderFailed
	return nil
}
func (s *paymentRepositoryStub) GetPaymentOrder(context.Context, string) (*domain.PaymentOrder, error) {
	return s.order, nil
}
func (s *paymentRepositoryStub) GetPaymentOrderByTrack(context.Context, int64) (*domain.PaymentOrder, error) {
	return s.order, nil
}
func (s *paymentRepositoryStub) CompletePaymentOrder(context.Context, string, int64, int64, string, time.Time) (*domain.PaymentOrder, error) {
	if s.completeError != nil {
		return nil, s.completeError
	}
	s.order.Status = domain.PaymentOrderSucceeded
	return s.order, nil
}

type paymentGatewayStub struct {
	verification GatewayVerification
	err          error
}

func (paymentGatewayStub) Request(context.Context, int64, string, string, string, string) (int64, error) {
	return 0, errors.New("unexpected call")
}
func (s paymentGatewayStub) Verify(context.Context, int64) (GatewayVerification, error) {
	return s.verification, s.err
}
func (paymentGatewayStub) StartURL(int64) string { return "" }

func TestVerifiedChargeRemainsRetryableWhenLocalFulfillmentFails(t *testing.T) {
	trackID := int64(123456)
	repository := &paymentRepositoryStub{
		order:         &domain.PaymentOrder{ID: "order-1", Status: domain.PaymentOrderFailed, TrackID: &trackID, Total: 123000},
		completeError: domain.ErrConflict,
	}
	service := NewFinanceService(nil, nil, nil, repository, paymentGatewayStub{
		verification: GatewayVerification{Amount: 123000, RefNumber: "987654321"},
	})

	order, err := service.VerifyZibalPayment(context.Background(), trackID)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected local fulfillment conflict, got %v", err)
	}
	if order.Status != domain.PaymentOrderPending || repository.reopened != 1 || repository.failed != 0 {
		t.Fatalf("order must remain retryable after a confirmed charge: order=%+v reopened=%d failed=%d", order, repository.reopened, repository.failed)
	}
}

func TestUnconfirmedGatewayPaymentIsMarkedFailed(t *testing.T) {
	trackID := int64(123456)
	repository := &paymentRepositoryStub{
		order: &domain.PaymentOrder{ID: "order-1", Status: domain.PaymentOrderPending, TrackID: &trackID, Total: 123000},
	}
	service := NewFinanceService(nil, nil, nil, repository, paymentGatewayStub{err: domain.ErrPaymentFailed})

	order, err := service.VerifyZibalPayment(context.Background(), trackID)
	if !errors.Is(err, domain.ErrPaymentFailed) {
		t.Fatalf("expected payment failure, got %v", err)
	}
	if order.Status != domain.PaymentOrderFailed || repository.failed != 1 {
		t.Fatalf("unconfirmed payment was not marked failed: order=%+v failed=%d", order, repository.failed)
	}
}
