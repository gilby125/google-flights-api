const API_BASE = '/api/v1/admin';
const META_BASE = '/api/v1';

const bulkElements = {
    runsTable: document.getElementById('bulkRunsTable'),
    runsCount: document.getElementById('bulkRunsCount'),
    runFilterInput: document.getElementById('runFilterInput'),
    runFilterButtons: document.querySelectorAll('[data-run-filter]'),
    resultsTable: document.getElementById('bulkResultsTable'),
    resultsCount: document.getElementById('bulkResultsCount'),
    resultsInfo: document.getElementById('bulkResultsInfo'),
    offersTable: document.getElementById('bulkOffersTable'),
    offersCount: document.getElementById('bulkOffersCount'),
    resultFilterInput: document.getElementById('resultFilterInput'),
    resultCurrencyFilter: document.getElementById('resultCurrencyFilter'),
    summary: document.getElementById('bulkSummary'),
    refreshBtn: document.getElementById('refreshBulkBtn'),
    saveBulkJobBtn: document.getElementById('saveBulkJobBtn'),
};

let bulkRuns = [];
let runsSortField = 'created';
let runsSortDir = 'desc';
let runsFilterText = '';
let runsStatusFilter = 'all';

let currentResults = [];
let currentOffers = [];
let currentRunId = null;
let resultsSortField = 'price';
let resultsSortDir = 'asc';
let resultsFilterText = '';
let resultsCurrencyFilter = '';

async function initBulkPage() {
    if (bulkElements.refreshBtn) {
        bulkElements.refreshBtn.addEventListener('click', loadBulkRuns);
    }
    if (bulkElements.saveBulkJobBtn) {
        bulkElements.saveBulkJobBtn.addEventListener('click', submitBulkJob);
    }
    if (bulkElements.runFilterInput) {
        bulkElements.runFilterInput.addEventListener('input', event => {
            runsFilterText = event.target.value.toLowerCase();
            renderBulkRuns();
        });
    }
    bulkElements.runFilterButtons.forEach(button => {
        button.addEventListener('click', () => {
            runsStatusFilter = button.dataset.runFilter;
            bulkElements.runFilterButtons.forEach(btn => btn.classList.toggle('active', btn === button));
            renderBulkRuns();
        });
    });
    document.querySelectorAll('.sortable').forEach(header => {
        header.addEventListener('click', () => updateRunsSort(header.dataset.sort));
    });
    document.querySelectorAll('.sortable-result').forEach(header => {
        header.addEventListener('click', () => updateResultsSort(header.dataset.sort));
    });
    if (bulkElements.resultFilterInput) {
        bulkElements.resultFilterInput.addEventListener('input', event => {
            resultsFilterText = event.target.value.toLowerCase();
            renderBulkResults();
        });
    }
    if (bulkElements.resultCurrencyFilter) {
        bulkElements.resultCurrencyFilter.addEventListener('change', event => {
            resultsCurrencyFilter = event.target.value;
            renderBulkResults();
        });
    }

    initBulkMacroTokenHelpers();

    await loadBulkRuns();
}

function showAlert(message, type = 'info') {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
    alertDiv.setAttribute('role', 'alert');
    alertDiv.innerHTML = `
        <div>${message}</div>
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    `;

    const main = document.querySelector('main');
    if (main) {
        main.insertBefore(alertDiv, main.firstChild);
    } else {
        document.body.insertBefore(alertDiv, document.body.firstChild);
    }

    setTimeout(() => {
        alertDiv.classList.remove('show');
        setTimeout(() => alertDiv.remove(), 150);
    }, 6000);
}

function parseAirportList(input) {
    return input
        .split(/[\n,]/)
        .map(code => code.trim().toUpperCase())
        .filter(code => code.length > 0);
}

function initBulkMacroTokenHelpers() {
    const originSelect = document.getElementById('bulkOriginRegionSelect');
    const originAddBtn = document.getElementById('bulkOriginRegionAddBtn');
    const destSelect = document.getElementById('bulkDestinationRegionSelect');
    const destAddBtn = document.getElementById('bulkDestinationRegionAddBtn');

    if (originAddBtn) {
        originAddBtn.addEventListener('click', () => {
            const token = originSelect?.value || '';
            if (!token) return;
            appendTokenToTextarea('bulkOrigins', token);
        });
    }

    if (destAddBtn) {
        destAddBtn.addEventListener('click', () => {
            const token = destSelect?.value || '';
            if (!token) return;
            appendTokenToTextarea('bulkDestinations', token);
        });
    }

    loadRegionsForBulk().catch(err => console.warn('Failed to load regions', err));
}

