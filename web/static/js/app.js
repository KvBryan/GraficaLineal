// GraficaLineal - Interactive Scientific Visualization with DeepSeek
let chartInstance = null;
let currentInspection = null;
let currentChartSpec = null;
let selectedTableIndex = null;
let selectedSectionIndex = null;

// Palette definitions
const PALETTES = {
  apa: ['#2b3a4a', '#4a6572', '#34495e', '#7f8c8d', '#95a5a6', '#bdc3c7'],
  navy: ['#1e40af', '#0284c7', '#0d9488', '#f59e0b', '#dc2626', '#7c3aed'],
  nature: ['#065f46', '#059669', '#10b981', '#34d399', '#6ee7b7', '#a7f3d0'],
  vibrant: ['#e11d48', '#d97706', '#2563eb', '#7c3aed', '#059669', '#0891b2']
};

document.addEventListener('DOMContentLoaded', () => {
  initDropzone();
  initSampleButtons();
  initChartToolbar();
  initFormControls();
});

// ----------------------------------------------------
// 1. File Upload & Inspection
// ----------------------------------------------------
function initDropzone() {
  const dropzone = document.getElementById('uploadDropzone');
  const fileInput = document.getElementById('fileInput');

  dropzone.addEventListener('click', () => fileInput.click());

  dropzone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropzone.classList.add('dragover');
  });

  dropzone.addEventListener('dragleave', () => {
    dropzone.classList.remove('dragover');
  });

  dropzone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropzone.classList.remove('dragover');
    if (e.dataTransfer.files.length > 0) {
      uploadFile(e.dataTransfer.files[0]);
    }
  });

  fileInput.addEventListener('change', (e) => {
    if (e.target.files.length > 0) {
      uploadFile(e.target.files[0]);
    }
  });
}

async function uploadFile(file) {
  showLoading('Inspeccionando y analizando documento...');
  const formData = new FormData();
  formData.append('file', file);

  try {
    const res = await fetch('/api/document/inspect', {
      method: 'POST',
      body: formData
    });

    if (!res.ok) {
      const errText = await res.text();
      throw new Error(errText || 'Error analizando documento');
    }

    const data = await res.json();
    currentInspection = data;
    renderInspectionUI(data);
  } catch (err) {
    alert('Error al subir el archivo: ' + err.message);
  } finally {
    hideLoading();
  }
}

// ----------------------------------------------------
// 2. Inspection UI & Interactive Word/Excel Selection
// ----------------------------------------------------
function renderInspectionUI(data) {
  const section = document.getElementById('inspectionSection');
  section.style.display = 'block';
  section.scrollIntoView({ behavior: 'smooth' });

  document.getElementById('docFileName').textContent = data.fileName;
  document.getElementById('docSummary').textContent = data.summary;

  const isWord = data.fileType === 'word';
  const wordPromptBox = document.getElementById('wordPromptBox');
  wordPromptBox.style.display = isWord ? 'block' : 'none';

  const tableList = document.getElementById('tablesList');
  tableList.innerHTML = '';
  selectedTableIndex = null;
  selectedSectionIndex = null;

  if (data.tables && data.tables.length > 0) {
    data.tables.forEach((tbl, idx) => {
      const card = document.createElement('div');
      card.className = `selection-card ${idx === 0 ? 'selected' : ''}`;
      card.dataset.index = idx;
      card.innerHTML = `
        <div class="selection-card-title">
          <span>📊 ${tbl.name}</span>
          <span style="font-size:0.75rem; color:#64748b;">${tbl.totalRows} filas × ${tbl.totalCols} cols</span>
        </div>
        <div class="selection-card-meta">
          <strong>Columnas:</strong> ${tbl.headers.slice(0, 5).join(', ')}${tbl.headers.length > 5 ? '...' : ''}
        </div>
      `;
      card.addEventListener('click', () => {
        document.querySelectorAll('.selection-card').forEach(c => c.classList.remove('selected'));
        card.classList.add('selected');
        selectedTableIndex = idx;
        selectedSectionIndex = null;
      });
      tableList.appendChild(card);
    });
    selectedTableIndex = 0; // Default select first
  } else {
    tableList.innerHTML = '<p style="color:#64748b; font-size:0.85rem;">No se detectaron tablas estructuradas directas.</p>';
  }

  // Word text sections if present
  const sectionsList = document.getElementById('sectionsList');
  sectionsList.innerHTML = '';
  if (data.sections && data.sections.length > 0) {
    document.getElementById('sectionsContainer').style.display = 'block';
    data.sections.forEach((sec, idx) => {
      const card = document.createElement('div');
      card.className = 'selection-card';
      card.dataset.index = idx;
      const snippet = sec.content.length > 120 ? sec.content.slice(0, 120) + '...' : sec.content;
      card.innerHTML = `
        <div class="selection-card-title">
          <span>📄 ${sec.title}</span>
        </div>
        <div class="selection-card-meta">${snippet}</div>
      `;
      card.addEventListener('click', () => {
        document.querySelectorAll('.selection-card').forEach(c => c.classList.remove('selected'));
        card.classList.add('selected');
        selectedSectionIndex = idx;
        selectedTableIndex = null;
      });
      sectionsList.appendChild(card);
    });
  } else {
    document.getElementById('sectionsContainer').style.display = 'none';
  }

  // Auto trigger first generation if Excel
  if (!isWord && data.tables && data.tables.length > 0) {
    generateChartWithAI();
  }
}

