package main

import (
	"math"
)

func calcularEvolucao(inicial, aporte, taxaAnual float64, meses int, aporteNoInicio bool) ([]Mes, float64, float64, float64) {
	taxaMensal := math.Pow(1+taxaAnual, 1.0/12.0) - 1
	saldo := inicial
	totalInvestido := inicial
	totalJuros := 0.0
	var evolucao []Mes

	for m := 1; m <= meses; m++ {
		saldoInicial := saldo
		if aporteNoInicio {
			saldo += aporte
			totalInvestido += aporte
		}
		rendimento := saldo * taxaMensal
		totalJuros += rendimento
		saldo += rendimento
		if !aporteNoInicio {
			saldo += aporte
			totalInvestido += aporte
		}
		evolucao = append(evolucao, Mes{
			Mes:             m,
			SaldoInicial:    saldoInicial,
			Aporte:          aporte,
			Rendimento:      rendimento,
			JurosAcumulados: totalJuros,
			SaldoFinal:      saldo,
		})
	}

	return evolucao, totalInvestido, totalJuros, saldo
}
