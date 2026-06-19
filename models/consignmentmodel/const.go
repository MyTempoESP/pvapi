package consignmentmodel

// Consignment lifecycle statuses.
// Active (visible in public listings): LIBERADO.
// Inactive (hidden from public listings): RASCUNHO, PAUSADO, ARQUIVADO.
const (
	ComissionPercent = "PERCENT"
	ComissionFixed   = "FIXED"

	StatusRascunho  = "RASCUNHO"
	StatusLiberado  = "LIBERADO"
	StatusPausado   = "PAUSADO"
	StatusArquivado = "ARQUIVADO"
)
