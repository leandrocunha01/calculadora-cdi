function Comparison({ results }) {
  if (results.length < 2) {
    return null
  }

  const brl = (n) => {
    return n.toLocaleString('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    })
  }

  const sorted = [...results].sort((a, b) =>
    b.data.resultadoLiquido - a.data.resultadoLiquido
  )

  const best = sorted[0]
  const worst = sorted[sorted.length - 1]
  const diff = best.data.resultadoLiquido - worst.data.resultadoLiquido
  const diffPercent = ((diff / worst.data.resultadoLiquido) * 100).toFixed(2)

  return (
    <div className="comparison">
      <div className="comparison-header">📊 Comparativo</div>
      <div className="comparison-grid">
        <div className="comparison-item best">
          <div className="comparison-label">Melhor Resultado</div>
          <div className="comparison-value">{brl(best.data.resultadoLiquido)}</div>
          <div className="comparison-detail">
            {best.payload.percentualCdi}% CDI | {best.payload.prazoMeses} meses
          </div>
        </div>
        <div className="comparison-item diff">
          <div className="comparison-label">Diferença</div>
          <div className="comparison-value">{brl(diff)}</div>
          <div className="comparison-detail">+{diffPercent}% a mais</div>
        </div>
        <div className="comparison-item worst">
          <div className="comparison-label">Pior Resultado</div>
          <div className="comparison-value">{brl(worst.data.resultadoLiquido)}</div>
          <div className="comparison-detail">
            {worst.payload.percentualCdi}% CDI | {worst.payload.prazoMeses} meses
          </div>
        </div>
      </div>
    </div>
  )
}

export default Comparison
