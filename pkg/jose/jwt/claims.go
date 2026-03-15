package jwt

import (
	"strconv"
	"time"

	"mefproxy/pkg/jose/json"
)

type Claims struct {
	Issuer    string       `json:"iss,omitempty"`
	Subject   string       `json:"sub,omitempty"`
	Audience  Audience     `json:"aud,omitempty"`
	Expiry    *NumericDate `json:"exp,omitempty"`
	NotBefore *NumericDate `json:"nbf,omitempty"`
	IssuedAt  *NumericDate `json:"iat,omitempty"`
	ID        string       `json:"jti,omitempty"`
}

type NumericDate int64

func NewNumericDate(t time.Time) *NumericDate {
	if t.IsZero() {
		return nil
	}

	out := NumericDate(t.Unix())
	return &out
}

func (n NumericDate) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(n), 10)), nil
}

func (n *NumericDate) UnmarshalJSON(b []byte) error {
	s := string(b)

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ErrUnmarshalNumericDate
	}

	*n = NumericDate(f)
	return nil
}

func (n *NumericDate) Time() time.Time {
	if n == nil {
		return time.Time{}
	}
	return time.Unix(int64(*n), 0)
}

type Audience []string

func (s *Audience) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch v := v.(type) {
	case string:
		*s = []string{v}
	case []interface{}:
		a := make([]string, len(v))
		for i, e := range v {
			s, ok := e.(string)
			if !ok {
				return ErrUnmarshalAudience
			}
			a[i] = s
		}
		*s = a
	default:
		return ErrUnmarshalAudience
	}

	return nil
}

func (s Audience) Contains(v string) bool {
	for _, a := range s {
		if a == v {
			return true
		}
	}
	return false
}
