package urlsigner

import (
	"errors"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

type Signer struct {
	secret []byte
	crypt  *goalone.Sword
}

func New(secret []byte) *Signer {

	return &Signer{
		secret: secret,
		crypt:  goalone.New(secret, goalone.Timestamp),
	}
}

func (s *Signer) GenerateTokenFromString(data string) string {

	var urlToSign string

	if strings.Contains(data, "?") {

		urlToSign = data + "&hash="

	} else {

		urlToSign = data + "?hash="
	}

	return string(s.crypt.Sign([]byte(urlToSign)))
}

func (s *Signer) VerifyToken(token string) error {

	_, err := s.crypt.Unsign([]byte(token))

	return err
}

func (s *Signer) Expired(token string, minutesUntilExpire int) bool {

	ts := s.crypt.Parse([]byte(token))

	return time.Since(ts.Timestamp) > time.Duration(minutesUntilExpire)*time.Minute
}

func (s *Signer) IsValid(token string, minutesUntilExpire int) error {

	if err := s.VerifyToken(token); err != nil {

		return err
	}

	if s.Expired(token, minutesUntilExpire) {

		return errors.New("token has expired")
	}

	return nil
}
