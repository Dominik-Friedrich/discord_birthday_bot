package complaint

import (
	"context"
	"fmt"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/database"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/user"

	"gorm.io/gorm"
)

type Complaint struct {
	gorm.Model
	ComplainantID uint
	Complainant   *user.User `gorm:"foreignKey:ComplainantID"`

	AgainstUserID *uint
	AgainstUser   *user.User `gorm:"foreignKey:AgainstUserID"`

	Text string
}

// UserRepository is the subset of user.Repository the complaint feature
// needs to resolve the Discord users involved in a complaint or reply.
type UserRepository interface {
	GetOrCreateUser(ctx context.Context, u *user.User) error
}

type Repository struct {
	db    *database.Connection
	users UserRepository
}

// NewRepository migrates the complaint/reply schema. It must be constructed
// after the user schema has been migrated, since Complaint and Reply both
// carry foreign keys into the users table.
func NewRepository(db *database.Connection, users UserRepository) (*Repository, error) {
	if err := db.AutoMigrate(&Reply{}, &Complaint{}); err != nil {
		return nil, fmt.Errorf("migrating complaint schema: %w", err)
	}

	return &Repository{db: db, users: users}, nil
}

func (r *Repository) AddComplaint(ctx context.Context, complaint Complaint) error {
	// too lazy to use a transaction here. its fiiiiine
	if complaint.Complainant != nil {
		if err := r.users.GetOrCreateUser(ctx, complaint.Complainant); err != nil {
			return err
		}
		complaint.ComplainantID = complaint.Complainant.ID
	}

	if complaint.AgainstUser != nil {
		if err := r.users.GetOrCreateUser(ctx, complaint.AgainstUser); err != nil {
			return err
		}
		complaint.AgainstUserID = &complaint.AgainstUser.ID
	}

	return gorm.G[Complaint](r.db.DB).Create(ctx, &complaint)
}