// ----------------------------------------------------
// 3. AI Generation with DeepSeek
// ----------------------------------------------------
async function generateChartWithAI(requestedType = '') {
  if (!currentInspection) return;

  showLoading('DeepSeek está analizando los datos y estructurando la gráfica...');

  const userPromptInput = document.getElementById('userPromptInput');
  const userPrompt = userPromptInput ? userPromptInput.value.trim() : '';

  let selectedTable = null;
  let selectedSection = '';

  if (selectedTableIndex !== null && currentInspection.tables && currentInspection.tables[selectedTableIndex]) {
    selectedTable = currentInspection.tables[selectedTableIndex];
  } else if (selectedSectionIndex !== null && currentInspection.sections && currentInspection.sections[selectedSectionIndex]) {
    selectedSection = currentInspection.sections[selectedSectionIndex].content;
  } else if (currentInspection.tables && currentInspection.tables.length > 0) {
    selectedTable = currentInspection.tables[0];
  }

  const payload = {
    documentType: currentInspection.fileType,
    fileName: currentInspection.fileName,
    selectedTable: selectedTable,
    selectedSection: selectedSection,
    userPrompt: userPrompt,
    requestedType: requestedType
  };

  try {
    const res = await fetch('/api/chart/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    if (!res.ok) {
      const err = await res.text();
      throw new Error(err || 'Error generando gráfica');
    }

    const chartSpec = await res.json();
    currentChartSpec = chartSpec;
    renderChartWorkspace(chartSpec);
  } catch (err) {
    alert('Error al generar la gráfica con DeepSeek: ' + err.message);
  } finally {
    hideLoading();
  }
}

// ----------------------------------------------------
// 4. Render Chart & Academic Dashboard
// ----------------------------------------------------
function renderChartWorkspace(spec) {
  const workspace = document.getElementById('chartWorkspace');
  workspace.style.display = 'block';
  workspace.scrollIntoView({ behavior: 'smooth' });

  // Update active type chip
  document.querySelectorAll('.type-chip').forEach(chip => {
    chip.classList.toggle('active', chip.dataset.type === spec.chartType);
  });

  // Render ECharts
  renderECharts(spec);

  // Render Cuadro de Información de la IA
  renderAIInfoBox(spec);

  // Render Table Editor
  renderTableEditor(spec.tableUsed);

  // Render Methodology & Auditing
  renderMethodology(spec);

  // Populate title edit input
  document.getElementById('chartTitleInput').value = spec.title || '';
  document.getElementById('chartSubtitleInput').value = spec.subtitle || '';
}

