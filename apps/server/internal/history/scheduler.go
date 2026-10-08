package history

import (
	"context"
	"fmt"
	"sort"
	"time"

	"carteira_investimentos/server/internal/domain"
	"carteira_investimentos/server/internal/storage"
)

// Scheduler gerencia o agendamento pós-fechamento do pregão da B3
type Scheduler struct {
	repo  Repository
	store storage.SnapshotStore
}

// NovoScheduler instancia o agendador
func NovoScheduler(repo Repository, store storage.SnapshotStore) *Scheduler {
	return &Scheduler{
		repo:  repo,
		store: store,
	}
}

// ProximoHorarioPregao calcula o próximo dia útil às 19:00:00 (pula Sábado e Domingo)
func ProximoHorarioPregao(agora time.Time) time.Time {
	loc := agora.Location()
	// Define o alvo de hoje às 19:00:00
	alvo := time.Date(agora.Year(), agora.Month(), agora.Day(), 19, 0, 0, 0, loc)

	// Se já passou das 19h de hoje, agenda para o dia seguinte
	if !agora.Before(alvo) {
		alvo = alvo.Add(24 * time.Hour)
	}

	// Se cair em fim de semana, avança até a próxima segunda-feira
	for alvo.Weekday() == time.Saturday || alvo.Weekday() == time.Sunday {
		alvo = alvo.Add(24 * time.Hour)
	}

	return alvo
}

// ExecutarSnapshotAgora consolida os rankings da memória e grava na base MySQL
func (s *Scheduler) ExecutarSnapshotAgora(ctx context.Context, dataReferencia time.Time) error {
	if s.repo == nil {
		return fmt.Errorf("repositório MySQL não configurado (modo resiliente sem banco)")
	}

	todosAtivos := s.store.ListarTodos()
	if len(todosAtivos) == 0 {
		return fmt.Errorf("nenhum ativo encontrado no cache em memória para gerar snapshot")
	}

	// 1. Separa por classe (Ações, FIIs, ETFs)
	var acoes, fiis, etfs []*domain.Ativo
	for _, a := range todosAtivos {
		switch a.Classe {
		case domain.ClasseFII:
			fiis = append(fiis, a)
		case domain.ClasseETF:
			etfs = append(etfs, a)
		default:
			acoes = append(acoes, a)
		}
	}

	// 2. Ordena por Score decrescente (desempate por Volume)
	ordenar := func(slice []*domain.Ativo) {
		sort.Slice(slice, func(i, j int) bool {
			if slice[i].Score == slice[j].Score {
				return slice[i].VolumeTotal > slice[j].VolumeTotal
			}
			return slice[i].Score > slice[j].Score
		})
	}

	ordenar(acoes)
	ordenar(fiis)
	ordenar(etfs)

	// 3. Converte para registros históricos numerando a posição no ranking (1º, 2º, ...)
	var registros []*RegistroHistorico

	for i, a := range acoes {
		registros = append(registros, ConverterAtivoParaRegistro(dataReferencia, i+1, a))
	}
	for i, f := range fiis {
		registros = append(registros, ConverterAtivoParaRegistro(dataReferencia, i+1, f))
	}
	for i, e := range etfs {
		registros = append(registros, ConverterAtivoParaRegistro(dataReferencia, i+1, e))
	}

	// 4. Grava em lote atômico no MySQL
	inicio := time.Now()
	if err := s.repo.SalvarSnapshotLote(ctx, registros); err != nil {
		return fmt.Errorf("falha ao persistir snapshot: %w", err)
	}

	dataStr := dataReferencia.Format("2006-01-02")
	fmt.Printf("\n💾 [CRON 19H] Snapshot diário gravado com sucesso no MySQL! (%s)\n", dataStr)
	fmt.Printf("   • Pregão: %s | Tempo de gravação: %v\n", dataStr, time.Since(inicio))
	fmt.Printf("   • Ativos persistidos: %d (Ações: %d | FIIs: %d | ETFs: %d)\n\n",
		len(registros), len(acoes), len(fiis), len(etfs))

	return nil
}

// IniciarRotinaDiaria inicia a goroutine perpétua agendada para as 19h
func (s *Scheduler) IniciarRotinaDiaria() {
	go func() {
		for {
			agora := time.Now()
			proximaExecucao := ProximoHorarioPregao(agora)
			espera := time.Until(proximaExecucao)

			fmt.Printf("🕒 [CRON 19H] Próximo snapshot pós-pregão agendado para: %s (em %v)\n",
				proximaExecucao.Format("02/01/2006 15:04:05"), espera.Round(time.Minute))

			// Aguarda até o horário exato
			timer := time.NewTimer(espera)
			<-timer.C

			// Executa o snapshot com timeout de 30 segundos
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := s.ExecutarSnapshotAgora(ctx, time.Now())
			cancel()

			if err != nil {
				fmt.Printf("⚠️ [CRON 19H] Erro na execução do snapshot diário: %v\n", err)
			}
		}
	}()
}
