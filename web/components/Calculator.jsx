import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import FormulaHelp from '@/components/FormulaHelp'

const inputClass = 'h-11 px-3.5 text-base'

function Calculator({ onSubmit }) {
  const [formData, setFormData] = useState({
    cdi: '14.90',
    percentual: '100',
    futureDate: '',
    prazo: '',
    inicial: '',
    aporte: '0',
    aporteNoInicio: true
  })
  const [error, setError] = useState('')
  const [lastEdited, setLastEdited] = useState('prazo')

  const parseDate = (str) => {
    const [d, m, y] = str.split('/').map(Number)
    return new Date(y, m - 1, d)
  }

  const formatDate = (date) => {
    return date.toLocaleDateString('pt-BR')
  }

  const formatCurrencyDisplay = (value) => {
    const num = parseFloat(value)
    if (Number.isNaN(num)) return ''
    return num.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  }

  const handleChange = (e) => {
    const { id, value } = e.target
    setFormData(prev => ({ ...prev, [id]: value }))
  }

  const handleCurrencyChange = (e) => {
    const { id, value } = e.target
    const digits = value.replace(/\D/g, '')
    const num = digits ? (parseInt(digits, 10) / 100).toFixed(2) : ''
    setFormData(prev => ({ ...prev, [id]: num }))
  }

  const handleAporteNoInicioChange = (checked) => {
    setFormData(prev => ({ ...prev, aporteNoInicio: checked === true }))
  }

  const handleFutureDateChange = (e) => {
    const digits = e.target.value.replace(/\D/g, '').slice(0, 8)
    let formatted = digits
    if (digits.length > 4) formatted = `${digits.slice(0, 2)}/${digits.slice(2, 4)}/${digits.slice(4)}`
    else if (digits.length > 2) formatted = `${digits.slice(0, 2)}/${digits.slice(2)}`

    setFormData(prev => ({ ...prev, futureDate: formatted }))
    setLastEdited('data')
  }

  const handleFutureDateBlur = () => {
    if (!formData.futureDate || formData.futureDate.length < 10) return
    const target = parseDate(formData.futureDate)
    const now = new Date()
    if (isNaN(target)) return

    const diff =
      (target.getFullYear() - now.getFullYear()) * 12 +
      (target.getMonth() - now.getMonth())

    setFormData(prev => ({ ...prev, prazo: diff > 0 ? diff.toString() : '0' }))
  }

  const handlePrazoChange = (e) => {
    const value = e.target.value
    setFormData(prev => ({ ...prev, prazo: value }))
    setLastEdited('prazo')
  }

  const handlePrazoBlur = () => {
    if (!formData.prazo) return
    const target = new Date()
    target.setMonth(target.getMonth() + parseInt(formData.prazo, 10))
    setFormData(prev => ({ ...prev, futureDate: formatDate(target) }))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')

    const payload = {
      cdiAnual: parseFloat(formData.cdi),
      percentualCdi: parseFloat(formData.percentual),
      valorInicial: parseFloat(formData.inicial),
      aporteMensal: parseFloat(formData.aporte || '0'),
      aporteNoInicio: formData.aporteNoInicio
    }

    if (lastEdited === 'data' && formData.futureDate) {
      payload.futureDate = formData.futureDate
      payload.modoCalculo = 'data'
    } else {
      payload.prazoMeses = parseInt(formData.prazo, 10) || 0
      payload.modoCalculo = 'prazo'
    }

    try {
      const res = await fetch('/simulate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      })

      const data = await res.json()
      if (!res.ok) throw new Error(data.erro || 'Erro na simulação')

      onSubmit(data, payload)
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <Card className="border-border shadow-none [--card-spacing:2rem]">
      <CardContent>
        <div className="mb-6 flex items-center justify-between gap-4">
          <h2 className="text-sm font-semibold text-foreground">Simulação de investimento</h2>
          <FormulaHelp />
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-8">
          <div className="grid gap-8 md:grid-cols-2">
            <section className="flex flex-col gap-5">
              <h3 className="text-xs font-bold tracking-wide text-muted-foreground uppercase">Parâmetros</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="cdi">Taxa Anual (%)</Label>
                  <div className="relative">
                    <Input
                      id="cdi"
                      type="number"
                      step="0.01"
                      className={`${inputClass} pr-8`}
                      value={formData.cdi}
                      onChange={handleChange}
                      required
                    />
                    <span className="pointer-events-none absolute inset-y-0 right-3 flex items-center text-sm text-muted-foreground">%</span>
                  </div>
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="percentual">Percentual do CDI (%)</Label>
                  <div className="relative">
                    <Input
                      id="percentual"
                      type="number"
                      step="1"
                      className={`${inputClass} pr-8`}
                      value={formData.percentual}
                      onChange={handleChange}
                      required
                    />
                    <span className="pointer-events-none absolute inset-y-0 right-3 flex items-center text-sm text-muted-foreground">%</span>
                  </div>
                </div>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="futureDate">Vencimento (DD/MM/AAAA)</Label>
                  <Input
                    id="futureDate"
                    type="text"
                    inputMode="numeric"
                    placeholder="10/10/2026"
                    className={inputClass}
                    value={formData.futureDate}
                    onChange={handleFutureDateChange}
                    onBlur={handleFutureDateBlur}
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="prazo">Prazo (Meses)</Label>
                  <Input
                    id="prazo"
                    type="number"
                    min="0"
                    max="500"
                    className={inputClass}
                    value={formData.prazo}
                    onChange={handlePrazoChange}
                    onBlur={handlePrazoBlur}
                  />
                </div>
              </div>
            </section>

            <Separator className="md:hidden" />

            <section className="flex flex-col gap-5">
              <h3 className="text-xs font-bold tracking-wide text-muted-foreground uppercase">Investimento</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="inicial">Inicial (R$)</Label>
                  <div className="relative">
                    <span className="pointer-events-none absolute inset-y-0 left-3.5 flex items-center text-sm text-muted-foreground">R$</span>
                    <Input
                      id="inicial"
                      type="text"
                      inputMode="decimal"
                      placeholder="0,00"
                      className={`${inputClass} pl-10`}
                      value={formatCurrencyDisplay(formData.inicial)}
                      onChange={handleCurrencyChange}
                      required
                    />
                  </div>
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="aporte">Aporte Mensal (R$)</Label>
                  <div className="relative">
                    <span className="pointer-events-none absolute inset-y-0 left-3.5 flex items-center text-sm text-muted-foreground">R$</span>
                    <Input
                      id="aporte"
                      type="text"
                      inputMode="decimal"
                      placeholder="0,00"
                      className={`${inputClass} pl-10`}
                      value={formatCurrencyDisplay(formData.aporte)}
                      onChange={handleCurrencyChange}
                    />
                  </div>
                </div>
              </div>
            </section>
          </div>

          <Separator />

          <div className="flex flex-col items-center gap-4 sm:flex-row sm:justify-between">
            <Label
              htmlFor="aporteNoInicio"
              className="w-fit cursor-pointer rounded-lg border border-border bg-muted/40 px-4 py-2.5 font-normal"
            >
              <Checkbox
                id="aporteNoInicio"
                checked={formData.aporteNoInicio}
                onCheckedChange={handleAporteNoInicioChange}
              />
              Aporte no início do mês
            </Label>
            <Button type="submit" size="lg" className="h-12 min-w-[200px] cursor-pointer text-sm">
              Calcular
            </Button>
          </div>
          {error && <div id="error">{error}</div>}
        </form>
      </CardContent>
    </Card>
  )
}

export default Calculator
