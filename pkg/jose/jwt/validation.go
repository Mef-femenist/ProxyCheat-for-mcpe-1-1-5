package jwt

import "time"

const (
	
	DefaultLeeway = 1.0 * time.Minute
)

type Expected struct {
	
	Issuer string
	
	Subject string
	
	Audience Audience
	
	ID string
	
	Time time.Time
}

func (e Expected) WithTime(t time.Time) Expected {
	e.Time = t
	return e
}

func (c Claims) Validate(e Expected) error {
	return c.ValidateWithLeeway(e, DefaultLeeway)
}

func (c Claims) ValidateWithLeeway(e Expected, leeway time.Duration) error {
	if e.Issuer != "" && e.Issuer != c.Issuer {
		return ErrInvalidIssuer
	}

	if e.Subject != "" && e.Subject != c.Subject {
		return ErrInvalidSubject
	}

	if e.ID != "" && e.ID != c.ID {
		return ErrInvalidID
	}

	if len(e.Audience) != 0 {
		for _, v := range e.Audience {
			if !c.Audience.Contains(v) {
				return ErrInvalidAudience
			}
		}
	}

	if !e.Time.IsZero() {
		if c.NotBefore != nil && e.Time.Add(leeway).Before(c.NotBefore.Time()) {
			return ErrNotValidYet
		}

		if c.Expiry != nil && e.Time.Add(-leeway).After(c.Expiry.Time()) {
			return ErrExpired
		}

		if c.IssuedAt != nil && e.Time.Add(leeway).Before(c.IssuedAt.Time()) {
			return ErrIssuedInTheFuture
		}
	}

	return nil
}