async function loadRegionsForBulk() {
    const originSelect = document.getElementById('bulkOriginRegionSelect');
    const originAddBtn = document.getElementById('bulkOriginRegionAddBtn');
    const destSelect = document.getElementById('bulkDestinationRegionSelect');
    const destAddBtn = document.getElementById('bulkDestinationRegionAddBtn');

    if (!originSelect && !destSelect) return;

    const response = await fetch(`${META_BASE}/regions`);
    if (!response.ok) {
        throw new Error(`Failed to load regions (HTTP ${response.status})`);
    }

    const regions = await response.json();
    if (!Array.isArray(regions) || regions.length === 0) return;

    populateRegionSelect(originSelect, regions);
    populateRegionSelect(destSelect, regions);

    if (originSelect) originSelect.disabled = false;
    if (destSelect) destSelect.disabled = false;
    if (originAddBtn) originAddBtn.disabled = false;
    if (destAddBtn) destAddBtn.disabled = false;
}

function populateRegionSelect(selectEl, regions) {
    if (!selectEl) return;

    const currentValue = selectEl.value;
    selectEl.innerHTML = '<option value="">Add a region…</option>';

    regions.forEach(region => {
        if (!region?.token) return;
        const count = typeof region.airport_count === 'number' ? region.airport_count : null;
        const samples = Array.isArray(region.sample_airports) ? region.sample_airports.slice(0, 3) : [];
        const labelParts = [region.token];
        if (count != null) labelParts.push(`(${count})`);
        if (samples.length) labelParts.push(`— ${samples.join(', ')}`);

        const option = document.createElement('option');
        option.value = region.token;
        option.textContent = labelParts.join(' ');
        selectEl.appendChild(option);
    });

    if (currentValue) {
        selectEl.value = currentValue;
    }
}

function appendTokenToTextarea(textareaId, token) {
    const textarea = document.getElementById(textareaId);
    if (!textarea) return;

    const normalized = String(token).trim().toUpperCase();
    if (!normalized) return;

    const existing = parseAirportList(textarea.value || '');
    if (existing.includes(normalized)) {
        return;
    }

    const next = existing.concat([normalized]);
    textarea.value = next.join(', ');
}

function buildCronExpression(interval, timeValue) {
    if (!timeValue) return '0 12 * * *';
    const [hour, minute] = timeValue.split(':');
    switch (interval) {
        case 'weekly':
            return `${minute} ${hour} * * 1`;
        case 'monthly':
            return `${minute} ${hour} 1 * *`;
        case 'daily':
        default:
            return `${minute} ${hour} * * *`;
    }
}

function updateBulkCronPreview() {
    const intervalInput = document.getElementById('bulkInterval');
    const timeInput = document.getElementById('bulkTime');
    const preview = document.getElementById('bulkCronPreview');
    if (!intervalInput || !timeInput || !preview) return;
    preview.textContent = buildCronExpression(intervalInput.value, timeInput.value);
}

