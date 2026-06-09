package booking

import (
	"context"
	"errors"
	"sup-anapa/internal/models"
	"time"
)

type SlotRepository interface {
	GetByIDWithLock(ctx context.Context, id int) (*models.Slot, error)
	SetPending(ctx context.Context, slotID int, holdExpiresAt time.Time) error
	SetAvailable(ctx context.Context, slotID int) error
}

type BookingRepository interface {
	Create(ctx context.Context, booking *models.Booking) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*models.User, error)
}

type CreateBookingUseCase struct {
	slotRepo    SlotRepository
	bookingRepo BookingRepository
	userRepo    UserRepository
	holdMinutes int
}

type CreateBookingInput struct {
	UserID      int
	SlotID      int
	PeopleCount int
	ClientEmail string
}

func NewCreateBookingUseCase(slotRepo SlotRepository, bookingRepo BookingRepository, userRepo UserRepository, holdMinutes int) *CreateBookingUseCase {
	if holdMinutes <= 0 {
		holdMinutes = 20
	}
	return &CreateBookingUseCase{
		slotRepo:    slotRepo,
		bookingRepo: bookingRepo,
		userRepo:    userRepo,
		holdMinutes: holdMinutes,
	}
}

func (uc *CreateBookingUseCase) Execute(ctx context.Context, input CreateBookingInput) (*models.Booking, time.Time, error) {
	if input.UserID < 1 {
		return nil, time.Time{}, errors.New("unauthorized")
	}
	if input.SlotID < 1 {
		return nil, time.Time{}, errors.New("slot_required")
	}
	if input.PeopleCount < 1 {
		return nil, time.Time{}, errors.New("invalid_people_count")
	}

	user, err := uc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, time.Time{}, errors.New("user_not_found")
	}

	slot, err := uc.slotRepo.GetByIDWithLock(ctx, input.SlotID)
	if err != nil {
		return nil, time.Time{}, errors.New("slot_not_found")
	}
	if slot.Status != "available" {
		return nil, time.Time{}, errors.New("slot_unavailable")
	}
	if input.PeopleCount > slot.MaxPeople {
		return nil, time.Time{}, errors.New("too_many_people")
	}

	holdExpires := time.Now().Add(time.Duration(uc.holdMinutes) * time.Minute)
	if err := uc.slotRepo.SetPending(ctx, input.SlotID, holdExpires); err != nil {
		return nil, time.Time{}, err
	}

	booking := &models.Booking{
		SlotID:      input.SlotID,
		UserID:      input.UserID,
		ClientName:  user.Username,
		ClientPhone: user.Phone,
		ClientEmail: input.ClientEmail,
		PeopleCount: input.PeopleCount,
		Status:      "pending",
	}

	if err := uc.bookingRepo.Create(ctx, booking); err != nil {
		_ = uc.slotRepo.SetAvailable(ctx, input.SlotID)
		return nil, time.Time{}, err
	}

	return booking, holdExpires, nil
}
