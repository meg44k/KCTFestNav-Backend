package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := []byte("test-secret-key")
	userID := uuid.New()
	role := domain.RoleAdmin

	t.Run("正常系: 生成したトークンを正しくパースして情報を復元できる", func(t *testing.T) {
		// トークン生成
		tokenString, err := GenerateToken(userID, role, 1, secret)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString, "トークン文字列が空ではないこと")

		// トークン検証・パース
		claims, err := ParseToken(tokenString, secret)
		require.NoError(t, err)

		// 中身が一致するか確認
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, role, claims.Role)
	})

	t.Run("異常系: 間違ったシークレットキーでは検証に失敗する", func(t *testing.T) {
		// 本来のキーで生成
		tokenString, err := GenerateToken(userID, role, 1, secret)
		require.NoError(t, err)

		// 間違ったキーでパース（偽造チェック）
		wrongSecret := []byte("wrong-secret-key")
		claims, err := ParseToken(tokenString, wrongSecret)

		require.Error(t, err, "エラーが発生すること")
		assert.Nil(t, claims, "クレームが取得できないこと")
	})

	t.Run("異常系: 不正な文字列のトークンは弾かれる", func(t *testing.T) {
		claims, err := ParseToken("invalid.token.string", secret)

		require.Error(t, err)
		assert.Nil(t, claims)
	})
}
