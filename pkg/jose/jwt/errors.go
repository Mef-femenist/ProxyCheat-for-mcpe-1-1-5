package jwt

import "errors"

var ErrUnmarshalAudience = errors.New("square/go-jose/jwt: expected string or array value to unmarshal to Audience")

var ErrUnmarshalNumericDate = errors.New("square/go-jose/jwt: expected number value to unmarshal NumericDate")

var ErrInvalidClaims = errors.New("square/go-jose/jwt: expected claims to be value convertible into JSON object")

var ErrInvalidIssuer = errors.New("square/go-jose/jwt: validation failed, invalid issuer claim (iss)")

var ErrInvalidSubject = errors.New("square/go-jose/jwt: validation failed, invalid subject claim (sub)")

var ErrInvalidAudience = errors.New("square/go-jose/jwt: validation failed, invalid audience claim (aud)")

var ErrInvalidID = errors.New("square/go-jose/jwt: validation failed, invalid ID claim (jti)")

var ErrNotValidYet = errors.New("square/go-jose/jwt: validation failed, token not valid yet (nbf)")

var ErrExpired = errors.New("square/go-jose/jwt: validation failed, token is expired (exp)")

var ErrIssuedInTheFuture = errors.New("square/go-jose/jwt: validation field, token issued in the future (iat)")

var ErrInvalidContentType = errors.New("square/go-jose/jwt: expected content type to be JWT (cty header)")
