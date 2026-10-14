package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ClaimsJWT define o payload criptografado do token
type ClaimsJWT struct {
	UsuarioID   int64  `json:"uid"`
	UsuarioUUID string `json:"sub"` // Subject padrão do JWT (UUID)
	Email       string `json:"email"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

// GerenciadorToken cuida da emissão e validação segura de JWTs
type GerenciadorToken struct {
	chaveSecreta []byte
	duracao      time.Duration
}

// NovoGerenciadorToken instancia o gerenciador
func NovoGerenciadorToken(chaveSecreta string, duracao time.Duration) *GerenciadorToken {
	return &GerenciadorToken{
		chaveSecreta: []byte(chaveSecreta),
		duracao:      duracao,
	}
}

// GerarToken emite um JWT assinado com HMAC-SHA256
func (g *GerenciadorToken) GerarToken(u *Usuario) (string, int64, error) {
	expiracao := time.Now().Add(g.duracao)

	claims := &ClaimsJWT{
		UsuarioID:   u.ID,
		UsuarioUUID: u.UUID,
		Email:       u.Email,
		Role:        u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiracao),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "b3-core-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(g.chaveSecreta)
	if err != nil {
		return "", 0, fmt.Errorf("falha ao assinar token: %w", err)
	}

	return tokenString, expiracao.Unix(), nil
}

// ValidarToken decodifica e verifica a assinatura e expiração do JWT
func (g *GerenciadorToken) ValidarToken(tokenString string) (*ClaimsJWT, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ClaimsJWT{}, func(token *jwt.Token) (interface{}, error) {
		// Proteção contra ataques de confusão de algoritmo (alg: none / RS256 forçado)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return g.chaveSecreta, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token inválido ou expirado: %w", err)
	}

	claims, ok := token.Claims.(*ClaimsJWT)
	if !ok || !token.Valid {
		return nil, errors.New("claims do token inválidas")
	}

	return claims, nil
}
