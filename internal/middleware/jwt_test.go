package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/auth"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTAuth(t *testing.T) {
	secret := []byte("test-secret-for-middleware")

	e := echo.New()

	// 検証成功時に呼び出されるダミーのハンドラ（本番のUsecaseの代わり）
	targetHandler := func(c *echo.Context) error {
		// ミドルウェアが context.Context にRequestUserをセットしてくれているかチェック
		u, ok := c.Request().Context().Value(usecase.ContextRequestUserKey).(usecase.RequestUser)
		if !ok {
			return c.String(http.StatusInternalServerError, "user not found in context")
		}
		// 取り出せたユーザー情報をそのままJSONで返す
		return c.JSON(http.StatusOK, u)
	}

	// ミドルウェアを設定したルーター（ハンドラ）を作成
	mw := JWTAuth(secret)
	handler := mw(targetHandler)

	t.Run("正常系: 正しいトークンをヘッダーにセットすると通過し、Contextに値が入る", func(t *testing.T) {
		userID := uuid.New()
		role := domain.RoleGakuseikai

		// 本物のトークンを生成
		tokenString, err := auth.GenerateToken(userID, role, 1, secret)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// ヘッダーにBearerトークンをセット
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// リクエストを実行
		err = handler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// レスポンス（ダミーハンドラが返した内容）に、元のユーザーIDが含まれているか確認
		assert.Contains(t, rec.Body.String(), userID.String())
	})

	t.Run("異常系: Authorizationヘッダーがないと401エラー", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)

		var he *echo.HTTPError
		require.ErrorAs(t, err, &he)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})

	t.Run("異常系: Bearerフォーマットが不正だと401エラー", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// Bearer ではなく直接文字列を入れるなど不正なフォーマット
		req.Header.Set("Authorization", "InvalidFormatToken123")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)

		var he *echo.HTTPError
		require.ErrorAs(t, err, &he)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})

	t.Run("異常系: 無効な（別の秘密鍵で作られた）トークンだと401エラー", func(t *testing.T) {
		// 別のサーバーで作られた想定のトークン
		wrongSecret := []byte("wrong-secret")
		tokenString, _ := auth.GenerateToken(uuid.New(), domain.RoleAdmin, 1, wrongSecret)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)

		var he *echo.HTTPError
		require.ErrorAs(t, err, &he)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})
}
