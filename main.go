package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
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

/* =========================
   CÁLCULO
========================= */

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

/* =========================
   UTIL
========================= */

func parseFloat(v string) float64 {
	v = strings.ReplaceAll(v, ".", "")
	v = strings.ReplaceAll(v, ",", ".")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

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

/* =========================
   MAIN
========================= */

func main() {
	app := tview.NewApplication()
	pages := tview.NewPages()
	var lastSimulation *Simulation
	var layout *tview.Flex

	/* ---------- FORM ESQUERDO ---------- */
	leftForm := tview.NewForm()
	leftForm.SetBorder(true)
	leftForm.SetTitle("+ Parâmetros F2 +")
	leftForm.AddInputField("Data futura (DD/MM/AAAA)", "", 20, nil, nil)
	leftForm.AddInputField("CDI anual (%)", "", 10, nil, nil)
	leftForm.AddInputField("Percentual do CDI (%)", "", 10, nil, nil)

	prazoField := tview.NewInputField().
		SetLabel("Prazo (meses)").
		SetText("").
		SetFieldWidth(6)
	leftForm.AddFormItem(prazoField)

	irpfField := tview.NewInputField().
		SetLabel("IRPF (%)").
		SetText("0.0").
		SetFieldWidth(6).
		SetAcceptanceFunc(func(text string, lastChar rune) bool {
			return false
		})

	/* ---------- FORM DIREITO ---------- */
	rightForm := tview.NewForm()
	rightForm.SetBorder(true)
	rightForm.SetTitle("+ Valores F3 +")
	rightForm.AddInputField("Valor inicial (R$)", "", 12, nil, nil)
	rightForm.AddInputField("Aporte mensal (R$)", "", 12, nil, nil)

	aporteNoInicio := true
	rightForm.AddCheckbox("Aporte no início do mês", true, func(v bool) { aporteNoInicio = v })

	/* ---------- RESULTADO ---------- */
	result := tview.NewTextView()
	result.SetDynamicColors(true)
	result.SetScrollable(true)
	result.SetWrap(false)
	result.SetBorder(true)
	result.SetTitle("+ Resultado F5 calcular | F9 carregar | F10 salvar | Esc sair +")

	scrollRow, scrollCol := 0, 0
	totalLines, maxCols := 0, 0

	/* ---------- FUNÇÃO CALCULAR ---------- */
	calculate := func() {
		layoutDate := "02/01/2006"
		now := time.Now()

		cdi := parseFloat(leftForm.GetFormItem(1).(*tview.InputField).GetText()) / 100
		perc := parseFloat(leftForm.GetFormItem(2).(*tview.InputField).GetText()) / 100
		inicial := parseFloat(rightForm.GetFormItem(0).(*tview.InputField).GetText())
		aporte := parseFloat(rightForm.GetFormItem(1).(*tview.InputField).GetText())

		taxaAnual := cdi * perc

		// Lógica Data <-> Prazo
		var meses int
		prazo := parseFloat(prazoField.GetText())
		futureText := leftForm.GetFormItem(0).(*tview.InputField).GetText()
		future, err := time.Parse(layoutDate, futureText)

		if err == nil && futureText != "" {
			// Data preenchida -> calcula prazo
			meses = (future.Year()-now.Year())*12 + int(future.Month()-now.Month())
			if future.Day() < now.Day() {
				meses--
			}
			if meses < 0 {
				meses = 0
			}
			prazoField.SetText(fmt.Sprintf("%d", meses))
		} else if prazo > 0 {
			// Prazo preenchido -> calcula data
			meses = int(prazo)
			futureCalc := now.AddDate(0, meses, 0)
			leftForm.GetFormItem(0).(*tview.InputField).SetText(futureCalc.Format(layoutDate))
		} else {
			result.SetText("[red]Preencha a data ou o prazo!")
			return
		}

		evolucao, totalInvestido, totalJurosBruto, saldoFinalBruto := calcularEvolucao(inicial, aporte, taxaAnual, meses, aporteNoInicio)

		aliquota := aliquotaIRPFRendaFixa(meses)
		irpfField.SetText(fmt.Sprintf("%.1f", aliquota))

		lastSimulation = &Simulation{
			DataExecucao:    time.Now(),
			Inicial:         inicial,
			Aporte:          aporte,
			TaxaAnual:       taxaAnual,
			Meses:           meses,
			AporteNoInicio:  aporteNoInicio,
			Evolucao:        evolucao,
			TotalInvestido:  totalInvestido,
			TotalJuros:      totalJurosBruto,
			SaldoFinal:      saldoFinalBruto,
			CDIInput:        leftForm.GetFormItem(1).(*tview.InputField).GetText(),
			PercentualInput: leftForm.GetFormItem(2).(*tview.InputField).GetText(),
			FutureDateInput: leftForm.GetFormItem(0).(*tview.InputField).GetText(),
			PrazoInput:      prazoField.GetText(),
			InicialInput:    rightForm.GetFormItem(0).(*tview.InputField).GetText(),
			AporteInput:     rightForm.GetFormItem(1).(*tview.InputField).GetText(),
		}

		// Monta resultado
		var sb strings.Builder
		totalIRPF, totalJurosLiquido := 0.0, 0.0
		for _, e := range evolucao {
			irMes := calcularIR(e.Rendimento, meses)
			totalIRPF += irMes
			totalJurosLiquido += e.Rendimento - irMes
		}
		resultadoLiquido := totalInvestido + totalJurosLiquido

		sb.WriteString(fmt.Sprintf("Período             : %d meses\n", meses))
		sb.WriteString(fmt.Sprintf("Total investido     : %s\n", formatBRL(totalInvestido)))
		sb.WriteString(fmt.Sprintf("Total de rendimentos: %s\n", formatBRL(totalJurosBruto)))
		sb.WriteString(fmt.Sprintf("Total IRPF          : %s\n", formatBRL(totalIRPF)))
		sb.WriteString(fmt.Sprintf("Resultado líquido   : [green]%s[-]\n", formatBRL(resultadoLiquido)))
		sb.WriteString(fmt.Sprintf("Valor final bruto   : [green]%s[-]\n\n", formatBRL(saldoFinalBruto)))

		sb.WriteString("Mês |  Saldo Inicial  |  Aporte Mensal  |    Rendimento   |    IRPF Mês     | Rendimento Líqu | Saldo Líquido  \n")
		sb.WriteString("----|-----------------|-----------------|-----------------|-----------------|-----------------|----------------\n")

		saldoAcumulado := inicial
		for _, e := range evolucao {
			irMes := calcularIR(e.Rendimento, meses)
			jurosLiquido := e.Rendimento - irMes
			saldoFinal := saldoAcumulado
			if aporteNoInicio {
				saldoFinal += e.Aporte
			}
			saldoFinal += jurosLiquido
			if !aporteNoInicio {
				saldoFinal += e.Aporte
			}

			sb.WriteString(fmt.Sprintf(
				"%3d | %15s | %15s | %15s | %15s | %15s | %15s\n",
				e.Mes,
				formatBRL(e.SaldoInicial),
				formatBRL(e.Aporte),
				formatBRL(e.Rendimento),
				formatBRL(irMes),
				formatBRL(jurosLiquido),
				formatBRL(saldoFinal),
			))
			saldoAcumulado = saldoFinal
		}
		text := sb.String()
		result.SetText(text)
		result.ScrollToBeginning()
		scrollRow, scrollCol = 0, 0
		totalLines, maxCols = strings.Count(text, "\n"), 0
		for _, line := range strings.Split(text, "\n") {
			if len(line) > maxCols {
				maxCols = len(line)
			}
		}
	}

	/* ---------- FUNÇÕES SALVAR/CARREGAR ---------- */
	saveSimulation := func() {
		if lastSimulation == nil {
			return
		}
		input := tview.NewInputField().SetLabel("Nome da simulação: ")
		form := tview.NewForm().
			AddFormItem(input).
			AddButton("Salvar", func() {
				name := strings.TrimSpace(input.GetText())
				if name == "" {
					return
				}
				data, _ := json.MarshalIndent(lastSimulation, "", "  ")
				_ = os.WriteFile(name+"_simulation.json", data, 0644)
				pages.RemovePage("save")
				app.SetFocus(layout)
			}).
			AddButton("Cancelar", func() {
				pages.RemovePage("save")
				app.SetFocus(layout)
			})
		form.SetBorder(true).SetTitle("Salvar simulação")
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(form, 50, 1, true).
			AddItem(nil, 0, 1, false)
		pages.AddPage("save", modal, true, true)
		app.SetFocus(form)
	}

	loadSimulation := func() {
		files, err := os.ReadDir(".")
		if err != nil {
			result.SetText("[red]Erro ao ler diretório: " + err.Error())
			return
		}
		var options []string
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), "_simulation.json") {
				options = append(options, f.Name())
			}
		}
		if len(options) == 0 {
			result.SetText("[yellow]Nenhuma simulação encontrada!")
			return
		}
		selectedFile := options[0]
		dropdown := tview.NewDropDown().
			SetLabel("Selecione a simulação: ").
			SetOptions(options, func(text string, index int) { selectedFile = text })
		form := tview.NewForm().
			AddFormItem(dropdown).
			AddButton("Carregar", func() {
				data, err := os.ReadFile(selectedFile)
				if err != nil {
					result.SetText("[red]Erro ao carregar arquivo: " + err.Error())
					pages.RemovePage("load")
					app.SetFocus(layout)
					return
				}
				var sim Simulation
				if err := json.Unmarshal(data, &sim); err != nil {
					result.SetText("[red]Erro ao decodificar JSON: " + err.Error())
					pages.RemovePage("load")
					app.SetFocus(layout)
					return
				}
				lastSimulation = &sim
				leftForm.GetFormItem(0).(*tview.InputField).SetText(sim.FutureDateInput)
				leftForm.GetFormItem(1).(*tview.InputField).SetText(sim.CDIInput)
				leftForm.GetFormItem(2).(*tview.InputField).SetText(sim.PercentualInput)
				prazoField.SetText(sim.PrazoInput)
				rightForm.GetFormItem(0).(*tview.InputField).SetText(sim.InicialInput)
				rightForm.GetFormItem(1).(*tview.InputField).SetText(sim.AporteInput)
				rightForm.GetFormItem(2).(*tview.Checkbox).SetChecked(sim.AporteNoInicio)
				aporteNoInicio = sim.AporteNoInicio
				pages.RemovePage("load")
				app.SetFocus(layout)
			}).
			AddButton("Cancelar", func() {
				pages.RemovePage("load")
				app.SetFocus(layout)
			})
		form.SetBorder(true).SetTitle("Carregar simulação")
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(form, 50, 1, true).
			AddItem(nil, 0, 1, false)
		pages.AddPage("load", modal, true, true)
		app.SetFocus(dropdown)
	}

	/* ---------- LAYOUT PRINCIPAL ---------- */
	forms := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(leftForm, 0, 1, true).
		AddItem(rightForm, 0, 1, false)
	layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(forms, 0, 1, true).
		AddItem(result, 0, 2, false)
	pages.AddPage("main", layout, true, true)

	/* ---------- ATALHOS ---------- */
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF2:
			app.SetFocus(leftForm)
			return nil
		case tcell.KeyF3:
			app.SetFocus(rightForm)
			return nil
		case tcell.KeyF5:
			calculate()
			return nil
		case tcell.KeyF9:
			loadSimulation()
			return nil
		case tcell.KeyF10:
			saveSimulation()
			return nil
		case tcell.KeyEsc:
			app.Stop()
			return nil

		case tcell.KeyUp:
			if scrollRow > 0 {
				scrollRow--
				result.ScrollTo(scrollRow, scrollCol)
			}
			return nil
		case tcell.KeyDown:
			if scrollRow < totalLines-1 {
				scrollRow++
				result.ScrollTo(scrollRow, scrollCol)
			}
			return nil
		case tcell.KeyPgUp:
			scrollRow -= 10
			if scrollRow < 0 {
				scrollRow = 0
			}
			result.ScrollTo(scrollRow, scrollCol)
			return nil
		case tcell.KeyPgDn:
			scrollRow += 10
			if scrollRow > totalLines-1 {
				scrollRow = totalLines - 1
			}
			result.ScrollTo(scrollRow, scrollCol)
			return nil
		case tcell.KeyHome:
			scrollRow = 0
			scrollCol = 0
			result.ScrollToBeginning()
			return nil
		case tcell.KeyEnd:
			scrollRow = totalLines - 1
			scrollCol = maxCols
			result.ScrollToEnd()
			return nil
		case tcell.KeyLeft:
			if scrollCol > 0 {
				scrollCol--
				result.ScrollTo(scrollRow, scrollCol)
			}
			return nil
		case tcell.KeyRight:
			if scrollCol < maxCols {
				scrollCol++
				result.ScrollTo(scrollRow, scrollCol)
			}
			return nil
		}
		return event
	})

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}
