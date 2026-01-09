let lastResultData = null;
const inputFuture = document.getElementById('future');
const inputPrazo = document.getElementById('prazo');
let lastEdited = 'prazo';

/* =========================
   THEME TOGGLE
========================= */
const themeToggle = document.getElementById('theme-toggle');
if (themeToggle) {
  themeToggle.addEventListener('click', () => {
    document.body.classList.toggle('dark-mode');
    themeToggle.textContent = document.body.classList.contains('dark-mode') ? '☀️' : '🌙';
  });
}

/* =========================
   DATE HELPERS
========================= */
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

/* =========================
   INPUT SYNC (DATA x PRAZO)
========================= */
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

/* =========================
   FORMATTERS
========================= */
function brl(n) {
  return n.toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL'
  });
}

/* =========================
   FORM SUBMIT
========================= */
const formEl = document.getElementById('form');
if (formEl) {
  formEl.addEventListener('submit', async (e) => {
    e.preventDefault();

    const resultDiv = document.getElementById('result');
    const tbody = document.querySelector('#tabela tbody');
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

      lastResultData = data;

      document.getElementById('summary').innerHTML = `
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

      tbody.innerHTML = '';
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

        tbody.innerHTML += `
          <tr>
            <td class="text-center">${m.mes}</td>
            <td>${formatDate(dataAporte)}</td>
            <td>${brl(m.saldoInicial)}</td>
            <td>${brl(m.aporte)}</td>
            <td>${brl(m.rendimento)}</td>
            <td>${brl(ir)}</td>
            <td>${brl(liq)}</td>
            <td style="font-weight:600">${brl(saldoAtivo)}</td>
          </tr>
        `;
      });

      resultDiv.hidden = false;

    } catch (err) {
      if (errorDiv) errorDiv.textContent = err.message;
    }
  });
}

/* =========================
   EXPORTS
========================= */
function exportPDF() {
  const { jsPDF } = window.jspdf;
  const doc = new jsPDF();
  doc.text("Simulação de Investimento CDI", 14, 15);
  doc.autoTable({
    html: '#tabela',
    startY: 25,
    theme: 'grid'
  });
  doc.save('simulacao.pdf');
}

function exportCSV() {
  let csv =
      "Mes;Data;Saldo Inicial;Aporte;Rendimento Bruto;IRPF;Rendimento Liquido;Saldo Final\n";

  const rows = document.querySelectorAll("#tabela tbody tr");
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
  link.setAttribute("download", "investimento.csv");
  link.click();
}

window.exportPDF = exportPDF;
window.exportCSV = exportCSV;
