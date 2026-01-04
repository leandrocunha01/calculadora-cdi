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
}

/* =========================
   CÁLCULO
========================= */

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
	pages := tview.NewPages()

	var lastSimulation *Simulation
	var layout *tview.Flex // <--- adicione esta linha

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
	result.SetTitle("+ Resultado F5 calcular | F9 carregar | F10 salvar | Esc sair +")

	scrollRow := 0
	scrollCol := 0
	totalLines := 0
	maxCols := 0

	/* ---------- FUNÇÃO CALCULAR ---------- */

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

		lastSimulation = &Simulation{
			DataExecucao:   time.Now(),
			Inicial:        inicial,
			Aporte:         aporte,
			TaxaAnual:      taxaAnual,
			Meses:          meses,
			AporteNoInicio: aporteNoInicio,
			Evolucao:       evolucao,
			TotalInvestido: investido,
			TotalJuros:     juros,
			SaldoFinal:     final,
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Período         : %d meses\n", meses))
		sb.WriteString(fmt.Sprintf("Total investido : R$ %.2f\n", investido))
		sb.WriteString(fmt.Sprintf("Total juros     : R$ %.2f\n", juros))
		sb.WriteString(fmt.Sprintf("Resultado       : R$ %.2f\n", final-investido))
		sb.WriteString(fmt.Sprintf("Valor final     : [green]R$ %.2f[-]\n\n", final))

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

		text := sb.String()
		result.SetText(text)
		result.ScrollToBeginning()

		scrollRow = 0
		scrollCol = 0
		totalLines = strings.Count(text, "\n")
		maxCols = 0
		for _, line := range strings.Split(text, "\n") {
			if len(line) > maxCols {
				maxCols = len(line)
			}
		}
	}

	/* ---------- FUNÇÃO SALVAR (F10) ---------- */
	saveSimulation := func() {
		if lastSimulation == nil {
			return
		}

		input := tview.NewInputField().
			SetLabel("Nome da simulação: ")

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

	/* ---------- FUNÇÃO CARREGAR (F9) ---------- */
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
			SetOptions(options, func(text string, index int) {
				selectedFile = text
			})

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

				layoutDate := "02/01/2006"
				future := time.Now().AddDate(0, sim.Meses, 0) // placeholder
				leftForm.GetFormItem(0).(*tview.InputField).SetText(future.Format(layoutDate))
				leftForm.GetFormItem(1).(*tview.InputField).SetText(fmt.Sprintf("%.2f", sim.TaxaAnual*100))
				leftForm.GetFormItem(2).(*tview.InputField).SetText("100") // percentual aproximado

				rightForm.GetFormItem(0).(*tview.InputField).SetText(fmt.Sprintf("%.2f", sim.Inicial))
				rightForm.GetFormItem(1).(*tview.InputField).SetText(fmt.Sprintf("%.2f", sim.Aporte))
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

	/* ---------- ATALHOS GLOBAIS ---------- */
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

		// scroll vertical
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

		// scroll horizontal
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
