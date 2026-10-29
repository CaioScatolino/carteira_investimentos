package portfolio_test

import (
	"context"
	"math"
	"testing"
	"time"

	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/portfolio"
	"carteira_investimentos/server/internal/storage"
)

// mockRepo implementa portfolio.Repository em memória para testes unitários.
type mockRepo struct {
	posicoes   map[string]portfolio.Posicao
	transacoes []portfolio.Transacao
	proximoID  int64
}

func novoMockRepo() *mockRepo {
	return &mockRepo{
		posicoes:   make(map[string]portfolio.Posicao),
		transacoes: make([]portfolio.Transacao, 0),
		proximoID:  1,
	}
}

func (m *mockRepo) chave(usuarioID int64, ticker, corretora string) string {
	return string(rune(usuarioID)) + "_" + ticker + "_" + corretora
}

func (m *mockRepo) ListarPosicoes(ctx context.Context, usuarioID int64) ([]portfolio.Posicao, error) {
	var lista []portfolio.Posicao
	for _, p := range m.posicoes {
		if p.UsuarioID == usuarioID {
			lista = append(lista, p)
		}
	}
	return lista, nil
}

func (m *mockRepo) ObterPosicaoPorID(ctx context.Context, usuarioID, posicaoID int64) (*portfolio.Posicao, error) {
	for _, p := range m.posicoes {
		if p.UsuarioID == usuarioID && p.ID == posicaoID {
			cp := p
			return &cp, nil
		}
	}
	return nil, portfolio.ErrPosicaoNaoEncontrada
}

func (m *mockRepo) ObterPosicaoPorTicker(ctx context.Context, usuarioID int64, ticker, corretora string) (*portfolio.Posicao, error) {
	k := m.chave(usuarioID, ticker, corretora)
	p, ok := m.posicoes[k]
	if !ok {
		return nil, portfolio.ErrPosicaoNaoEncontrada
	}
	cp := p
	return &cp, nil
}

func (m *mockRepo) SalvarPosicao(ctx context.Context, pos *portfolio.Posicao) error {
	k := m.chave(pos.UsuarioID, pos.Ticker, pos.Corretora)
	if pos.ID == 0 {
		pos.ID = m.proximoID
		m.proximoID++
		pos.CreatedAt = time.Now()
	}
	pos.UpdatedAt = time.Now()
	m.posicoes[k] = *pos
	return nil
}

func (m *mockRepo) ExcluirPosicao(ctx context.Context, usuarioID, posicaoID int64) error {
	for k, p := range m.posicoes {
		if p.UsuarioID == usuarioID && p.ID == posicaoID {
			delete(m.posicoes, k)
			return nil
		}
	}
	return portfolio.ErrPosicaoNaoEncontrada
}

func (m *mockRepo) RegistrarTransacao(ctx context.Context, t *portfolio.Transacao) error {
	t.ID = int64(len(m.transacoes) + 1)
	m.transacoes = append(m.transacoes, *t)
	return nil
}

func (m *mockRepo) ListarTransacoes(ctx context.Context, usuarioID int64) ([]portfolio.Transacao, error) {
	var lista []portfolio.Transacao
	for _, t := range m.transacoes {
		if t.UsuarioID == usuarioID {
			lista = append(lista, t)
		}
	}
	return lista, nil
}

func (m *mockRepo) ListarTransacoesPorTicker(ctx context.Context, usuarioID int64, ticker string) ([]portfolio.Transacao, error) {
	var lista []portfolio.Transacao
	for _, t := range m.transacoes {
		if t.UsuarioID == usuarioID && t.Ticker == ticker {
			lista = append(lista, t)
		}
	}
	return lista, nil
}

