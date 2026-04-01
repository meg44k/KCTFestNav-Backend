package user

import "testing"

func TestUser(t *testing.T) {
	t.Run("ユーザーコンストラクタのチェック", func(t *testing.T) {
		user, err := NewUser("TestUser", "9", "password", RoleAdmin)
		t.Logf("Error: %v", err)
		t.Logf("ID: %s", user.ID)

		if user.Name != "TestUser" {
			t.Error("名前がうまくいってないよ")
		}
		if user.AssignedBoothID != "9" {
			t.Error("ブースIDがうまくいってないよ")
		}
		if user.Password != "password" {
			t.Error("パスワードがうまくいってないよ")
		}
		if user.Role != "Admin" {
			t.Error("役職割り当てがうまくいってないよ")
		}
	})

	t.Run("役職のバリデーションチェック", func(t *testing.T) {
		_, err := NewUser("TestUser", "9", "password", "Hello")
		t.Logf("Error: %v", err)
		if err == nil {
			t.Error("役職のバリデーションがうまくできてないよ")
		}
	})
}
