let savedResults = [];
let resultCounter = 0;
const inputFuture = document.getElementById('future');
const inputPrazo = document.getElementById('prazo');
let lastEdited = 'prazo';

const themeToggle = document.getElementById('theme-toggle');
if (themeToggle) {
  themeToggle.addEventListener('click', () => {
    document.body.classList.toggle('dark-mode');
    themeToggle.textContent = document.body.classList.contains('dark-mode') ? '☀️' : '🌙';
  });
}

function parseDate(str) {
  const [d, m, y] = str.split('/').map(Number);
  return new Date(y, m - 1, d);
}

function formatDate(date) {
  return date.toLocaleDateString('pt-BR');
}

function addMonths(date, months) {
  const d = new Date(date);
  d.setMonth(d.getMonth() + months);
  return d;
}

if (inputFuture) {
  inputFuture.addEventListener('input', () => { lastEdited = 'data'; });
  inputFuture.addEventListener('blur', () => {
    if (!inputFuture.value) return;
    const target = parseDate(inputFuture.value);
    const now = new Date();
    if (isNaN(target)) return;

    const diff =
        (target.getFullYear() - now.getFullYear()) * 12 +
        (target.getMonth() - now.getMonth());

    inputPrazo.value = diff > 0 ? diff : 0;
  });
}

if (inputPrazo) {
  inputPrazo.addEventListener('input', () => { lastEdited = 'prazo'; });
  inputPrazo.addEventListener('blur', () => {
    if (!inputPrazo.value) return;
    const target = new Date();
    target.setMonth(target.getMonth() + parseInt(inputPrazo.value, 10));
    inputFuture.value = formatDate(target);
  });
}

function brl(n) {
  return n.toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL'
  });
}

