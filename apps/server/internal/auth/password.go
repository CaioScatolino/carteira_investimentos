package auth

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// CustoBcrypt define o custo computacional (Work Factor).
// 12 = 4096 iterações (Equilíbrio ideal entre segurança institucional e tempo de resposta ~100ms)
const CustoBcrypt = 12

// GerarHashSenha gera o hash bcrypt com salt dinâmico embutido
func GerarHashSenha(senha string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(senha), CustoBcrypt)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerificarSenha compara o hash com a senha em tempo constante (anti-timing attack)
func VerificarSenha(hash, senha string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}

// ValidarForcaSenha valida regras institucionais mínimas de segurança (OWASP)
func ValidarForcaSenha(senha string) error {
	if len(senha) < 8 {
		return errors.New("a senha deve ter no mínimo 8 caracteres")
	}

	var temMaiuscula, temMinuscula, temNumero, temEspecial bool

	for _, char := range senha {
		switch {
		case unicode.IsUpper(char):
			temMaiuscula = true
		case unicode.IsLower(char):
			temMinuscula = true
		case unicode.IsNumber(char):
			temNumero = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			temEspecial = true
		}
	}

	if !temMaiuscula {
		return errors.New("a senha deve conter pelo menos uma letra maiúscula")
	}
	if !temMinuscula {
		return errors.New("a senha deve conter pelo menos uma letra minúscula")
	}
	if !temNumero {
		return errors.New("a senha deve conter pelo menos um número")
	}
	if !temEspecial {
		return errors.New("a senha deve conter pelo menos um caractere especial (!@#$%...)")
	}

	return nil
}