function renderAIInfoBox(spec) {
  const info = spec.aiInfoBox || {};
  const typeMap = {
    bar: 'Barras / Columnas',
    radar: 'Radial / Araña',
    line: 'Líneas',
    area: 'Áreas',
    gantt: 'Diagrama de Gantt',
    pie: 'Pastel / Sectores',
    doughnut: 'Anillos (Dona)',
    funnel: 'Embudo (Funnel)',
    waterfall: 'Cascada (Waterfall)',
    scatter: 'Dispersión (Scatter)',
    histogram: 'Histograma',
    boxplot: 'Diagrama de Caja (Boxplot)',
    heatmap: 'Mapa de Calor (Heatmap)'
  };

  const badgeElem = document.getElementById('aiChartTypeBadge');
  if (badgeElem) {
    badgeElem.textContent = typeMap[spec.chartType] || (spec.chartType || 'bar').toUpperCase();
  }

  const docSourceElem = document.getElementById('aiDocSource');
  if (docSourceElem) {
    docSourceElem.textContent = info.sourceDocument || (currentInspection ? currentInspection.fileName : 'Documento fuente');
  }

  const docTypeDetail = document.getElementById('aiDocTypeDetail');
  if (docTypeDetail) {
    docTypeDetail.textContent = spec.title || 'Figura de investigación';
  }

  const docLocationElem = document.getElementById('aiDocLocation');
  if (docLocationElem) {
    docLocationElem.textContent = info.sourceLocation || (spec.tableUsed ? spec.tableUsed.sourceLocation : 'Ubicación identificada');
  }

  const docRangeDetail = document.getElementById('aiDocRangeDetail');
  if (docRangeDetail) {
    const rCount = spec.tableUsed && spec.tableUsed.rows ? spec.tableUsed.rows.length : 0;
    const cCount = spec.tableUsed && spec.tableUsed.headers ? spec.tableUsed.headers.length : 0;
    docRangeDetail.textContent = `${rCount} observaciones × ${cCount} dimensiones`;
  }

  const dataSummaryElem = document.getElementById('aiDataSummary');
  if (dataSummaryElem) {
    dataSummaryElem.textContent = info.dataUsedSummary || `Datos tabulados con ${spec.categories ? spec.categories.length : 0} categorías analizadas.`;
  }

  const variablesDetailElem = document.getElementById('aiVariablesDetail');
  if (variablesDetailElem) {
    variablesDetailElem.textContent = info.variablesDetail || (spec.xAxisName ? `Eje X: ${spec.xAxisName} | Eje Y: ${spec.yAxisName}` : 'Variables categóricas y cuantitativas');
  }

  const chartExplanationElem = document.getElementById('aiChartExplanationText');
  if (chartExplanationElem) {
    chartExplanationElem.textContent = info.chartExplanation || (spec.methodology ? spec.methodology.explanation : 'Análisis explicativo generado por DeepSeek.');
  }

  const keyFindingsElem = document.getElementById('aiKeyFindingsText');
  if (keyFindingsElem) {
    keyFindingsElem.textContent = info.keyFindings || 'La visualización permite contrastar el comportamiento cuantitativo y evaluar las diferencias entre los grupos de estudio.';
  }

  const paperNarrativeElem = document.getElementById('aiPaperNarrativeText');
  if (paperNarrativeElem) {
    paperNarrativeElem.textContent = info.paperNarrative || `Tal como se exhibe en la ${spec.title || 'gráfica'}, los datos obtenidos en ${info.sourceLocation || 'la fuente'} reflejan las magnitudes analizadas.`;
  }
}

function renderECharts(spec) {
  const chartDom = document.getElementById('echartsCanvas');
  if (!chartInstance) {
    chartInstance = echarts.init(chartDom);
    window.addEventListener('resize', () => chartInstance.resize());
  }

  const palette = PALETTES[spec.palette || 'apa'] || PALETTES.apa;
  const option = buildEChartsOption(spec, palette);
  chartInstance.clear();
  chartInstance.setOption(option, true);
}