const formEl = document.getElementById('form');
if (formEl) {
  formEl.addEventListener('submit', async (e) => {
    e.preventDefault();

    const errorDiv = document.getElementById('error');
    if (errorDiv) errorDiv.textContent = '';

    const payload = {
      cdiAnual: parseFloat(document.getElementById('cdi').value),
      percentualCdi: parseFloat(document.getElementById('percentual').value),
      valorInicial: parseFloat(document.getElementById('inicial').value),
      aporteMensal: parseFloat(document.getElementById('aporte').value),
      aporteNoInicio: document.getElementById('inicio').checked
    };

    if (lastEdited === 'data' && inputFuture.value) {
      payload.futureDate = inputFuture.value;
      payload.modoCalculo = 'data';
    } else {
      payload.prazoMeses = parseInt(inputPrazo.value, 10) || 0;
      payload.modoCalculo = 'prazo';
    }

    try {
      const res = await fetch('/simulate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      const data = await res.json();
      if (!res.ok) throw new Error(data.erro || 'Erro na simulação');

      // Limite de 2 simulações
      if (savedResults.length >= 2) {
        const oldest = savedResults[0];
        removeResult(oldest.id);
      }

      resultCounter++;
      const resultId = `result-${resultCounter}`;

      savedResults.push({
        id: resultId,
        data: data,
        payload: payload,
        timestamp: new Date()
      });

      addResultCard(resultId, data, payload);
      updateComparison();

    } catch (err) {
      if (errorDiv) errorDiv.textContent = err.message;
    }
  });
}

function addResultCard(resultId, data, payload) {
  const container = document.getElementById('results-container');

  const resultCard = document.createElement('div');
  resultCard.className = 'result';
  resultCard.id = resultId;

  const resultTitle = document.createElement('div');
  resultTitle.className = 'result-header';
  resultTitle.innerHTML = `
    <div class="result-title">
      <span>Simulação #${resultCounter}</span>
      <span class="result-params">${payload.percentualCdi}% CDI | ${payload.prazoMeses || '?'} meses | Inicial: ${brl(payload.valorInicial)}</span>
    </div>
    <button class="btn-remove" onclick="removeResult('${resultId}')">✕</button>
  `;

  const summary = document.createElement('div');
  summary.className = 'summary';
  summary.innerHTML = `
    <div class="metric-card">
      <div style="color:var(--muted)">Investido</div>
      <div class="metric-value">${brl(data.totalInvestido)}</div>
    </div>
    <div class="metric-card">
      <div style="color:var(--muted)">Rend. Bruto</div>
      <div class="metric-value">${brl(data.totalJurosBruto)}</div>
    </div>
    <div class="metric-card">
      <div style="color:var(--muted)">Total IRPF</div>
      <div class="metric-value" style="color:var(--danger)">
        ${brl(data.totalIrpf)}
      </div>
    </div>
    <div class="metric-card" style="border-color:var(--accent)">
      <div style="color:var(--accent)">Líquido Final</div>
      <div class="metric-value">${brl(data.resultadoLiquido)}</div>
    </div>
  `;

  const tableContainer = document.createElement('div');
  tableContainer.className = 'table-container';

  const table = document.createElement('table');
  table.innerHTML = `
    <thead>
      <tr>
        <th class="text-center">Mês</th>
        <th>Data</th>
        <th>Saldo Inicial</th>
        <th>Aporte</th>
        <th>Rend. Bruto</th>
        <th>IRPF</th>
        <th>Rend. Líquido</th>
        <th>Saldo Líquido</th>
      </tr>
    </thead>
    <tbody></tbody>
  `;

  const tbody = table.querySelector('tbody');
  let saldoAtivo = payload.valorInicial;

  const dataBase =
      payload.modoCalculo === 'data' && payload.futureDate
          ? parseDate(payload.futureDate)
          : new Date();

  data.evolucao.forEach(m => {
    const ir = m.rendimento * (data.aliquotaIrpf / 100);
    const liq = m.rendimento - ir;

    if (payload.aporteNoInicio) saldoAtivo += m.aporte;
    saldoAtivo += liq;
    if (!payload.aporteNoInicio) saldoAtivo += m.aporte;

    const dataAporte = addMonths(dataBase, m.mes - 1);

    const row = document.createElement('tr');
    row.innerHTML = `
      <td class="text-center">${m.mes}</td>
      <td>${formatDate(dataAporte)}</td>
      <td>${brl(m.saldoInicial)}</td>
      <td>${brl(m.aporte)}</td>
      <td>${brl(m.rendimento)}</td>
      <td>${brl(ir)}</td>
      <td>${brl(liq)}</td>
      <td style="font-weight:600">${brl(saldoAtivo)}</td>
    `;
    tbody.appendChild(row);
  });

  tableContainer.appendChild(table);

  const btnGroup = document.createElement('div');
  btnGroup.className = 'btn-group';
  btnGroup.innerHTML = `
    <button onclick="exportPDF('${resultId}')" class="btn-secondary">Exportar PDF</button>
    <button onclick="exportCSV('${resultId}')" class="btn-secondary">Exportar CSV</button>
  `;

  resultCard.appendChild(resultTitle);
  resultCard.appendChild(summary);
  resultCard.appendChild(tableContainer);
  resultCard.appendChild(btnGroup);

  container.insertBefore(resultCard, container.firstChild);
}

function removeResult(resultId) {
  const element = document.getElementById(resultId);
  if (element) {
    element.remove();
    savedResults = savedResults.filter(r => r.id !== resultId);
    updateComparison();
  }
}

function updateComparison() {
  const comparisonDiv = document.getElementById('comparison');

  if (savedResults.length < 2) {
    comparisonDiv.hidden = true;
    return;
  }

  const sorted = [...savedResults].sort((a, b) =>
    b.data.resultadoLiquido - a.data.resultadoLiquido
  );

  const best = sorted[0];
  const worst = sorted[sorted.length - 1];
  const diff = best.data.resultadoLiquido - worst.data.resultadoLiquido;
  const diffPercent = ((diff / worst.data.resultadoLiquido) * 100).toFixed(2);

  comparisonDiv.innerHTML = `
    <div class="comparison-header">📊 Comparativo</div>
    <div class="comparison-grid">
      <div class="comparison-item best">
        <div class="comparison-label">Melhor Resultado</div>
        <div class="comparison-value">${brl(best.data.resultadoLiquido)}</div>
        <div class="comparison-detail">${best.payload.percentualCdi}% CDI | ${best.payload.prazoMeses} meses</div>
      </div>
      <div class="comparison-item diff">
        <div class="comparison-label">Diferença</div>
        <div class="comparison-value">${brl(diff)}</div>
        <div class="comparison-detail">+${diffPercent}% a mais</div>
      </div>
      <div class="comparison-item worst">
        <div class="comparison-label">Pior Resultado</div>
        <div class="comparison-value">${brl(worst.data.resultadoLiquido)}</div>
        <div class="comparison-detail">${worst.payload.percentualCdi}% CDI | ${worst.payload.prazoMeses} meses</div>
      </div>
    </div>
  `;

  comparisonDiv.hidden = false;
}

window.removeResult = removeResult;

/* =========================
   EXPORTS
========================= */
function exportPDF(resultId) {
  const { jsPDF } = window.jspdf;
  const doc = new jsPDF();
  doc.text("Simulação de Investimento CDI", 14, 15);

  const table = document.querySelector(`#${resultId} table`);
  doc.autoTable({
    html: table,
    startY: 25,
    theme: 'grid'
  });
  doc.save(`simulacao-${resultId}.pdf`);
}

function exportCSV(resultId) {
  let csv =
      "Mes;Data;Saldo Inicial;Aporte;Rendimento Bruto;IRPF;Rendimento Liquido;Saldo Final\n";

  const rows = document.querySelectorAll(`#${resultId} tbody tr`);
  rows.forEach(row => {
    const cols = row.querySelectorAll("td");
    const rowData = Array.from(cols).map(c =>
        c.innerText
            .replace(/[R$\s.]/g, '')
            .replace(',', '.')
    );
    csv += rowData.join(";") + "\n";
  });

  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.setAttribute("download", `investimento-${resultId}.csv`);
  link.click();
}

window.exportPDF = exportPDF;
window.exportCSV = exportCSV;
