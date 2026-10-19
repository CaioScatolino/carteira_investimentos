export interface UsuarioPublico {
  id?: number;
  uuid: string;
  nome: string;
  email: string;
  role: string;
}

export interface RespostaAuth {
  token: string;
  tipo: string;
  expira_em: number;
  usuario: UsuarioPublico;
}

export interface RequisicaoRegistro {
  nome: string;
  email: string;
  senha: string;
}

export interface RequisicaoLogin {
  email: string;
  senha: string;
}