// ----------------------------------------------------
// 5. Complete 13 Scientific Chart ECharts Builder
// ----------------------------------------------------
function buildEChartsOption(spec, palette) {
  const type = spec.chartType;

  // Base options
  const baseOption = {
    color: palette,
    title: {
      text: spec.title || '',
      subtext: spec.subtitle || '',
      left: 'center',
      textStyle: { fontSize: 16, fontWeight: 700, color: '#1e293b' },
      subtextStyle: { fontSize: 12, color: '#64748b' }
    },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: {
      top: 36,
      textStyle: { color: '#475569', fontSize: 11 }
    },
    grid: { top: 80, right: 30, bottom: 45, left: 60, containLabel: true }
  };

  switch (type) {
    case 'bar':
      return {
        ...baseOption,
        xAxis: {
          type: 'category',
          data: spec.categories || [],
          name: spec.xAxisName || '',
          axisLabel: { interval: 0, rotate: (spec.categories && spec.categories.length > 6) ? 25 : 0 }
        },
        yAxis: { type: 'value', name: spec.yAxisName || '' },
        series: spec.series.map(s => ({
          name: s.name,
          type: 'bar',
          data: s.data,
          barMaxWidth: 45,
          itemStyle: { borderRadius: [4, 4, 0, 0] }
        }))
      };

    case 'radar':
      return {
        ...baseOption,
        tooltip: { trigger: 'item' },
        radar: {
          indicator: spec.radarIndicators && spec.radarIndicators.length > 0 
            ? spec.radarIndicators 
            : (spec.categories || []).map(c => ({ name: c, max: 100 })),
          center: ['50%', '55%'],
          radius: '65%'
        },
        series: [{
          name: spec.title || 'Métricas Radiales',
          type: 'radar',
          data: spec.series.map(s => ({
            name: s.name,
            value: s.data
          }))
        }]
      };

    case 'line':
      return {
        ...baseOption,
        tooltip: { trigger: 'axis' },
        xAxis: {
          type: 'category',
          data: spec.categories || [],
          name: spec.xAxisName || ''
        },
        yAxis: { type: 'value', name: spec.yAxisName || '' },
        series: spec.series.map(s => ({
          name: s.name,
          type: 'line',
          data: s.data,
          smooth: true,
          symbolSize: 7
        }))
      };

    case 'area':
      return {
        ...baseOption,
        tooltip: { trigger: 'axis' },
        xAxis: {
          type: 'category',
          boundaryGap: false,
          data: spec.categories || [],
          name: spec.xAxisName || ''
        },
        yAxis: { type: 'value', name: spec.yAxisName || '' },
        series: spec.series.map((s, idx) => ({
          name: s.name,
          type: 'line',
          data: s.data,
          smooth: true,
          areaStyle: { opacity: 0.35 },
          symbolSize: 6
        }))
      };

    case 'gantt':
      // Gantt chart rendered via category Y-axis and duration bars
      const tasks = (spec.ganttItems || []).map(g => g.task);
      const durations = (spec.ganttItems || []).map(g => g.durationDays || 10);
      return {
        ...baseOption,
        tooltip: {
          trigger: 'item',
          formatter: (params) => {
            const item = (spec.ganttItems || [])[params.dataIndex];
            if (item) {
              return `<strong>${item.task}</strong><br/>Inicio: ${item.startDate || 'N/A'}<br/>Fin: ${item.endDate || 'N/A'}<br/>Duración: ${item.durationDays || 0} días`;
            }
            return `${params.name}: ${params.value} días`;
          }
        },
        grid: { top: 80, right: 30, bottom: 45, left: 140, containLabel: true },
        xAxis: { type: 'value', name: 'Días / Duración' },
        yAxis: { type: 'category', data: tasks, inverse: true },
        series: [{
          name: 'Duración de Tarea (días)',
          type: 'bar',
          data: durations,
          itemStyle: { borderRadius: [0, 6, 6, 0] }
        }]
      };

    case 'pie':
    case 'doughnut':
      const isDoughnut = type === 'doughnut';
      let pieData = [];
      if (spec.series && spec.series.length > 0 && Array.isArray(spec.series[0].data)) {
        pieData = spec.series[0].data.map((d, i) => {
          if (typeof d === 'object' && d !== null && d.name) return d;
          return { name: (spec.categories && spec.categories[i]) || `Cat ${i+1}`, value: d };
        });
      }
      return {
        ...baseOption,
        tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
        series: [{
          name: spec.title || 'Proporción',
          type: 'pie',
          radius: isDoughnut ? ['40%', '70%'] : '65%',
          center: ['50%', '55%'],
          itemStyle: { borderRadius: 4, borderColor: '#fff', borderWidth: 2 },
          data: pieData
        }]
      };

    case 'funnel':
      let funnelData = [];
      if (spec.series && spec.series.length > 0 && Array.isArray(spec.series[0].data)) {
        funnelData = spec.series[0].data.map((d, i) => {
          if (typeof d === 'object' && d !== null && d.name) return d;
          return { name: (spec.categories && spec.categories[i]) || `Etapa ${i+1}`, value: d };
        });
      }
      return {
        ...baseOption,
        tooltip: { trigger: 'item', formatter: '{b}: {c}' },
        series: [{
          name: spec.title || 'Embudo',
          type: 'funnel',
          left: '15%',
          width: '70%',
          top: 90,
          bottom: 30,
          label: { show: true, position: 'inside' },
          data: funnelData
        }]
      };

    case 'waterfall':
      // Waterfall calculation: baseline transparent bars + delta bars
      const cats = spec.categories || [];
      const vals = (spec.series && spec.series[0] ? spec.series[0].data : []).map(v => typeof v === 'number' ? v : parseFloat(v) || 0);
      let cumulative = 0;
      const baseVals = [];
      const barVals = [];

      vals.forEach((val, i) => {
        if (i === 0 || i === vals.length - 1) {
          baseVals.push(0);
          barVals.push(val);
          cumulative = val;
        } else if (val >= 0) {
          baseVals.push(cumulative);
          barVals.push(val);
          cumulative += val;
        } else {
          cumulative += val;
          baseVals.push(cumulative);
          barVals.push(Math.abs(val));
        }
      });

      return {
        ...baseOption,
        tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
        xAxis: { type: 'category', data: cats },
        yAxis: { type: 'value', name: spec.yAxisName || '' },
        series: [
          {
            name: 'Base',
            type: 'bar',
            stack: 'total',
            itemStyle: { borderColor: 'transparent', color: 'transparent' },
            data: baseVals
          },
          {
            name: 'Variación',
            type: 'bar',
            stack: 'total',
            data: barVals
          }
        ]
      };

    case 'scatter':
      return {
        ...baseOption,
        tooltip: { trigger: 'item', formatter: (p) => `X: ${p.value[0]}<br/>Y: ${p.value[1]}` },
        xAxis: { type: 'value', name: spec.xAxisName || 'Variable X' },
        yAxis: { type: 'value', name: spec.yAxisName || 'Variable Y' },
        series: spec.series.map(s => ({
          name: s.name || 'Dispersión',
          type: 'scatter',
          symbolSize: 10,
          data: s.data
        }))
      };

    case 'histogram':
      return {
        ...baseOption,
        xAxis: { type: 'category', data: spec.categories || [], name: spec.xAxisName || 'Intervalos' },
        yAxis: { type: 'value', name: spec.yAxisName || 'Frecuencia' },
        series: [{
          name: 'Frecuencia',
          type: 'bar',
          data: (spec.series && spec.series[0]) ? spec.series[0].data : [],
          barMaxWidth: 60,
          itemStyle: { borderRadius: [2, 2, 0, 0] }
        }]
      };

    case 'boxplot':
      let bpData = [];
      if (spec.series && spec.series[0] && Array.isArray(spec.series[0].data)) {
        bpData = spec.series[0].data;
      }
      return {
        ...baseOption,
        tooltip: {
          trigger: 'item',
          formatter: (p) => {
            return `<strong>${p.name}</strong><br/>
                    Máx: ${p.value[5]}<br/>
                    Q3: ${p.value[4]}<br/>
                    Mediana: ${p.value[3]}<br/>
                    Q1: ${p.value[2]}<br/>
                    Mín: ${p.value[1]}`;
          }
        },
        xAxis: { type: 'category', data: spec.categories || [] },
        yAxis: { type: 'value', name: spec.yAxisName || 'Valores' },
        series: [{
          name: 'Distribución',
          type: 'boxplot',
          data: bpData,
          itemStyle: { borderColor: palette[0], borderWidth: 1.5 }
        }]
      };

    case 'heatmap':
      const maxVal = (spec.heatmapData || []).reduce((max, p) => p.value > max ? p.value : max, 10);
      return {
        ...baseOption,
        tooltip: { position: 'top' },
        grid: { top: 70, bottom: 80, left: 80, right: 80 },
        xAxis: { type: 'category', data: spec.xAxisLabels || [], splitArea: { show: true } },
        yAxis: { type: 'category', data: spec.yAxisLabels || [], splitArea: { show: true } },
        visualMap: {
          min: 0,
          max: maxVal,
          calculable: true,
          orient: 'horizontal',
          left: 'center',
          bottom: 15,
          inRange: { color: ['#eff6ff', '#3b82f6', '#1e3a8a'] }
        },
        series: [{
          name: 'Intensidad',
          type: 'heatmap',
          data: (spec.heatmapData || []).map(p => [p.x, p.y, p.value]),
          label: { show: true }
        }]
      };

    default:
      return baseOption;
  }
}

