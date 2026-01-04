package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

/* =========================
   MODELO FINANCEIRO
========================= */

type Mes struct {
	Mes             int
	SaldoInicial    float64
	Aporte          float64
	Rendimento      float64
	JurosAcumulados float64
	SaldoFinal      float64
}

func calcularEvolucao(
	inicial, aporte, taxaAnual float64,
	meses int,
	aporteNoInicio bool,
) ([]Mes, float64, float64, float64) {

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

/* =========================
   UTIL
========================= */

func parseFloat(v string) float64 {
	v = strings.ReplaceAll(v, ",", ".")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

/* =========================
   MAIN
========================= */

func main() {
	app := tview.NewApplication()

	/* ---------- FORM ESQUERDO ---------- */

	leftForm := tview.NewForm()
	leftForm.SetBorder(true)
	leftForm.SetTitle("+ Parâmetros F2 +")

	leftForm.AddInputField("Data futura (DD/MM/AAAA)", "", 20, nil, nil)
	leftForm.AddInputField("CDI anual (%)", "", 10, nil, nil)
	leftForm.AddInputField("Percentual do CDI (%)", "", 10, nil, nil)

	/* ---------- FORM DIREITO ---------- */

	rightForm := tview.NewForm()
	rightForm.SetBorder(true)
	rightForm.SetTitle("+ Valores F3 +")

	rightForm.AddInputField("Valor inicial (R$)", "", 12, nil, nil)
	rightForm.AddInputField("Aporte mensal (R$)", "", 12, nil, nil)

	aporteNoInicio := true
	rightForm.AddCheckbox("Aporte no início do mês", true, func(v bool) {
		aporteNoInicio = v
	})

	/* ---------- RESULTADO ---------- */

	result := tview.NewTextView()
	result.SetDynamicColors(true)
	result.SetScrollable(true)
	result.SetWrap(false)
	result.SetBorder(true)
	result.SetTitle("+ Resultado F5 para calcular | Esc para sair +")

	scrollRow := 0
	totalLines := 0

	/* ---------- FUNÇÃO DE CÁLCULO ---------- */

	calculate := func() {
		layout := "02/01/2006"
		now := time.Now()

		future, err := time.Parse(layout, leftForm.GetFormItem(0).(*tview.InputField).GetText())
		if err != nil || !future.After(now) {
			result.SetText("[red]Data inválida")
			return
		}

		cdi := parseFloat(leftForm.GetFormItem(1).(*tview.InputField).GetText()) / 100
		perc := parseFloat(leftForm.GetFormItem(2).(*tview.InputField).GetText()) / 100
		inicial := parseFloat(rightForm.GetFormItem(0).(*tview.InputField).GetText())
		aporte := parseFloat(rightForm.GetFormItem(1).(*tview.InputField).GetText())

		taxaAnual := cdi * perc

		meses := (future.Year()-now.Year())*12 + int(future.Month()-now.Month())
		if future.Day() < now.Day() {
			meses--
		}
		if meses < 0 {
			meses = 0
		}

		evolucao, investido, juros, final := calcularEvolucao(
			inicial,
			aporte,
			taxaAnual,
			meses,
			aporteNoInicio,
		)

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Período         : %d meses\n", meses))
		sb.WriteString(fmt.Sprintf("Total investido : R$ %.2f\n", investido))
		sb.WriteString(fmt.Sprintf("Total juros     : R$ %.2f\n", juros))
		sb.WriteString(fmt.Sprintf("Resultado       : R$ %.2f\n", final-investido))
		sb.WriteString(fmt.Sprintf("Valor final     : [green]R$ %.2f[-]\n", final))

		sb.WriteString("\n")

		sb.WriteString("Mes | Saldo Inicial | Aporte | Juros Mes | Juros Acum. | Saldo Final\n")
		sb.WriteString("----|---------------|--------|-----------|-------------|------------\n")

		for _, e := range evolucao {
			sb.WriteString(fmt.Sprintf(
				"%3d | %13.2f | %6.2f | %9.2f | %11.2f | %10.2f\n",
				e.Mes,
				e.SaldoInicial,
				e.Aporte,
				e.Rendimento,
				e.JurosAcumulados,
				e.SaldoFinal,
			))
		}

		sb.WriteString("\n")

		text := sb.String()
		result.SetText(text)
		result.ScrollToBeginning()

		scrollRow = 0
		totalLines = strings.Count(text, "\n")
	}

	/* ---------- CONTROLE DE FOCO ENTRE FORMS ---------- */

	// TAB no último campo do form esquerdo → vai para o direito
	leftForm.GetFormItem(2).(*tview.InputField).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyTab || key == tcell.KeyEnter {
				app.SetFocus(rightForm)
			}
		})

	/* ---------- LAYOUT ---------- */

	forms := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(leftForm, 0, 1, true).
		AddItem(rightForm, 0, 1, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(forms, 0, 1, true).
		AddItem(result, 0, 2, false)

	/* ---------- ATALHOS GLOBAIS ---------- */

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyF2:
			app.SetFocus(leftForm)
			return nil

		case tcell.KeyF3:
			app.SetFocus(rightForm)
			return nil

		case tcell.KeyUp:
			if scrollRow > 0 {
				scrollRow--
				result.ScrollTo(scrollRow, 0)
			}
			return nil

		case tcell.KeyDown:
			if scrollRow < totalLines-1 {
				scrollRow++
				result.ScrollTo(scrollRow, 0)
			}
			return nil

		case tcell.KeyPgUp:
			scrollRow -= 10
			if scrollRow < 0 {
				scrollRow = 0
			}
			result.ScrollTo(scrollRow, 0)
			return nil

		case tcell.KeyPgDn:
			scrollRow += 10
			if scrollRow > totalLines-1 {
				scrollRow = totalLines - 1
			}
			result.ScrollTo(scrollRow, 0)
			return nil

		case tcell.KeyHome:
			scrollRow = 0
			result.ScrollToBeginning()
			return nil

		case tcell.KeyEnd:
			scrollRow = totalLines - 1
			result.ScrollToEnd()
			return nil

		case tcell.KeyEsc:
			app.Stop()
			return nil
		case tcell.KeyF5:
			calculate()
			return nil
		}

		return event
	})

	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
	}
}
