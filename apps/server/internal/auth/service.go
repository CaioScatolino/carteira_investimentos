package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service orquestra as regras de negócio de autenticação
type Service struct {
	repo         Repository
	tokenManager *GerenciadorToken
}

// NovoService instancia o serviço de autenticação
func NovoService(repo Repository, tokenManager *GerenciadorToken) *Service {
	return &Service{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

// Registrar valida, cria o hash e cadastra um novo usuário
func (s *Service) Registrar(ctx context.Context, req *RequisicaoRegistro) (*RespostaAuth, error) {
	req.Nome = strings.TrimSpace(req.Nome)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Nome == "" {
		return nil, errors.New("o nome é obrigatório")
	}
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return nil, errors.New("e-mail inválido")
	}

	// 1. Validação de força de senha (OWASP)
	if err := ValidarForcaSenha(req.Senha); err != nil {
		return nil, err
	}

	// 2. Verifica se o e-mail já existe
	existente, err := s.repo.BuscarPorEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existente != nil {
		return nil, errors.New("este e-mail já está cadastrado na plataforma")
	}

	// 3. Gera hash bcrypt seguro (custo 12)
	hash, err := GerarHashSenha(req.Senha)
	if err != nil {
		return nil, fmt.Errorf("falha ao processar senha: %w", err)
	}

	novoUsuario := &Usuario{
		UUID:      uuid.NewString(),
		Nome:      req.Nome,
		Email:     req.Email,
		SenhaHash: hash,
		Role:      "USER",
		Status:    "ATIVO",
	}

	// 4. Salva no banco de dados
	if err := s.repo.Criar(ctx, novoUsuario); err != nil {
		return nil, err
	}

	// 5. Emite o token JWT para login automático pós-cadastro
	token, expiraEm, err := s.tokenManager.GerarToken(novoUsuario)
	if err != nil {
		return nil, fmt.Errorf("usuário criado, mas falhou ao emitir token: %w", err)
	}

	return &RespostaAuth{
		Token:    token,
		Tipo:     "Bearer",
		ExpiraEm: expiraEm,
		Usuario:  novoUsuario.ParaPublico(),
	}, nil
}

// Login autentica o usuário com proteção contra força bruta
func (s *Service) Login(ctx context.Context, req *RequisicaoLogin) (*RespostaAuth, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	u, err := s.repo.BuscarPorEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	// Regra Anti-Enumeração: mensagem idêntica se o usuário não existir
	if u == nil {
		return nil, errors.New("e-mail ou senha incorretos")
	}

	// Verifica se a conta está congelada por força bruta
	if u.BloqueadoAte != nil && u.BloqueadoAte.After(time.Now()) {
		minutosRestantes := int(time.Until(*u.BloqueadoAte).Minutes()) + 1
		return nil, fmt.Errorf("conta temporariamente bloqueada por excesso de tentativas incorretas. Tente novamente em %d minutos", minutosRestantes)
	}

	// Compara a senha em tempo constante
	if !VerificarSenha(u.SenhaHash, req.Senha) {
		novasFalhas := u.TentativasFalhas + 1
		var novoBloqueio *time.Time

		// Ao atingir 5 falhas seguidas, bloqueia por 15 minutos
		if novasFalhas >= 5 {
			bloqueio := time.Now().Add(15 * time.Minute)
			novoBloqueio = &bloqueio
		}

		_ = s.repo.AtualizarTentativasFalhas(ctx, u.ID, novasFalhas, novoBloqueio)

		if novasFalhas >= 5 {
			return nil, errors.New("conta bloqueada por 15 minutos devido a 5 tentativas incorretas consecutivas")
		}
		return nil, errors.New("e-mail ou senha incorretos")
	}

	// Senha correta: zera contadores de falhas e registra o login
	_ = s.repo.RegistrarLoginSucesso(ctx, u.ID)

	// Emite o token JWT
	token, expiraEm, err := s.tokenManager.GerarToken(u)
	if err != nil {
		return nil, fmt.Errorf("falha ao emitir token de acesso: %w", err)
	}

	return &RespostaAuth{
		Token:    token,
		Tipo:     "Bearer",
		ExpiraEm: expiraEm,
		Usuario:  u.ParaPublico(),
	}, nil
}