// ----------------------------------------------------
// 6. Live Table Editor
// ----------------------------------------------------
function renderTableEditor(tableUsed) {
  const container = document.getElementById('tableEditorContainer');
  container.innerHTML = '';

  if (!tableUsed || !tableUsed.headers) return;

  const table = document.createElement('table');
  table.className = 'editable-table';

  // Thead
  const thead = document.createElement('thead');
  const trHead = document.createElement('tr');
  tableUsed.headers.forEach((h, colIdx) => {
    const th = document.createElement('th');
    th.textContent = h;
    trHead.appendChild(th);
  });
  thead.appendChild(trHead);
  table.appendChild(thead);

  // Tbody
  const tbody = document.createElement('tbody');
  (tableUsed.rows || []).forEach((row, rowIdx) => {
    const tr = document.createElement('tr');
    tableUsed.headers.forEach((h, colIdx) => {
      const td = document.createElement('td');
      const input = document.createElement('input');
      input.value = (row && row[colIdx] !== undefined) ? row[colIdx] : '';
      input.addEventListener('change', (e) => {
        const val = e.target.value.trim();
        const num = parseFloat(val.replace(',', '.'));
        tableUsed.rows[rowIdx][colIdx] = isNaN(num) ? val : num;
        syncTableToChart();
      });
      td.appendChild(input);
      tr.appendChild(td);
    });
    tbody.appendChild(tr);
  });
  table.appendChild(tbody);
  container.appendChild(table);

  document.getElementById('tableSourceRef').textContent = tableUsed.sourceLocation || 'Documento fuente';
}

