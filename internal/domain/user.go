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
	LoginID         string    // ログインに用いる文字列
	Password        []byte    // パスワード
	AssignedBoothID int       // 配属されたブースのID(1-1の人ならID:1-1など)
	Role            Role      // 役職 (Admin / Gakuseikai / Student / Member)
}

// Roleについて
// Admin: システム管理者 全ての権限を持つ
// Gakuseikai: 学生会員 ほとんど全ての権限を持つ　削除などのぶっこわれる系の権限は外す予定
// Student: 学生 配属されたブースを編集することができる権限を持つ
// Member: 一般ユーザー 閲覧する権限のみを持つ
type Role string

const (
	RoleAdmin      Role = "Admin"
	RoleGakuseikai Role = "Gakuseikai"
	RoleStudent    Role = "Student"
	RoleMember     Role = "Member"
)

func NewUser(name string, loginID string, password []byte, assignedBoothID int, role Role) (*User, error) {
	// Roleのバリデーション
	if !role.IsValid() {
		return nil, errors.New("Role should be Admin, Gakuseikai, Student or Member")
	}
	// ID生成
	UUID := uuid.New()
	return &User{
		ID:              UUID,
		Name:            name,
		LoginID:         loginID,
		Password:        password,
		AssignedBoothID: assignedBoothID,
		Role:            role,
	}, nil
}

func ReconstructUser(id uuid.UUID, name string, loginID string, password []byte, assignedBoothID int, role Role) (*User, error) {
	return &User{
		ID:              id,
		Name:            name,
		LoginID:         loginID,
		Password:        password,
		AssignedBoothID: assignedBoothID,
		Role:            role,
	}, nil
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
	GetByLoginID(ctx context.Context, loginID string) (*User, error)
}

// Roleのバリデーション
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleGakuseikai, RoleStudent, RoleMember:
		return true
	}
	return false
}
