package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// Carregar lê o arquivo .env (se existir) e preenche as variáveis de configuração
func Carregar() *Config {
	carregarArquivoEnv(".env")

	cfg := &Config{
		DBHost:     obterEnv("DB_HOST", "127.0.0.1"),
		DBPort:     obterEnv("DB_PORT", "3306"),
		DBUser:     obterEnv("DB_USER", "root"),
		DBPassword: obterEnv("DB_PASSWORD", ""),
		DBName:     obterEnv("DB_NAME", "carteira_investimentos"),
	}

	return cfg
}

func obterEnv(chave, padrao string) string {
	if valor, existe := os.LookupEnv(chave); existe {
		return valor
	}
	return padrao
}

// carregarArquivoEnv faz o parse nativo do .env sem bibliotecas externas
func carregarArquivoEnv(caminho string) {
	file, err := os.Open(caminho)
	if err != nil {
		return // Se não existir, usa as variáveis de ambiente do SO
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue
		}

		partes := strings.SplitN(linha, "=", 2)
		if len(partes) == 2 {
			chave := strings.TrimSpace(partes[0])
			valor := strings.TrimSpace(partes[1])
			// Se o valor estiver entre aspas, remove
			valor = strings.Trim(valor, `"'`)
			os.Setenv(chave, valor)
		}
	}
}