async function submitBulkJob() {
    const form = document.getElementById('newBulkJobForm');
    if (!form) return;
    if (!form.checkValidity()) {
        form.reportValidity();
        return;
    }

    const payload = {
        name: document.getElementById('bulkJobName')?.value?.trim(),
        origins: parseAirportList(document.getElementById('bulkOrigins')?.value || ''),
        destinations: parseAirportList(document.getElementById('bulkDestinations')?.value || ''),
        date_start: document.getElementById('bulkDateStart')?.value,
        date_end: document.getElementById('bulkDateEnd')?.value,
        return_date_start: document.getElementById('bulkReturnDateStart')?.value || '',
        return_date_end: document.getElementById('bulkReturnDateEnd')?.value || '',
        trip_length: parseInt(document.getElementById('bulkTripLength')?.value || '0', 10),
        adults: parseInt(document.getElementById('bulkAdults')?.value || '1', 10),
        children: parseInt(document.getElementById('bulkChildren')?.value || '0', 10),
        infants_lap: parseInt(document.getElementById('bulkInfantsLap')?.value || '0', 10),
        infants_seat: parseInt(document.getElementById('bulkInfantsSeat')?.value || '0', 10),
        trip_type: document.getElementById('bulkTripType')?.value || 'round_trip',
        class: document.getElementById('bulkClass')?.value || 'economy',
        stops: document.getElementById('bulkStops')?.value || 'any',
        currency: document.getElementById('bulkCurrency')?.value || 'USD',
    };

    const interval = document.getElementById('bulkInterval')?.value || 'daily';
    const timeValue = document.getElementById('bulkTime')?.value || '12:00';
    payload.cron_expression = buildCronExpression(interval, timeValue);

    if (!payload.name) {
        showAlert('Job name is required.', 'warning');
        return;
    }
    if (!payload.origins.length || !payload.destinations.length) {
        showAlert('At least one origin and destination are required.', 'warning');
        return;
    }

    const btn = bulkElements.saveBulkJobBtn;
    if (btn) {
        btn.disabled = true;
        btn.dataset.originalText = btn.dataset.originalText || btn.textContent;
        btn.textContent = 'Saving...';
    }

    try {
        const response = await fetch(`${API_BASE}/bulk-jobs`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });

        const responseText = await response.text();
        let responseBody = {};
        try {
            responseBody = JSON.parse(responseText);
        } catch (_) {
            responseBody = { raw: responseText };
        }

        if (!response.ok) {
            throw new Error(responseBody?.error || `HTTP ${response.status}`);
        }

        showAlert(responseBody?.message || 'Bulk job scheduled.', 'success');

        const modalEl = document.getElementById('newBulkJobModal');
        if (modalEl && window.bootstrap?.Modal) {
            const instance = window.bootstrap.Modal.getInstance(modalEl) || new window.bootstrap.Modal(modalEl);
            instance.hide();
        }

        await loadBulkRuns();
    } catch (err) {
        showAlert(`Failed to schedule bulk job: ${err.message}`, 'danger');
    } finally {
        if (btn) {
            btn.disabled = false;
            btn.textContent = btn.dataset.originalText || 'Save Bulk Job';
        }
    }
}

function updateRunsSort(field) {
    if (runsSortField === field) {
        runsSortDir = runsSortDir === 'asc' ? 'desc' : 'asc';
    } else {
        runsSortField = field;
        runsSortDir = field === 'id' ? 'desc' : 'asc';
    }
    renderBulkRuns();
}

function updateResultsSort(field) {
    if (resultsSortField === field) {
        resultsSortDir = resultsSortDir === 'asc' ? 'desc' : 'asc';
    } else {
        resultsSortField = field;
        resultsSortDir = field === 'price' ? 'asc' : 'desc';
    }
    renderBulkResults();
}

async function loadBulkRuns() {
    try {
        const response = await fetch(`${API_BASE}/bulk-jobs?limit=200`);
        if (!response.ok) {
            throw new Error(`Failed to load bulk runs (${response.status})`);
        }
        const data = await response.json();
        bulkRuns = Array.isArray(data.items) ? data.items : [];
        bulkElements.runsCount.textContent = bulkRuns.length;
        renderBulkRuns();
    } catch (error) {
        console.error('Error loading bulk jobs:', error);
        bulkElements.runsTable.innerHTML = `
            <tr>
                <td colspan="10" class="text-center text-danger py-3">
                    ${error.message}
                </td>
            </tr>
        `;
    }
}

