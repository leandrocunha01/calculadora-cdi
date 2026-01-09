package main

import "time"

type Simulation struct {
	DataExecucao   time.Time
	Inicial        float64
	Aporte         float64
	TaxaAnual      float64
	Meses          int
	AporteNoInicio bool
	Evolucao       []Mes
	TotalInvestido float64
	TotalJuros     float64
	SaldoFinal     float64

	CDIInput        string
	PercentualInput string
	FutureDateInput string
	PrazoInput      string
	InicialInput    string
	AporteInput     string
}

type Mes struct {
    Mes             int     `json:"mes"`
    SaldoInicial    float64 `json:"saldoInicial"`
    Aporte          float64 `json:"aporte"`
    Rendimento      float64 `json:"rendimento"`
    JurosAcumulados float64 `json:"jurosAcumulados"`
    SaldoFinal      float64 `json:"saldoFinal"`
    DataAporte      string  `json:"dataAporte"`
}