func TestRecalculoPrecoMedioPonderado(t *testing.T) {
	repo := novoMockRepo()
	snapStore := storage.NovoInMemoryStore()
	_ = snapStore.Salvar(&domain.Ativo{
		Ticker:     "BBAS3",
		PrecoAtual: 25.0,
		Classe:     "ACAO",
	})

	svc := portfolio.NovoService(repo, snapStore)
	ctx := context.Background()
	usuarioID := int64(1001)

	// 1. Cadastro inicial: 100 cotas a R$ 20.00 (Exemplo do usuário)
	req1 := portfolio.CriarPosicaoRequest{
		Ticker:     "BBAS3",
		Classe:     portfolio.ClasseAcaoBR,
		Quantidade: 100.0,
		PrecoMedio: 20.0,
		Corretora:  "Principal",
	}

	pos1, err := svc.AdicionarOuAtualizarPosicao(ctx, usuarioID, req1)
	if err != nil {
		t.Fatalf("falha no cadastro inicial: %v", err)
	}

	if pos1.Quantidade != 100.0 {
		t.Errorf("esperava quantidade 100.0, obteve %.2f", pos1.Quantidade)
	}
	if pos1.PrecoMedio != 20.0 {
		t.Errorf("esperava PM 20.0, obteve %.2f", pos1.PrecoMedio)
	}
	if pos1.ValorAplicado != 2000.0 {
		t.Errorf("esperava valor aplicado 2000.0, obteve %.2f", pos1.ValorAplicado)
	}

	// 2. Novo Aporte: +10 cotas a R$ 18.00 (Exemplo do usuário)
	// Esperado: (100 * 20 + 10 * 18) / 110 = 2180 / 110 = 19.818181...
	req2 := portfolio.CriarPosicaoRequest{
		Ticker:     "BBAS3",
		Classe:     portfolio.ClasseAcaoBR,
		Quantidade: 10.0,
		PrecoMedio: 18.0,
		Corretora:  "Principal",
	}

	pos2, err := svc.AdicionarOuAtualizarPosicao(ctx, usuarioID, req2)
	if err != nil {
		t.Fatalf("falha no aporte: %v", err)
	}

	if pos2.Quantidade != 110.0 {
		t.Errorf("esperava quantidade total 110.0, obteve %.2f", pos2.Quantidade)
	}

	pmEsperado := 2180.0 / 110.0 // ~19.81818
	if math.Abs(pos2.PrecoMedio-pmEsperado) > 0.0001 {
		t.Errorf("esperava PM ponderado %.4f, obteve %.4f", pmEsperado, pos2.PrecoMedio)
	}

	if math.Abs(pos2.ValorAplicado-2180.0) > 0.0001 {
		t.Errorf("esperava valor aplicado total 2180.0, obteve %.2f", pos2.ValorAplicado)
	}

	// 3. Verifica o Livro-Razão Contábil (Transações)
	transacoes, err := svc.ListarTransacoes(ctx, usuarioID)
	if err != nil {
		t.Fatalf("falha ao listar transacoes: %v", err)
	}

	if len(transacoes) != 2 {
		t.Fatalf("esperava 2 transacoes registradas, obteve %d", len(transacoes))
	}

	// Transação 1: Compra inicial de 100 cotas a R$ 20.00
	t1 := transacoes[0]
	if t1.Tipo != portfolio.TransacaoCompra || t1.Quantidade != 100.0 || t1.PrecoUnitario != 20.0 {
		t.Errorf("transacao 1 incorreta: %+v", t1)
	}

	// Transação 2: Aporte de 10 cotas a R$ 18.00
	t2 := transacoes[1]
	if t2.Tipo != portfolio.TransacaoAporte || t2.Quantidade != 10.0 || t2.PrecoUnitario != 18.0 || t2.ValorTotal != 180.0 {
		t.Errorf("transacao 2 de aporte incorreta: %+v", t2)
	}

	// 4. Terceiro Aporte: +20 cotas a R$ 22.00
	// Esperado: (2180 + 20 * 22) / 130 = (2180 + 440) / 130 = 2620 / 130 = 20.1538...
	req3 := portfolio.CriarPosicaoRequest{
		Ticker:     "BBAS3",
		Classe:     portfolio.ClasseAcaoBR,
		Quantidade: 20.0,
		PrecoMedio: 22.0,
		Corretora:  "Principal",
	}

	pos3, err := svc.AdicionarOuAtualizarPosicao(ctx, usuarioID, req3)
	if err != nil {
		t.Fatalf("falha no terceiro aporte: %v", err)
	}

	if pos3.Quantidade != 130.0 {
		t.Errorf("esperava quantidade total 130.0, obteve %.2f", pos3.Quantidade)
	}

	pmEsperado3 := 2620.0 / 130.0
	if math.Abs(pos3.PrecoMedio-pmEsperado3) > 0.0001 {
		t.Errorf("esperava PM ponderado %.4f, obteve %.4f", pmEsperado3, pos3.PrecoMedio)
	}

	// 5. Teste de Substituição Explícita (Correção manual)
	reqSubstituir := portfolio.CriarPosicaoRequest{
		Ticker:       "BBAS3",
		Classe:       portfolio.ClasseAcaoBR,
		Quantidade:   50.0,
		PrecoMedio:   15.0,
		Corretora:    "Principal",
		TipoOperacao: "SUBSTITUIR",
	}

	posSub, err := svc.AdicionarOuAtualizarPosicao(ctx, usuarioID, reqSubstituir)
	if err != nil {
		t.Fatalf("falha ao substituir posicao: %v", err)
	}

	if posSub.Quantidade != 50.0 {
		t.Errorf("esperava quantidade 50.0 após substituição, obteve %.2f", posSub.Quantidade)
	}
	if posSub.PrecoMedio != 15.0 {
		t.Errorf("esperava PM 15.0 após substituição, obteve %.2f", posSub.PrecoMedio)
	}
}

