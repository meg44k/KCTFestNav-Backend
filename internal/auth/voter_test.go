package auth

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestVoterToken_作って読める(t *testing.T) {
	secret := []byte("s")
	token, id := NewVoterToken(secret)
	if _, err := uuid.Parse(id); err != nil {
		t.Fatalf("UUID ではない: %q", id)
	}
	got, err := ParseVoterToken(token, secret)
	if err != nil || got != id {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestVoterToken_不正は弾く(t *testing.T) {
	secret := []byte("s")
	token, id := NewVoterToken(secret)
	other, _ := NewVoterToken([]byte("other"))
	tampered := uuid.NewString() + token[strings.Index(token, "."):]
	for _, bad := range []string{"", "abc", id, token + "x", other, tampered, ".", id + "."} {
		if _, err := ParseVoterToken(bad, secret); err != ErrInvalidVoter {
			t.Errorf("%q は弾くはず (err=%v)", bad, err)
		}
	}
}

func TestVoterToken_秘密が空なら読めない(t *testing.T) {
	token, _ := NewVoterToken([]byte("s"))
	if _, err := ParseVoterToken(token, nil); err != ErrInvalidVoter {
		t.Fatal("秘密が空なら弾くはず")
	}
}
