package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type simulateRequest struct {
	FutureDate     string  `json:"futureDate"` // DD/MM/AAAA
	PrazoMeses     int     `json:"prazoMeses"`
	CDIAnual       float64 `json:"cdiAnual"`      // percentual, ex: 13.65
	PercentualCDI  float64 `json:"percentualCdi"` // percentual, ex: 100
	ValorInicial   float64 `json:"valorInicial"`
	AporteMensal   float64 `json:"aporteMensal"`
	AporteNoInicio bool    `json:"aporteNoInicio"`
}

type simulateResponse struct {
	Meses            int     `json:"meses"`
	AliquotaIRPF     float64 `json:"aliquotaIrpf"`
	TotalInvestido   float64 `json:"totalInvestido"`
	TotalJurosBruto  float64 `json:"totalJurosBruto"`
	TotalIRPF        float64 `json:"totalIrpf"`
	ResultadoLiquido float64 `json:"resultadoLiquido"`
	SaldoFinalBruto  float64 `json:"saldoFinalBruto"`
	Evolucao         []Mes   `json:"evolucao"`
	Erro             string  `json:"erro,omitempty"`
}

func parseBrazilDate(s string) (time.Time, error) {
	layout := "02/01/2006"
	return time.Parse(layout, strings.TrimSpace(s))
}

func calcularPrazoMeses(futureText string, prazoMeses int) (int, error) {
	now := time.Now()
	if strings.TrimSpace(futureText) != "" {
		future, err := parseBrazilDate(futureText)
		if err != nil {
			return 0, err
		}
		meses := (future.Year()-now.Year())*12 + int(future.Month()-now.Month())
		if future.Day() < now.Day() {
			meses--
		}
		if meses < 0 {
			meses = 0
		}
		return meses, nil
	}
	if prazoMeses > 0 {
		return prazoMeses, nil
	}
	return 0, ErrInvalidPrazo
}

var ErrInvalidPrazo = &httpError{Status: http.StatusBadRequest, Msg: "Forneça 'futureDate' (DD/MM/AAAA) ou 'prazoMeses' > 0"}

type httpError struct {
	Status int
	Msg    string
}

func (e *httpError) Error() string { return e.Msg }

func init() {
	// attempt to load environment variables from a .env file in the project root
	// ignores error so that absence of file is not fatal.
	_ = godotenv.Load()
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// saveSimulation sends the request/response pair to Supabase table `simulations`.
// It uses the REST API; the environment variables SUPABASE_URL and SUPABASE_KEY
// must be configured. Optionally SUPABASE_USER_ID may provide the `user_id`
// foreign key value, otherwise the column will be omitted (and default to null).
func saveSimulation(req simulateRequest, resp simulateResponse) error {
	supaURL := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")
	if supaURL == "" || key == "" {
		// nothing to do if not configured
		return nil
	}

	payload := map[string]interface{}{
		"request":  req,
		"response": resp,
	}
	if uid := os.Getenv("SUPABASE_USER_ID"); uid != "" {
		payload["user_id"] = uid
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := strings.TrimRight(supaURL, "/") + "/rest/v1/simulations"
	r, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("apikey", key)
	r.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("supabase error: %s", string(body))
	}
	return nil
}

func firstDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func lastDayOfMonth(t time.Time) time.Time {
	// day 0 of next month is the last day of current month
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location())
}

func addMonths(t time.Time, months int) time.Time {
	// Keep at first day to avoid DST issues, then restore chosen day rule elsewhere
	base := firstDayOfMonth(t)
	return time.Date(base.Year(), base.Month()+time.Month(months), 1, 0, 0, 0, 0, base.Location())
}

func formatDateBR(t time.Time) string {
	return t.Format("02/01/2006")
}

func simulateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"erro": "Método não permitido"})
		return
	}
	var req simulateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "JSON inválido"})
		return
	}

	meses, err := calcularPrazoMeses(req.FutureDate, req.PrazoMeses)
	if err != nil {
		if herr, ok := err.(*httpError); ok {
			writeJSON(w, herr.Status, map[string]string{"erro": herr.Msg})
		} else {
			writeJSON(w, http.StatusBadRequest, map[string]string{"erro": err.Error()})
		}
		return
	}

	taxaAnual := (req.CDIAnual / 100.0) * (req.PercentualCDI / 100.0)

	evolucao, totalInvestido, totalJurosBruto, saldoFinalBruto := calcularEvolucao(
		req.ValorInicial, req.AporteMensal, taxaAnual, meses, req.AporteNoInicio,
	)

	// IR por mês e totais
	totalIRPF := 0.0
	for _, e := range evolucao {
		totalIRPF += calcularIR(e.Rendimento, meses)
	}
	totalJurosLiquido := 0.0
	for _, e := range evolucao {
		irMes := calcularIR(e.Rendimento, meses)
		totalJurosLiquido += e.Rendimento - irMes
	}
	resultadoLiquido := totalInvestido + totalJurosLiquido

	// Preencher as datas de aporte por mês
	// Referência: mês corrente como mês 1
	now := time.Now()
	for i := range evolucao {
		// i = 0 para mês 1 -> adicionar i meses
		refMonth := addMonths(now, i)
		var d time.Time
		if req.AporteNoInicio {
			d = firstDayOfMonth(refMonth)
		} else {
			d = lastDayOfMonth(refMonth)
		}
		evolucao[i].DataAporte = formatDateBR(d)
	}

	resp := simulateResponse{
		Meses:            meses,
		AliquotaIRPF:     aliquotaIRPFRendaFixa(meses),
		TotalInvestido:   totalInvestido,
		TotalJurosBruto:  totalJurosBruto,
		TotalIRPF:        totalIRPF,
		ResultadoLiquido: resultadoLiquido,
		SaldoFinalBruto:  saldoFinalBruto,
		Evolucao:         evolucao,
	}

	// save asynchronously so that the API call is not delayed
	go func() {
		if err := saveSimulation(req, resp); err != nil {
			log.Printf("erro ao salvar simulação no supabase: %v", err)
		}
	}()

	writeJSON(w, http.StatusOK, resp)
}

func main() {
	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/simulate", simulateHandler)

	// Static: serve ./web/dist for production (vite build output) or ./web for development
	distDir := filepath.Join("web", "dist")
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try dist first (production build)
		distPath := filepath.Join(distDir, path)
		if info, err := os.Stat(distPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, distPath)
			return
		}

		// Try web directory (development)
		webPath := filepath.Join("web", path)
		if info, err := os.Stat(webPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, webPath)
			return
		}

		// Try assets in dist
		if strings.HasPrefix(path, "/assets/") {
			assetsPath := filepath.Join(distDir, path)
			if info, err := os.Stat(assetsPath); err == nil && !info.IsDir() {
				http.ServeFile(w, r, assetsPath)
				return
			}
		}

		// Default to index.html for SPA routing
		if _, err := os.Stat(filepath.Join(distDir, "index.html")); err == nil {
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
		} else {
			http.ServeFile(w, r, filepath.Join("web", "index.html"))
		}
	}))

	addr := ":8080"
	log.Printf("Servidor iniciado em http://localhost%v\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