func TestDecomposicaoHierarquicaCarteira(t *testing.T) {
	repo := novoMockRepo()
	snapStore := storage.NovoInMemoryStore()
	svc := portfolio.NovoService(repo, snapStore)
	ctx := context.Background()
	usuarioID := int64(2002)

	// Simula exatamente o exemplo do usuário:
	// Total: R$ 100.000
	// Renda Fixa: R$ 50.000 (50%)
	// FIIs: R$ 30.000 (30%)
	// Ações: R$ 20.000 (20%)
	//   - PETR4: R$ 10.000 (50% das Ações, 10% do Total)
	//   - VALE3: R$ 7.000 (35% das Ações, 7% do Total)
	//   - AGRO3: R$ 3.000 (15% das Ações, 3% do Total)

	posicoes := []portfolio.CriarPosicaoRequest{
		{Ticker: "TESOURO_IPCA", Classe: portfolio.ClasseRendaFixa, Quantidade: 50, PrecoMedio: 1000.0, Corretora: "XP"},
		{Ticker: "HGLG11", Classe: portfolio.ClasseFII, Quantidade: 200, PrecoMedio: 150.0, Corretora: "Clear"},
		{Ticker: "PETR4", Classe: portfolio.ClasseAcaoBR, Quantidade: 200, PrecoMedio: 50.0, Corretora: "Genial"},
		{Ticker: "VALE3", Classe: portfolio.ClasseAcaoBR, Quantidade: 100, PrecoMedio: 70.0, Corretora: "Genial"},
		{Ticker: "AGRO3", Classe: portfolio.ClasseAcaoBR, Quantidade: 300, PrecoMedio: 10.0, Corretora: "Genial"},
	}

	for _, req := range posicoes {
		if _, err := svc.AdicionarOuAtualizarPosicao(ctx, usuarioID, req); err != nil {
			t.Fatalf("falha ao adicionar %s: %v", req.Ticker, err)
		}
	}

	resumo, err := svc.ObterResumoCarteira(ctx, usuarioID)
	if err != nil {
		t.Fatalf("falha ao obter resumo: %v", err)
	}

	if math.Abs(resumo.PatrimonioTotal-100000.0) > 0.01 {
		t.Errorf("esperava patrimonio total de 100.000, obteve %.2f", resumo.PatrimonioTotal)
	}

	// Localiza as posições de ações
	var petr4, vale3, agro3 *portfolio.PosicaoConsolidada
	for i := range resumo.Posicoes {
		p := &resumo.Posicoes[i]
		if p.Ticker == "PETR4" {
			petr4 = p
		} else if p.Ticker == "VALE3" {
			vale3 = p
		} else if p.Ticker == "AGRO3" {
			agro3 = p
		}
	}

	if petr4 == nil || vale3 == nil || agro3 == nil {
		t.Fatalf("posicoes de acoes nao encontradas")
	}

	// Valida PETR4: R$ 10.000 => 50% das Ações, 10% do Total
	if math.Abs(petr4.PercentualNaClasse-50.0) > 0.01 {
		t.Errorf("PETR4 esperava 50%% na classe, obteve %.2f%%", petr4.PercentualNaClasse)
	}
	if math.Abs(petr4.PercentualNaCarteira-10.0) > 0.01 {
		t.Errorf("PETR4 esperava 10%% na carteira, obteve %.2f%%", petr4.PercentualNaCarteira)
	}

	// Valida VALE3: R$ 7.000 => 35% das Ações, 7% do Total
	if math.Abs(vale3.PercentualNaClasse-35.0) > 0.01 {
		t.Errorf("VALE3 esperava 35%% na classe, obteve %.2f%%", vale3.PercentualNaClasse)
	}
	if math.Abs(vale3.PercentualNaCarteira-7.0) > 0.01 {
		t.Errorf("VALE3 esperava 7%% na carteira, obteve %.2f%%", vale3.PercentualNaCarteira)
	}

	// Valida AGRO3: R$ 3.000 => 15% das Ações, 3% do Total
	if math.Abs(agro3.PercentualNaClasse-15.0) > 0.01 {
		t.Errorf("AGRO3 esperava 15%% na classe, obteve %.2f%%", agro3.PercentualNaClasse)
	}
	if math.Abs(agro3.PercentualNaCarteira-3.0) > 0.01 {
		t.Errorf("AGRO3 esperava 3%% na carteira, obteve %.2f%%", agro3.PercentualNaCarteira)
	}

	// Valida o agrupamento hierárquico em Classes
	if len(resumo.Classes) != 5 {
		t.Errorf("esperava 5 classes consolidadas, obteve %d", len(resumo.Classes))
	}

	// Ações Brasil deve conter 3 ativos e R$ 20.000
	var classeAcoes *portfolio.ClasseConsolidada
	for i := range resumo.Classes {
		if resumo.Classes[i].Classe == portfolio.ClasseAcaoBR {
			classeAcoes = &resumo.Classes[i]
			break
		}
	}

	if classeAcoes == nil {
		t.Fatalf("classe Ações Brasil não encontrada")
	}
	if classeAcoes.QuantidadeAtivos != 3 {
		t.Errorf("esperava 3 ativos em ações, obteve %d", classeAcoes.QuantidadeAtivos)
	}
	if math.Abs(classeAcoes.ValorAtualMercado-20000.0) > 0.01 {
		t.Errorf("esperava R$ 20.000 em ações, obteve %.2f", classeAcoes.ValorAtualMercado)
	}
	if math.Abs(classeAcoes.PercentualNaCarteira-20.0) > 0.01 {
		t.Errorf("esperava 20%% na carteira, obteve %.2f%%", classeAcoes.PercentualNaCarteira)
	}
	// O primeiro ativo da classe ordenada deve ser PETR4 (maior valor)
	if classeAcoes.Posicoes[0].Ticker != "PETR4" {
		t.Errorf("esperava primeiro ativo ser PETR4, obteve %s", classeAcoes.Posicoes[0].Ticker)
	}
}

