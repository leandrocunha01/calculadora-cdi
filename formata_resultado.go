package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseFloat(v string) float64 {
	v = strings.ReplaceAll(v, ".", "")
	v = strings.ReplaceAll(v, ",", ".")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func formatBRL(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := parts[1]

	n := len(intPart)
	var sb strings.Builder
	for i, c := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			sb.WriteRune('.')
		}
		sb.WriteRune(c)
	}

	return "R$ " + sb.String() + "," + decPart
}

func formatMatricial(evolucao []Mes, inicial float64, totalInvestido, totalJurosBruto, saldoFinalBruto float64, meses int, aporteNoInicio bool) string {
	var sb strings.Builder
	sb.WriteString("==============================================\n")
	sb.WriteString("          RELATÓRIO DE INVESTIMENTO          \n")
	sb.WriteString("==============================================\n")
	sb.WriteString(fmt.Sprintf("Período         : %d meses\n", meses))
	sb.WriteString(fmt.Sprintf("Total investido : R$ %.2f\n", totalInvestido))
	sb.WriteString(fmt.Sprintf("Total rendimento: R$ %.2f\n", totalJurosBruto))
	sb.WriteString(fmt.Sprintf("Saldo final     : R$ %.2f\n", saldoFinalBruto))
	sb.WriteString("==============================================\n")
	sb.WriteString("Mês | Saldo Inicial | Aporte Mensal | Rendimento | Saldo Final\n")
	sb.WriteString("----|---------------|---------------|------------|------------\n")

	saldoAcumulado := inicial
	for _, m := range evolucao {
		jurosLiquido := m.Rendimento
		saldoFinal := saldoAcumulado
		if aporteNoInicio {
			saldoFinal += m.Aporte
		}
		saldoFinal += jurosLiquido
		if !aporteNoInicio {
			saldoFinal += m.Aporte
		}
		sb.WriteString(fmt.Sprintf(
			"%3d | %13.2f | %13.2f | %10.2f | %10.2f\n",
			m.Mes,
			m.SaldoInicial,
			m.Aporte,
			m.Rendimento,
			saldoFinal,
		))
		saldoAcumulado = saldoFinal
	}
	sb.WriteString("==============================================\n")
	sb.WriteString("FIM DO RELATÓRIO\n")
	sb.WriteString("==============================================\n")
	return sb.String()
}
