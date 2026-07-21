import { useState } from 'react'

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

  const handleChange = (e) => {
    const { id, value, type, checked } = e.target
    setFormData(prev => ({
      ...prev,
      [id]: type === 'checkbox' ? checked : value
    }))
  }

  const handleFutureDateChange = (e) => {
    const value = e.target.value
    setFormData(prev => ({ ...prev, futureDate: value }))
    setLastEdited('data')
  }

  const handleFutureDateBlur = () => {
    if (!formData.futureDate) return
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
      aporteMensal: parseFloat(formData.aporte),
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
    <form onSubmit={handleSubmit}>
      <div className="row">
        <fieldset>
          <legend>Parâmetros</legend>
          <label>Taxa Anual (%)</label>
          <input
            id="cdi"
            type="number"
            step="0.01"
            value={formData.cdi}
            onChange={handleChange}
            required
          />
          <label>Percentual do CDI (%)</label>
          <input
            id="percentual"
            type="number"
            step="1"
            value={formData.percentual}
            onChange={handleChange}
            required
          />
          <label>Vencimento (DD/MM/AAAA)</label>
          <input
            id="futureDate"
            type="text"
            placeholder="Ex: 10/10/2026"
            value={formData.futureDate}
            onChange={handleFutureDateChange}
            onBlur={handleFutureDateBlur}
          />
          <label>Prazo (Meses)</label>
          <input
            id="prazo"
            type="number"
            min="0"
            max="500"
            value={formData.prazo}
            onChange={handlePrazoChange}
            onBlur={handlePrazoBlur}
          />
        </fieldset>
        <fieldset>
          <legend>Investimento</legend>
          <label>Inicial (R$)</label>
          <input
            id="inicial"
            type="number"
            step="0.01"
            value={formData.inicial}
            onChange={handleChange}
            required
          />
          <label>Aporte Mensal (R$)</label>
          <input
            id="aporte"
            type="number"
            step="0.01"
            value={formData.aporte}
            onChange={handleChange}
          />
          <label style={{ display: 'flex', alignItems: 'center', gap: '8px', marginTop: '10px' }}>
            <input
              id="aporteNoInicio"
              type="checkbox"
              checked={formData.aporteNoInicio}
              onChange={handleChange}
              style={{ width: 'auto', margin: 0 }}
            />
            Aporte no início
          </label>
        </fieldset>
      </div>
      <div style={{ textAlign: 'center' }}>
        <button type="submit">Calcular</button>
        {error && <div id="error">{error}</div>}
      </div>
    </form>
  )
}

export default Calculator
