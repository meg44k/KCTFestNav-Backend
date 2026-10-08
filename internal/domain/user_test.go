package domain

import (
	"bytes"
	"testing"
)

func TestUser(t *testing.T) {
	t.Run("ユーザーコンストラクタのチェック", func(t *testing.T) {
		// NewUser(name, loginID, password, assignedBoothID, role) の順に修正
		user, err := NewUser(UserParams{
			Name:            "TestUser",
			LoginID:         "test_login_id",
			Password:        []byte("password"),
			AssignedBoothID: 9,
			Role:            RoleAdmin,
		})
		t.Logf("Error: %v", err)
		t.Logf("ID: %s", user.ID)

		if user.Name != "TestUser" {
			t.Error("名前がうまくいってないよ")
		}
		if user.LoginID != "test_login_id" {
			t.Error("ログインIDがうまくいってないよ")
		}
		if user.AssignedBoothID != 9 {
			t.Error("ブースIDがうまくいってないよ")
		}
		if !bytes.Equal(user.Password, []byte("password")) {
			t.Error("パスワードがうまくいってないよ")
		}
		if user.Role != RoleAdmin {
			t.Error("役職割り当てがうまくいってないよ")
		}
	})

	t.Run("役職のバリデーションチェック", func(t *testing.T) {
		_, err := NewUser(UserParams{
			Name:            "TestUser",
			LoginID:         "test_login_id",
			Password:        []byte("password"),
			AssignedBoothID: 9,
			Role:            "Hello",
		})
		t.Logf("Error: %v", err)
		if err == nil {
			t.Error("役職のバリデーションがうまくできてないよ")
		}
	})
}

