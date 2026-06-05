// ユーザーに関するプログラムです

package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID // UserのID UUIDv4
	Name            string    // ユーザー名
	AssignedBoothID int       // 配属されたブースのID(1-1の人ならID:1-1など)
	Password        string    // パスワード
	Role            Role      // 役職 (Admin / Gakuseikai / Student / Member)
}

// Roleについて
// Admin: システム管理者 全ての権限を持つ
// Gakuseikai: 学生会員 ほとんど全ての権限を持つ　削除などのぶっこわれる系の権限は外す予定
// Student: 学生 配属されたブースを編集することができる権限を持つ
// Member: 一般ユーザー 閲覧する権限のみを持つ
type Role string

// contextにJWTから読み取ったRoleを入れるためのキー
// contextはkey-valueで入ってる
type contextKey string
const UserRoleKey contextKey = "userRole"

const (
	RoleAdmin      Role = "Admin"
	RoleGakuseikai Role = "Gakuseikai"
	RoleStudent    Role = "Student"
	RoleMember     Role = "Member"
)

func NewUser(name string, assignedBoothID int, password string, role Role) (*User, error) {
	// Roleのバリデーション
	if !role.IsValid() {
		return nil, errors.New("Role should be Admin, Gakuseikai, Student or Member")
	}
	// ID生成
	UUID := uuid.New()
	return &User{
		ID:              UUID,
		Name:            name,
		AssignedBoothID: assignedBoothID,
		Password:        password,
		Role:            role,
	}, nil
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
}

// Roleのバリデーション
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleGakuseikai, RoleStudent, RoleMember:
		return true
	}
	return false
}
