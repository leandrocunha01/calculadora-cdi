import { Sigma } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog'

function FormulaHelp() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <button
          type="button"
          className="inline-flex h-7 cursor-pointer items-center gap-1.5 rounded-full border border-border bg-muted/40 px-3 text-xs font-medium text-muted-foreground hover:bg-muted"
        >
          <Sigma className="size-3.5" />
          Como calculamos?
        </button>
      </DialogTrigger>
      <DialogContent className="flex max-h-[85vh] w-[calc(100%-2rem)] max-w-2xl flex-col gap-0 p-0 sm:max-w-2xl">
        <DialogHeader className="border-b border-border px-6 py-4">
          <DialogTitle className="text-lg">Como calculamos</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-6 overflow-y-auto px-6 py-5 text-sm leading-relaxed">
          <div>
            <p className="mb-1.5 text-base font-semibold text-foreground">Variáveis</p>
            <pre className="mt-1 overflow-x-auto rounded-md bg-muted p-3 text-sm text-foreground">
{`n  = prazo total, em meses (campo "Prazo")
m  = índice do mês corrente, de 1 até n
S  = saldo acumulado
A  = valor do aporte mensal
R  = rendimento gerado no mês
α  = alíquota de IR sobre o rendimento`}
            </pre>
          </div>

          <div>
            <p className="mb-1.5 text-base font-semibold text-foreground">1. Taxa mensal equivalente</p>
            <p className="text-muted-foreground">A taxa anual é composta pelo % do CDI escolhido e convertida para uma taxa mensal equivalente:</p>
            <pre className="mt-2 overflow-x-auto rounded-md bg-muted p-3 text-sm text-foreground">
{`i_a = CDI × (%CDI ÷ 100)
i_m = (1 + i_a)^(1/12) − 1`}
            </pre>
          </div>

          <div>
            <p className="mb-1.5 text-base font-semibold text-foreground">2. Evolução do saldo (mês a mês)</p>
            <p className="text-muted-foreground">Para cada mês m, o aporte A entra antes ou depois do rendimento:</p>
            <pre className="mt-2 overflow-x-auto rounded-md bg-muted p-3 text-sm text-foreground">
{`aporte no início:
  Sₘ = (Sₘ₋₁ + A) × (1 + i_m)

aporte no fim:
  Sₘ = Sₘ₋₁ × (1 + i_m) + A`}
            </pre>
          </div>

          <div>
            <p className="mb-1.5 text-base font-semibold text-foreground">3. Rendimento bruto acumulado</p>
            <p className="text-muted-foreground">Soma do rendimento gerado em cada um dos n meses:</p>
            <pre className="mt-2 overflow-x-auto rounded-md bg-muted p-3 text-sm text-foreground">
{`      n
J = Σ  Rₘ
     m=1`}
            </pre>
          </div>

          <div>
            <p className="mb-1.5 text-base font-semibold text-foreground">4. IR regressivo (renda fixa)</p>
            <p className="text-muted-foreground">A alíquota α depende do prazo total n, em meses:</p>
            <pre className="mt-2 overflow-x-auto rounded-md bg-muted p-3 text-sm text-foreground">
{`α(n) = 22,5%  se n ≤ 6
α(n) = 20,0%  se n ≤ 12
α(n) = 17,5%  se n ≤ 24
α(n) = 15,0%  se n > 24`}
            </pre>
          </div>

          <div>
            <p className="mb-1.5 text-base font-semibold text-foreground">5. Resultado líquido final</p>
            <pre className="mt-2 overflow-x-auto rounded-md bg-muted p-3 text-sm text-foreground">
{`               n
L = investido + Σ  Rₘ × (1 − α(n))
              m=1`}
            </pre>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export default FormulaHelp
