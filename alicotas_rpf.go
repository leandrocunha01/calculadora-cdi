package main

func aliquotaIRPFRendaFixa(meses int) float64 {
	switch {
	case meses <= 6:
		return 22.5
	case meses <= 12:
		return 20
	case meses <= 24:
		return 17.5
	default:
		return 15
	}
}

func calcularIR(valorJuros float64, meses int) float64 {
	return valorJuros * aliquotaIRPFRendaFixa(meses) / 100
}
