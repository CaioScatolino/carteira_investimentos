package auth

import "time"

// Usuario representa a entidade de segurança persistida no MySQL
type Usuario struct {
	ID               int64      `json:"id"`
	UUID             string     `json:"uuid"`
	Nome             string     `json:"nome"`
	Email            string     `json:"email"`
	SenhaHash        string     `json:"-"` // NUNCA exposto em respostas JSON
	Role             string     `json:"role"`
	Status           string     `json:"status"`
	TentativasFalhas int        `json:"-"`
	BloqueadoAte     *time.Time `json:"-"`
	UltimoLoginAt    *time.Time `json:"ultimo_login_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// UsuarioPublico é a projeção segura enviada ao frontend
type UsuarioPublico struct {
	UUID  string `json:"uuid"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (u *Usuario) ParaPublico() *UsuarioPublico {
	return &UsuarioPublico{
		UUID:  u.UUID,
		Nome:  u.Nome,
		Email: u.Email,
		Role:  u.Role,
	}
}

// DTOs de Entrada e Saída
type RequisicaoRegistro struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type RequisicaoLogin struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type RespostaAuth struct {
	Token    string          `json:"token"`
	Tipo     string          `json:"tipo"` // "Bearer"
	ExpiraEm int64           `json:"expira_em"`
	Usuario  *UsuarioPublico `json:"usuario"`
}