function renderBulkRuns() {
    if (!bulkRuns.length) {
        bulkElements.runsTable.innerHTML = `
            <tr>
                <td colspan="10" class="text-center text-muted py-3">
                    No bulk runs found.
                </td>
            </tr>
        `;
        return;
    }

    const filtered = bulkRuns
        .filter(run => {
            if (runsStatusFilter !== 'all' && (run.status || '').toLowerCase() !== runsStatusFilter) {
                return false;
            }
            if (!runsFilterText) return true;
            const searchTarget = [
                run.status,
                run.currency,
                run.id,
                run.job_id,
            ].join(' ').toLowerCase();
            return searchTarget.includes(runsFilterText);
        })
        .sort((a, b) => compareRuns(a, b, runsSortField, runsSortDir));

    if (!filtered.length) {
        bulkElements.runsTable.innerHTML = `
            <tr>
                <td colspan="10" class="text-center text-muted py-3">
                    No runs match your filters.
                </td>
            </tr>
        `;
        return;
    }

    bulkElements.runsTable.innerHTML = '';
    filtered.forEach(run => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${run.id}</td>
            <td><span class="badge ${statusBadgeClass(run.status)}">${run.status}</span></td>
            <td>${run.total_routes ?? '—'}</td>
            <td>${run.completed ?? 0}</td>
            <td>${run.total_offers ?? 0}</td>
            <td>${run.error_count ?? 0}</td>
            <td>${run.currency ?? '—'}</td>
            <td>${formatDate(run.created_at)}</td>
            <td>${run.completed_at ? formatDate(run.completed_at) : '—'}</td>
            <td>
                <button class="btn btn-sm btn-outline-primary" data-run-id="${run.id}">
                    View
                </button>
            </td>
        `;
        const viewBtn = row.querySelector('button');
        viewBtn.addEventListener('click', () => loadBulkResults(run.id));
        bulkElements.runsTable.appendChild(row);
    });
}

function compareRuns(a, b, field, dir) {
    const direction = dir === 'asc' ? 1 : -1;
    switch (field) {
        case 'status':
            return direction * compareStrings(a.status, b.status);
        case 'routes':
            return direction * compareNumbers(a.total_routes, b.total_routes);
        case 'completed':
            return direction * compareNumbers(a.completed, b.completed);
        case 'offers':
            return direction * compareNumbers(a.total_offers, b.total_offers);
        case 'errors':
            return direction * compareNumbers(a.error_count, b.error_count);
        case 'currency':
            return direction * compareStrings(a.currency, b.currency);
        case 'created':
            return direction * compareDates(a.created_at, b.created_at);
        case 'completed_at':
            return direction * compareDates(a.completed_at, b.completed_at);
        case 'id':
        default:
            return direction * compareNumbers(a.id, b.id);
    }
}

async function loadBulkResults(runId) {
    try {
        currentRunId = runId;
        resultsFilterText = '';
        resultsCurrencyFilter = '';
        resultsSortField = 'price';
        resultsSortDir = 'asc';
        if (bulkElements.resultFilterInput) {
            bulkElements.resultFilterInput.value = '';
        }
        if (bulkElements.resultCurrencyFilter) {
            bulkElements.resultCurrencyFilter.value = '';
        }
        bulkElements.summary.innerHTML = `<span class="text-muted">Loading results for bulk run #${runId}...</span>`;
        bulkElements.resultsTable.innerHTML = `
            <tr>
                <td colspan="8" class="text-center text-muted py-3">Loading...</td>
            </tr>
        `;

        const response = await fetch(`${API_BASE}/bulk-jobs/${runId}?limit=500`);
        if (!response.ok) {
            throw new Error(`Failed to load results (${response.status})`);
        }
        const data = await response.json();
        currentResults = Array.isArray(data.results) ? data.results : [];
        renderBulkSummary(data.summary || {}, currentResults.length);
        renderBulkResults();
        loadBulkOffers(runId);
    } catch (error) {
        console.error('Error loading bulk results:', error);
        bulkElements.summary.innerHTML = `<span class="text-danger">${error.message}</span>`;
        bulkElements.resultsTable.innerHTML = `
            <tr>
                <td colspan="8" class="text-center text-danger py-3">
                    Failed to load results.
                </td>
            </tr>
        `;
        currentResults = [];
        currentRunId = null;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    initBulkPage();
    const intervalInput = document.getElementById('bulkInterval');
    const timeInput = document.getElementById('bulkTime');
    if (intervalInput) intervalInput.addEventListener('change', updateBulkCronPreview);
    if (timeInput) timeInput.addEventListener('change', updateBulkCronPreview);
    updateBulkCronPreview();
});

