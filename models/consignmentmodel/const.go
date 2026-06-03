package consignmentmodel

// Consignment lifecycle statuses.
// Active statuses (visible in public listings): EM TRANSITO, RECEBIDO, CONFERIDO.
// Inactive statuses (hidden from public listings): RASCUNHO, ARQUIVADO.
const (
	ComissionPercent = "PERCENT"
	ComissionFixed   = "FIXED"

	StatusRascunho  = "RASCUNHO"
	StatusTransito  = "EM TRANSITO"
	StatusRecebido  = "RECEBIDO"
	StatusConferido = "CONFERIDO"
	StatusArquivado = "ARQUIVADO"
)
