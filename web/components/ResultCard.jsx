import jsPDF from 'jspdf'
import 'jspdf-autotable'

function ResultCard({ result, onRemove }) {
  const { id, data, payload } = result

  const brl = (n) => {
    return n.toLocaleString('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    })
  }

  const parseDate = (str) => {
    const [d, m, y] = str.split('/').map(Number)
    return new Date(y, m - 1, d)
  }

  const formatDate = (date) => {
    return date.toLocaleDateString('pt-BR')
  }

  const addMonths = (date, months) => {
    const d = new Date(date)
    d.setMonth(d.getMonth() + months)
    return d
  }

  const dataBase =
    payload.modoCalculo === 'data' && payload.futureDate
      ? parseDate(payload.futureDate)
      : new Date()

  const tableRows = []
  let saldoAtivo = payload.valorInicial

  data.evolucao.forEach(m => {
    const ir = m.rendimento * (data.aliquotaIrpf / 100)
    const liq = m.rendimento - ir

    if (payload.aporteNoInicio) saldoAtivo += m.aporte
    saldoAtivo += liq
    if (!payload.aporteNoInicio) saldoAtivo += m.aporte

    const dataAporte = addMonths(dataBase, m.mes - 1)

    tableRows.push({
      mes: m.mes,
      data: formatDate(dataAporte),
      saldoInicial: m.saldoInicial,
      aporte: m.aporte,
      rendimento: m.rendimento,
      ir,
      liq,
      saldoAtivo
    })
  })

  const exportPDF = () => {
    const doc = new jsPDF()
    doc.text('Simulação de Investimento CDI', 14, 15)

    const tableData = tableRows.map(row => [
      row.mes,
      row.data,
      brl(row.saldoInicial),
      brl(row.aporte),
      brl(row.rendimento),
      brl(row.ir),
      brl(row.liq),
      brl(row.saldoAtivo)
    ])

    doc.autoTable({
      head: [['Mês', 'Data', 'Saldo Inicial', 'Aporte', 'Rend. Bruto', 'IRPF', 'Rend. Líquido', 'Saldo Líquido']],
      body: tableData,
      startY: 25,
      theme: 'grid'
    })

    doc.save(`simulacao-${id}.pdf`)
  }

  const exportCSV = () => {
    let csv = 'Mes;Data;Saldo Inicial;Aporte;Rendimento Bruto;IRPF;Rendimento Liquido;Saldo Final\n'

    tableRows.forEach(row => {
      const rowData = [
        row.mes,
        row.data,
        row.saldoInicial.toFixed(2).replace('.', ','),
        row.aporte.toFixed(2).replace('.', ','),
        row.rendimento.toFixed(2).replace('.', ','),
        row.ir.toFixed(2).replace('.', ','),
        row.liq.toFixed(2).replace('.', ','),
        row.saldoAtivo.toFixed(2).replace('.', ',')
      ]
      csv += rowData.join(';') + '\n'
    })

    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.setAttribute('download', `investimento-${id}.csv`)
    link.click()
  }

  const simNumber = id.replace('result-', '')

  return (
    <div className="result" id={id}>
      <div className="result-header">
        <div className="result-title">
          <span>Simulação #{simNumber}</span>
          <span className="result-params">
            {payload.percentualCdi}% CDI | {payload.prazoMeses || '?'} meses | Inicial: {brl(payload.valorInicial)} | Recorrência: {brl(payload.aporteMensal)}
          </span>
        </div>
        <button className="btn-remove" onClick={() => onRemove(id)}>✕</button>
      </div>

      <div className="summary">
        <div className="metric-card">
          <div style={{ color: 'var(--muted)' }}>Investido</div>
          <div className="metric-value">{brl(data.totalInvestido)}</div>
        </div>
        <div className="metric-card">
          <div style={{ color: 'var(--muted)' }}>Rend. Bruto</div>
          <div className="metric-value">{brl(data.totalJurosBruto)}</div>
        </div>
        <div className="metric-card">
          <div style={{ color: 'var(--muted)' }}>Total IRPF</div>
          <div className="metric-value" style={{ color: 'var(--danger)' }}>
            {brl(data.totalIrpf)}
          </div>
        </div>
        <div className="metric-card" style={{ borderColor: 'var(--accent)' }}>
          <div style={{ color: 'var(--accent)' }}>Líquido Final</div>
          <div className="metric-value">{brl(data.resultadoLiquido)}</div>
        </div>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th className="text-center">Mês</th>
              <th>Data</th>
              <th>Saldo Inicial</th>
              <th>Aporte</th>
              <th>Rend. Bruto</th>
              <th>IRPF</th>
              <th>Rend. Líquido</th>
              <th>Saldo Líquido</th>
            </tr>
          </thead>
          <tbody>
            {tableRows.map(row => (
              <tr key={row.mes}>
                <td className="text-center">{row.mes}</td>
                <td>{row.data}</td>
                <td>{brl(row.saldoInicial)}</td>
                <td>{brl(row.aporte)}</td>
                <td>{brl(row.rendimento)}</td>
                <td>{brl(row.ir)}</td>
                <td>{brl(row.liq)}</td>
                <td style={{ fontWeight: 600 }}>{brl(row.saldoAtivo)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="btn-group">
        <button onClick={exportPDF} className="btn-secondary">Exportar PDF</button>
        <button onClick={exportCSV} className="btn-secondary">Exportar CSV</button>
      </div>
    </div>
  )
}

export default ResultCard