async function loadBulkOffers(runId) {
    if (!bulkElements.offersTable) {
        return;
    }

    bulkElements.offersTable.innerHTML = `
        <tr>
            <td colspan="8" class="text-center text-muted py-3">Loading offers...</td>
        </tr>
    `;
    if (bulkElements.offersCount) {
        bulkElements.offersCount.textContent = '0';
    }

    try {
        const response = await fetch(`${API_BASE}/bulk-jobs/${runId}/offers?limit=500`);
        if (!response.ok) {
            throw new Error(`Failed to load offers (${response.status})`);
        }
        const data = await response.json();
        currentOffers = Array.isArray(data.grid) ? data.grid : [];
        renderBulkOffers();
    } catch (error) {
        console.error('Error loading bulk offers:', error);
        bulkElements.offersTable.innerHTML = `
            <tr>
                <td colspan="8" class="text-center text-danger py-3">${error.message}</td>
            </tr>
        `;
    }
}

function renderBulkOffers() {
    if (!bulkElements.offersTable) {
        return;
    }

    const totalCells = currentOffers.reduce((total, route) => {
        if (!route || !Array.isArray(route.cells)) {
            return total;
        }
        return total + route.cells.length;
    }, 0);

    if (!totalCells) {
        bulkElements.offersTable.innerHTML = `
            <tr>
                <td colspan="8" class="text-center text-muted py-3">
                    No offers recorded for this run.
                </td>
            </tr>
        `;
        if (bulkElements.offersCount) {
            bulkElements.offersCount.textContent = '0';
        }
        return;
    }

    const fragment = document.createDocumentFragment();
    currentOffers.forEach(route => {
        const routeName = `${route.origin ?? '???'} → ${route.destination ?? '???'}`;
        const cells = Array.isArray(route.cells) ? [...route.cells] : [];

        const headerRow = document.createElement('tr');
        headerRow.classList.add('table-active');
        headerRow.innerHTML = `
            <td colspan="8">
                <strong>${routeName}</strong>
                <span class="text-muted ms-2">${cells.length} fare${cells.length === 1 ? '' : 's'}</span>
            </td>
        `;
        fragment.appendChild(headerRow);

        cells.sort((a, b) => {
            const depA = a && a.departure_date ? new Date(a.departure_date).getTime() : 0;
            const depB = b && b.departure_date ? new Date(b.departure_date).getTime() : 0;
            if (depA !== depB) {
                return depA - depB;
            }
            const retA = a && a.return_date ? new Date(a.return_date).getTime() : 0;
            const retB = b && b.return_date ? new Date(b.return_date).getTime() : 0;
            return retA - retB;
        });

        if (!cells.length) {
            const emptyRow = document.createElement('tr');
            emptyRow.innerHTML = `
                <td colspan="8" class="text-center text-muted py-3">
                    No fares captured for ${routeName}.
                </td>
            `;
            fragment.appendChild(emptyRow);
            return;
        }

        cells.forEach(cell => {
            const departureDisplay = formatDate(cell ? cell.departure_date : null);
            const returnDisplay = cell && cell.return_date ? formatDate(cell.return_date) : '—';
            const priceDisplay = formatNumber(cell ? cell.price : null);
            const currencyDisplay = cell && cell.currency ? cell.currency : '—';
            const airlinesDisplay =
                cell && Array.isArray(cell.airline_codes) && cell.airline_codes.length
                    ? cell.airline_codes.join(', ')
                    : '—';
            const createdDisplay = cell && cell.created_at ? formatDate(cell.created_at) : '—';

            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${route.origin ?? '—'}</td>
                <td>${route.destination ?? '—'}</td>
                <td>${departureDisplay}</td>
                <td>${returnDisplay}</td>
                <td>${priceDisplay}</td>
                <td>${currencyDisplay}</td>
                <td>${airlinesDisplay}</td>
                <td>${createdDisplay}</td>
            `;
            fragment.appendChild(row);
        });
    });

    bulkElements.offersTable.innerHTML = '';
    bulkElements.offersTable.appendChild(fragment);
    if (bulkElements.offersCount) {
        bulkElements.offersCount.textContent = totalCells.toString();
    }
}

function renderBulkSummary(summary, resultCount) {
    const parts = [
        `Run #<strong>${summary.id ?? '—'}</strong>`,
        `Status: <span class="badge ${statusBadgeClass(summary.status)}">${summary.status ?? 'unknown'}</span>`,
        `Routes: ${summary.completed ?? 0} / ${summary.total_routes ?? 0}`,
        `Offers: ${summary.total_offers ?? 0}`,
        `Errors: ${summary.error_count ?? 0}`,
    ];

    if (summary.min_price != null) {
        parts.push(`Min: ${formatNumber(summary.min_price)}`);
    }
    if (summary.max_price != null) {
        parts.push(`Max: ${formatNumber(summary.max_price)}`);
    }
    if (summary.average_price != null) {
        parts.push(`Avg: ${formatNumber(summary.average_price)}`);
    }

    parts.push(`Currency: ${summary.currency ?? '—'}`);
    parts.push(`Rows shown: ${resultCount}`);

    bulkElements.summary.innerHTML = parts.join(' &bull; ');
}

function renderBulkResults() {
    const results = currentResults
        .filter(result => {
            if (resultsCurrencyFilter && result.currency !== resultsCurrencyFilter) {
                return false;
            }
            if (!resultsFilterText) return true;
            const searchTarget = [
                result.origin,
                result.destination,
                result.airline_code,
                result.currency,
                formatDate(result.departure_date),
                result.price,
            ]
                .join(' ')
                .toLowerCase();
            return searchTarget.includes(resultsFilterText);
        })
        .sort((a, b) => compareResults(a, b, resultsSortField, resultsSortDir));

    bulkElements.resultsCount.textContent = results.length;
    bulkElements.resultsInfo.textContent = currentRunId ? `Run #${currentRunId}` : '';

    if (!results.length) {
        bulkElements.resultsTable.innerHTML = `
            <tr>
                <td colspan="8" class="text-center text-muted py-3">
                    No results match your filters.
                </td>
            </tr>
        `;
        return;
    }

    bulkElements.resultsTable.innerHTML = '';
    results.forEach(result => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${result.origin}</td>
            <td>${result.destination}</td>
            <td>${formatDate(result.departure_date)}</td>
            <td>${result.return_date ? formatDate(result.return_date) : '—'}</td>
            <td>${formatNumber(result.price)}</td>
            <td>${result.currency ?? '—'}</td>
            <td>${result.airline_code ?? '—'}</td>
            <td>${result.duration ?? '—'}</td>
        `;
        bulkElements.resultsTable.appendChild(row);
    });
}

function compareResults(a, b, field, dir) {
    const direction = dir === 'asc' ? 1 : -1;
    switch (field) {
        case 'origin':
            return direction * compareStrings(a.origin, b.origin);
        case 'destination':
            return direction * compareStrings(a.destination, b.destination);
        case 'departure_date':
            return direction * compareDates(a.departure_date, b.departure_date);
        case 'return_date':
            return direction * compareDates(a.return_date, b.return_date);
        case 'price':
            return direction * compareNumbers(a.price, b.price);
        case 'currency':
            return direction * compareStrings(a.currency, b.currency);
        case 'airline_code':
            return direction * compareStrings(a.airline_code, b.airline_code);
        case 'duration':
            return direction * compareNumbers(a.duration, b.duration);
        default:
            return 0;
    }
}

function statusBadgeClass(status) {
    switch ((status || '').toLowerCase()) {
        case 'completed':
            return 'bg-success';
        case 'completed_with_errors':
            return 'bg-warning text-dark';
        case 'failed':
            return 'bg-danger';
        case 'running':
            return 'bg-primary';
        case 'queued':
        default:
            return 'bg-secondary';
    }
}

function formatDate(value) {
    if (!value) return '—';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }
    return date.toLocaleString();
}

function formatNumber(value) {
    if (value == null) return '—';
    const number = Number(value);
    if (Number.isNaN(number)) return value;
    return number.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function compareStrings(a, b) {
    return (a || '').localeCompare(b || '');
}

function compareNumbers(a, b) {
    const numA = Number(a) || 0;
    const numB = Number(b) || 0;
    return numA - numB;
}

function compareDates(a, b) {
    const dateA = a ? new Date(a).getTime() : 0;
    const dateB = b ? new Date(b).getTime() : 0;
    return dateA - dateB;
}