function syncTableToChart() {
  if (!currentChartSpec || !currentChartSpec.tableUsed) return;
  const tbl = currentChartSpec.tableUsed;
  
  // Re-extract categories and series from the updated table
  const newCats = [];
  const seriesValues = [];

  tbl.rows.forEach(r => {
    if (r.length > 0) newCats.push(String(r[0]));
    if (r.length > 1) seriesValues.push(parseFloat(r[1]) || 0);
  });

  currentChartSpec.categories = newCats;
  if (currentChartSpec.series && currentChartSpec.series.length > 0) {
    if (currentChartSpec.chartType === 'pie' || currentChartSpec.chartType === 'doughnut') {
      currentChartSpec.series[0].data = newCats.map((c, i) => ({ name: c, value: seriesValues[i] || 0 }));
    } else {
      currentChartSpec.series[0].data = seriesValues;
    }
  }

  renderECharts(currentChartSpec);
}

// ----------------------------------------------------
// 7. Render Methodology & Auditing Panel
// ----------------------------------------------------
function renderMethodology(spec) {
  const meth = spec.methodology || {};
  document.getElementById('methExplanation').textContent = meth.explanation || 'Datos procesados directamente.';
  document.getElementById('methAggregation').textContent = meth.aggregationMethod || 'Directo';
  document.getElementById('methCitation').textContent = meth.academicCitationNote || `Nota: Elaboración propia a partir de ${spec.tableUsed.sourceLocation || 'fuente original'}.`;

  const mapList = document.getElementById('methVariablesMap');
  mapList.innerHTML = '';
  if (meth.variablesMapping) {
    for (const [k, v] of Object.entries(meth.variablesMapping)) {
      const item = document.createElement('div');
      item.innerHTML = `<span class="meta-label">${k}:</span> <span class="meta-value">${v}</span>`;
      mapList.appendChild(item);
    }
  }
}

