package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/google/uuid"
)

// 来場者のいいねに使う番号。<UUID>.<署名>。番号を自分で作れないようにサーバーが署名する
var ErrInvalidVoter = errors.New("invalid voter")

func signVoter(id string, secret []byte) string {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(id))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// 新しい番号を作る。戻り値は cookie に入れる文字列と、保存に使う UUID
func NewVoterToken(secret []byte) (string, string) {
	id := uuid.NewString()
	return id + "." + signVoter(id, secret), id
}

// 署名を確かめて UUID を返す。秘密が空のときは何も通さない
func ParseVoterToken(token string, secret []byte) (string, error) {
	if len(secret) == 0 {
		return "", ErrInvalidVoter
	}
	id, sig, ok := strings.Cut(token, ".")
	if !ok {
		return "", ErrInvalidVoter
	}
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidVoter
	}
	if !hmac.Equal([]byte(sig), []byte(signVoter(id, secret))) {
		return "", ErrInvalidVoter
	}
	return id, nil
}