// ----------------------------------------------------
// 8. Toolbar & Controls Setup
// ----------------------------------------------------
function initChartToolbar() {
  // Chart type switcher
  document.querySelectorAll('.type-chip').forEach(chip => {
    chip.addEventListener('click', async () => {
      const targetType = chip.dataset.type;
      if (!currentChartSpec) return;

      showLoading('Transformando formato de gráfica...');
      try {
        const res = await fetch('/api/chart/transform', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ chartSpec: currentChartSpec, targetType: targetType })
        });
        if (!res.ok) throw new Error('Error al transformar tipo');
        const updated = await res.json();
        currentChartSpec = updated;
        renderChartWorkspace(updated);
      } catch (err) {
        alert('Error: ' + err.message);
      } finally {
        hideLoading();
      }
    });
  });

  // Palette selector
  const paletteSelect = document.getElementById('paletteSelect');
  paletteSelect.addEventListener('change', (e) => {
    if (currentChartSpec) {
      currentChartSpec.palette = e.target.value;
      renderECharts(currentChartSpec);
    }
  });

  // Generate / DeepSeek trigger button in inspection section
  document.getElementById('btnTriggerDeepSeek').addEventListener('click', () => {
    generateChartWithAI();
  });
}

function initFormControls() {
  // Live Title and Subtitle inputs
  document.getElementById('chartTitleInput').addEventListener('input', (e) => {
    if (currentChartSpec) {
      currentChartSpec.title = e.target.value;
      renderECharts(currentChartSpec);
    }
  });

  document.getElementById('chartSubtitleInput').addEventListener('input', (e) => {
    if (currentChartSpec) {
      currentChartSpec.subtitle = e.target.value;
      renderECharts(currentChartSpec);
    }
  });

  // Export Excel with embedded chart picture and full AI info box
  document.getElementById('btnExportExcel').addEventListener('click', async () => {
    if (!currentChartSpec) return;
    showLoading('Generando libro Excel con gráfica incrustada, datos y cuadro de información...');
    try {
      let chartImgBase64 = '';
      if (chartInstance) {
        chartImgBase64 = chartInstance.getDataURL({
          type: 'png',
          pixelRatio: 2.0,
          backgroundColor: '#ffffff'
        });
      }

      const exportPayload = {
        ...currentChartSpec,
        chartImageBase64: chartImgBase64
      };

      const res = await fetch('/api/chart/export-excel', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(exportPayload)
      });
      if (!res.ok) throw new Error('Error en exportación a Excel');

      const blob = await res.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${(currentChartSpec.title || 'grafica').replace(/[^a-zA-Z0-9]/g, '_')}_investigacion.xlsx`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (err) {
      alert('Error descargando Excel: ' + err.message);
    } finally {
      hideLoading();
    }
  });

  // Copy AI Explanation for Paper
  const btnCopyExp = document.getElementById('btnCopyExplanation');
  if (btnCopyExp) {
    btnCopyExp.addEventListener('click', () => {
      const exp = document.getElementById('aiChartExplanationText').textContent;
      const findings = document.getElementById('aiKeyFindingsText').textContent;
      const narrative = document.getElementById('aiPaperNarrativeText').textContent;
      const fullText = `EXPLICACIÓN DE LA GRÁFICA:\n${exp}\n\nHALLAZGOS CLAVE:\n${findings}\n\nREDACCIÓN PARA EL DOCUMENTO:\n${narrative}`;
      navigator.clipboard.writeText(fullText).then(() => {
        alert('¡Explicación completa de la gráfica copiada al portapapeles!');
      });
    });
  }

  // Export Image PNG
  document.getElementById('btnExportImage').addEventListener('click', () => {
    if (!chartInstance) return;
    const url = chartInstance.getDataURL({
      type: 'png',
      pixelRatio: 2.5,
      backgroundColor: '#ffffff'
    });
    const a = document.createElement('a');
    a.href = url;
    a.download = `${(currentChartSpec.title || 'figura').replace(/[^a-zA-Z0-9]/g, '_')}.png`;
    document.body.appendChild(a);
    a.click();
    a.remove();
  });

  // Copy Citation Note
  document.getElementById('btnCopyCitation').addEventListener('click', () => {
    const note = document.getElementById('methCitation').textContent;
    navigator.clipboard.writeText(note).then(() => {
      alert('¡Cita metodológica copiada al portapapeles!');
    });
  });
}

// ----------------------------------------------------
// 9. Sample Datasets for Instant 1-Click Testing
// ----------------------------------------------------
function initSampleButtons() {
  document.querySelectorAll('.sample-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const sampleKey = btn.dataset.sample;
      loadSampleData(sampleKey);
    });
  });
}

function loadSampleData(key) {
  let mockInspection = null;

  if (key === 'clinical') {
    mockInspection = {
      fileType: 'excel',
      fileName: 'Ensayo_Clinico_Fase3.xlsx',
      summary: 'Excel analizado: 2 hojas detectadas. Datos de eficacia y retención de pacientes por grupo.',
      tables: [
        {
          index: 1,
          name: 'Eficacia y Retención',
          headers: ['Grupo de Tratamiento', 'Tasa de Respuesta (%)', 'Adherencia (%)', 'Eventos Adversos (%)'],
          previewRows: [
            ['Placebo Control', 28.4, 82.1, 8.5],
            ['Dosis Baja (25mg)', 54.2, 88.0, 11.2],
            ['Dosis Media (50mg)', 72.8, 91.5, 14.8],
            ['Dosis Alta (100mg)', 89.1, 85.3, 21.0]
          ],
          totalRows: 4,
          totalCols: 4,
          sourceRef: "Hoja 'Eficacia' (Filas 2 a 5)"
        }
      ]
    };
  } else if (key === 'gantt') {
    mockInspection = {
      fileType: 'word',
      fileName: 'Protocolo_Tesis_Cronograma.docx',
      summary: 'Word analizado: Cronograma de fases y actividades de la investigación doctoral.',
      tables: [
        {
          index: 1,
          name: 'Cronograma de Fases',
          headers: ['Fase Metodológica', 'Duración (días)', 'Fecha Inicio', 'Fecha Término'],
          previewRows: [
            ['Revisión Sistemática', 45, '2026-01-01', '2026-02-15'],
            ['Diseño de Muestreo', 30, '2026-02-15', '2026-03-15'],
            ['Levantamiento de Datos', 60, '2026-03-15', '2026-05-15'],
            ['Análisis Estadístico', 40, '2026-05-15', '2026-06-25'],
            ['Redacción del Manuscrito', 50, '2026-06-25', '2026-08-15']
          ],
          totalRows: 5,
          totalCols: 4,
          sourceRef: 'Documento Word: Tabla 1 (Cronograma de Actividades)'
        }
      ],
      sections: [
        {
          index: 1,
          title: 'Metodología y Plan de Trabajo',
          content: 'El proyecto se desarrollará a lo largo de 8 meses divididos en 5 fases secuenciales desde la revisión sistemática hasta la publicación final.'
        }
      ]
    };
  } else if (key === 'multivar') {
    mockInspection = {
      fileType: 'excel',
      fileName: 'Evaluacion_Multidimensional.xlsx',
      summary: 'Excel analizado: Matriz de competencias y variables evaluadas en 5 departamentos de investigación.',
      tables: [
        {
          index: 1,
          name: 'Métricas Multidimensionales',
          headers: ['Departamento', 'Publicaciones Q1', 'Citaciones', 'Financiamiento (k$)', 'Patentes', 'Docencia'],
          previewRows: [
            ['Biotecnología', 88, 92, 95, 78, 80],
            ['Ciencias de la Computación', 94, 96, 90, 85, 88],
            ['Ingeniería Biomédica', 82, 85, 89, 90, 84]
          ],
          totalRows: 3,
          totalCols: 6,
          sourceRef: "Hoja 'Rankings' (Filas 1 a 4)"
        }
      ]
    };
  }

  if (mockInspection) {
    currentInspection = mockInspection;
    renderInspectionUI(mockInspection);
  }
}

// ----------------------------------------------------
// UI Helpers
// ----------------------------------------------------
function showLoading(msg) {
  const overlay = document.getElementById('loadingOverlay');
  document.getElementById('loadingMessage').textContent = msg || 'Procesando...';
  overlay.style.display = 'flex';
}

function hideLoading() {
  document.getElementById('loadingOverlay').style.display = 'none';
}
