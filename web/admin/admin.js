// Admin Panel JavaScript

// API endpoints - use relative paths to work with any host
const API_BASE = '/api/v1';
const ENDPOINTS = {
    JOBS: `${API_BASE}/admin/jobs`,
    WORKERS: `${API_BASE}/admin/workers`,
    QUEUE: `${API_BASE}/admin/queue`,
    AIRPORTS: `${API_BASE}/airports`,
    AIRLINES: `${API_BASE}/airlines`,
    PRICE_GRAPH_SWEEPS: `${API_BASE}/admin/price-graph-sweeps`
};

// DOM elements
const elements = {
    jobsCount: document.getElementById('jobsCount'),
    workersCount: document.getElementById('workersCount'),
    queueSize: document.getElementById('queueSize'),
    searchesCount: document.getElementById('searchesCount'),
    jobsTable: document.getElementById('jobsTable'),
    workersTable: document.getElementById('workersTable'),
    queueTable: document.getElementById('queueTable'),
    refreshBtn: document.getElementById('refreshBtn'),
    saveJobBtn: document.getElementById('saveJobBtn'),
    newJobForm: document.getElementById('newJobForm'),
    priceGraphForm: document.getElementById('priceGraphForm'),
    priceGraphTable: document.getElementById('priceGraphTable'),
    priceGraphResultsTable: document.getElementById('priceGraphResultsTable'),
    priceGraphResultsMeta: document.getElementById('priceGraphResultsMeta'),
    priceGraphEmpty: document.getElementById('priceGraphEmpty'),
    refreshSweepsBtn: document.getElementById('refreshSweepsBtn'),
    sweepsStatus: document.getElementById('sweepsStatus'),
    exportResultsBtn: document.getElementById('exportResultsBtn')
};

// State for price graph results
let currentSweepId = null;
let currentSweepResults = [];

// Initialize the admin panel
async function initAdminPanel() {
    // Add event listeners
    elements.refreshBtn.addEventListener('click', refreshData);
    elements.saveJobBtn.addEventListener('click', saveJob);
    if (elements.refreshSweepsBtn) {
        elements.refreshSweepsBtn.addEventListener('click', () => loadPriceGraphSweeps(true));
    }
    if (elements.priceGraphForm) {
        elements.priceGraphForm.addEventListener('submit', submitPriceGraphSweep);
    }
    if (elements.exportResultsBtn) {
        elements.exportResultsBtn.addEventListener('click', exportPriceGraphResults);
    }
    
    // Initialize Bootstrap modals
    const newJobModalEl = document.getElementById('newJobModal');
    if (newJobModalEl) {
        new bootstrap.Modal(newJobModalEl);
    }

    // Set default dates (30 days from now)
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 30);
    const dateString = futureDate.toISOString().split('T')[0];
    
    const dateStartEl = document.getElementById('dateStart');
    const dateEndEl = document.getElementById('dateEnd');
    if (dateStartEl) dateStartEl.value = dateString;
    if (dateEndEl) dateEndEl.value = dateString;

    const pgDepartureFrom = document.getElementById('pgDepartureFrom');
    const pgDepartureTo = document.getElementById('pgDepartureTo');
    if (pgDepartureFrom && pgDepartureTo) {
        const departStart = new Date();
        departStart.setDate(departStart.getDate() + 21);
        const departEnd = new Date(departStart);
        departEnd.setDate(departEnd.getDate() + 7);
        pgDepartureFrom.value = departStart.toISOString().split('T')[0];
        pgDepartureTo.value = departEnd.toISOString().split('T')[0];
    }

    // Initialize cron preview
    updateCronPreview();

    // Load initial data
    await refreshData();

    // Set up auto-refresh every 30 seconds
    setInterval(refreshData, 30000);
}

// Refresh all data
async function refreshData() {
    try {
        console.log('Refreshing admin panel data...');
        // Load data with error handling, don't fail entire refresh if one fails
        const loadPromises = [
            loadJobs().catch(err => {
                console.error('Error loading jobs:', err);
                elements.jobsCount.textContent = '?';
            }),
            loadWorkers().catch(err => {
                console.error('Error loading workers:', err);
                elements.workersCount.textContent = '?';
            }),
            loadQueueStatus().catch(err => {
                console.error('Error loading queue:', err);
                elements.queueSize.textContent = '?';
            }),
            loadSearchCounts().catch(err => {
                console.error('Error loading search counts:', err);
                elements.searchesCount.textContent = '?';
            }),
            loadPriceGraphSweeps().catch(err => {
                console.error('Error loading price graph sweeps:', err);
                if (elements.sweepsStatus) elements.sweepsStatus.textContent = 'Failed to load sweeps';
            })
        ];
        
        // Wait for all to complete (success or failure)
        await Promise.allSettled(loadPromises);
        console.log('Data refresh completed');
    } catch (error) {
        console.error('Error refreshing data:', error);
        showAlert('Some data could not be loaded. Check console for details.', 'warning');
    }
}

// Load jobs data
async function loadJobs() {
    const response = await fetch(ENDPOINTS.JOBS);
    if (!response.ok) throw new Error('Failed to load jobs');

    // Handle potential duplicate response format using safe parsing
    const responseText = await response.text();
    const jobs = safeParseJSON(responseText, []);
    
    // Ensure we have an array
    if (!Array.isArray(jobs)) {
        console.warn('Jobs response is not an array:', jobs);
        return;
    }

    // Update jobs count
    elements.jobsCount.textContent = jobs.length;

    // Clear table
    elements.jobsTable.innerHTML = '';

    // Add jobs to table
    jobs.forEach(job => {
        const row = document.createElement('tr');
        
        // Get job details safely
        const origin = job.details?.origin || job.origin || 'N/A';
        const destination = job.details?.destination || job.destination || 'N/A';
        const schedule = job.cron_expression || 'N/A';
        const isDynamic = job.details?.dynamic_dates || false;
        
        // Format route with dynamic indicator
        let routeDisplay = `${origin} → ${destination}`;
        if (isDynamic) {
            const daysFromExecution = job.details?.days_from_execution || 'N/A';
            const searchWindowDays = job.details?.search_window_days || 'N/A';
            routeDisplay += `<br><small class="text-info"><i class="bi bi-arrow-clockwise"></i> Dynamic: +${daysFromExecution}d (${searchWindowDays}d window)</small>`;
        }
        
        row.innerHTML = `
            <td>${job.id || 'N/A'}</td>
            <td>${job.name || 'Unnamed Job'}</td>
            <td><code>${schedule}</code></td>
            <td>${routeDisplay}</td>
            <td>
                <span class="badge ${job.enabled ? 'bg-success' : 'bg-secondary'}">
                    ${job.enabled ? 'Enabled' : 'Disabled'}
                </span>
                ${isDynamic ? '<br><span class="badge bg-info mt-1">Dynamic</span>' : ''}
            </td>
            <td>${job.last_run ? new Date(job.last_run).toLocaleString() : 'Never'}</td>
            <td>
                <div class="btn-group btn-group-sm">
                    <button class="btn btn-outline-primary" onclick="runJob(${job.id})" title="Run Job">
                        <i class="bi bi-play-fill"></i>
                    </button>
                    <button class="btn btn-outline-${job.enabled ? 'warning' : 'success'}"
                        onclick="${job.enabled ? 'disableJob' : 'enableJob'}(${job.id})" 
                        title="${job.enabled ? 'Disable' : 'Enable'} Job">
                        <i class="bi bi-${job.enabled ? 'pause-fill' : 'check-lg'}"></i>
                    </button>
                    <button class="btn btn-outline-danger" onclick="deleteJob(${job.id})" title="Delete Job">
                        <i class="bi bi-trash"></i>
                    </button>
                </div>
            </td>
        `;
        elements.jobsTable.appendChild(row);
    });
}

// Load workers data
async function loadWorkers() {
    const response = await fetch(ENDPOINTS.WORKERS);
    if (!response.ok) throw new Error('Failed to load workers');
    
    // Handle potential duplicate response format using safe parsing
    const responseText = await response.text();
    let workers = safeParseJSON(responseText, {});
    
    // Convert worker status object to workers array if needed
    if (!Array.isArray(workers)) {
        if (workers.status === 'running') {
            workers = Array.from({length: 5}, (_, i) => ({
                id: i + 1,
                status: 'active',
                current_job: null,
                processed_jobs: 0,
                uptime: 0
            }));
        } else {
            workers = [];
        }
    }
    
    // Update workers count
    const activeWorkers = workers.filter(w => w.status === 'active').length;
    elements.workersCount.textContent = activeWorkers;
    
    // Clear table
    elements.workersTable.innerHTML = '';
    
    // Add workers to table
    workers.forEach(worker => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${worker.id}</td>
            <td>
                <span class="badge ${worker.status === 'active' ? 'bg-success' : 'bg-secondary'}">
                    ${worker.status}
                </span>
            </td>
            <td>${worker.current_job || 'None'}</td>
            <td>${worker.processed_jobs}</td>
            <td>${formatDuration(worker.uptime)}</td>
        `;
        elements.workersTable.appendChild(row);
    });
}

// Load queue status
async function loadQueueStatus() {
    const response = await fetch(ENDPOINTS.QUEUE);
    if (!response.ok) throw new Error('Failed to load queue status');
    
    // Handle potential duplicate response format using safe parsing
    const responseText = await response.text();
    const queues = safeParseJSON(responseText, {});
    
    // Ensure we have an object
    if (typeof queues !== 'object' || Array.isArray(queues)) {
        console.warn('Queue response is not an object:', queues);
        return;
    }
    
    // Update queue size
    let totalPending = 0;
    Object.values(queues).forEach(q => totalPending += q.pending || 0);
    elements.queueSize.textContent = totalPending;
    
    // Clear table
    elements.queueTable.innerHTML = '';
    
    // Add queues to table
    Object.entries(queues).forEach(([name, stats]) => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${name}</td>
            <td>${stats.pending || 0}</td>
            <td>${stats.processing || 0}</td>
            <td>${stats.completed || 0}</td>
            <td>${stats.failed || 0}</td>
        `;
        elements.queueTable.appendChild(row);
    });
}

// Load search counts with retry logic
async function loadSearchCounts() {
    let retries = 3;
    while (retries > 0) {
        try {
            const response = await fetch(`${API_BASE}/search?limit=0`, {
                timeout: 10000,
                headers: {
                    'Accept': 'application/json'
                }
            });
            if (!response.ok) throw new Error(`HTTP ${response.status}: Failed to load search counts`);
            
            const data = await response.json();
            elements.searchesCount.textContent = data.total || 0;
            return; // Success, exit retry loop
        } catch (error) {
            console.error(`Error loading search counts (${4-retries}/3):`, error);
            retries--;
            if (retries > 0) {
                // Wait before retry
                await new Promise(resolve => setTimeout(resolve, 1000));
            } else {
                // Set fallback value on final failure
                elements.searchesCount.textContent = '?';
            }
        }
    }
}

// Submit a new price graph sweep request
async function submitPriceGraphSweep(event) {
    event.preventDefault();

    if (!elements.priceGraphForm) return;
    if (!elements.priceGraphForm.checkValidity()) {
        elements.priceGraphForm.reportValidity();
        return;
    }

    const submitBtn = document.getElementById('submitPriceGraphBtn');
    setButtonLoading(submitBtn, true);

    try {
        const origins = parseAirportList(document.getElementById('pgOrigins')?.value || '');
        const destinations = parseAirportList(document.getElementById('pgDestinations')?.value || '');

        if (!origins.length || !destinations.length) {
            throw new Error('At least one origin and destination are required');
        }

        const departureFrom = document.getElementById('pgDepartureFrom')?.value;
        const departureTo = document.getElementById('pgDepartureTo')?.value;
        if (!departureFrom || !departureTo) {
            throw new Error('Departure window is required');
        }

        const tripLengths = parseNumberList(document.getElementById('pgTripLengths')?.value || '');

        const payload = {
            origins,
            destinations,
            departure_date_from: `${departureFrom}T00:00:00Z`,
            departure_date_to: `${departureTo}T00:00:00Z`,
            trip_lengths: tripLengths.length ? tripLengths : undefined,
            trip_type: document.getElementById('pgTripType')?.value || 'round_trip',
            class: document.getElementById('pgClass')?.value || 'economy',
            stops: document.getElementById('pgStops')?.value || 'any',
            adults: parseInt(document.getElementById('pgAdults')?.value || '1', 10),
            children: parseInt(document.getElementById('pgChildren')?.value || '0', 10),
            infants_lap: parseInt(document.getElementById('pgInfantsLap')?.value || '0', 10),
            infants_seat: parseInt(document.getElementById('pgInfantsSeat')?.value || '0', 10),
            currency: (document.getElementById('pgCurrency')?.value || 'USD').trim().toUpperCase()
        };

        const rateLimitRaw = document.getElementById('pgRateLimit')?.value;
        const rateLimit = parseInt(rateLimitRaw || '', 10);
        if (!Number.isNaN(rateLimit) && rateLimit >= 0) {
            payload.rate_limit_millis = rateLimit;
        }

        // Remove undefined keys
        Object.keys(payload).forEach(key => {
            if (payload[key] === undefined || payload[key] === null) {
                delete payload[key];
            }
        });

        const response = await fetch(ENDPOINTS.PRICE_GRAPH_SWEEPS, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errorText = await response.text();
            const errorBody = safeParseJSON(errorText, {});
            throw new Error(errorBody.error || `Failed to enqueue sweep (HTTP ${response.status})`);
        }

        const dataText = await response.text();
        const data = safeParseJSON(dataText, {});
        const sweepId = data?.sweep_id;

        showAlert('Price graph sweep enqueued.', 'success');

        await loadPriceGraphSweeps(true);

        if (sweepId) {
            await viewPriceGraphResults(sweepId, true);
            if (elements.priceGraphResultsMeta) {
                elements.priceGraphResultsMeta.textContent = `Sweep #${sweepId} queued. Results will populate as the worker progresses.`;
            }
        }
    } catch (error) {
        console.error('Error enqueueing price graph sweep:', error);
        showAlert(`Error enqueueing sweep: ${error.message}`, 'danger');
    } finally {
        setButtonLoading(submitBtn, false);
    }
}

// Load recent price graph sweeps
async function loadPriceGraphSweeps(refreshResults = false) {
    if (!elements.priceGraphTable) return;

    if (elements.sweepsStatus) {
        elements.sweepsStatus.textContent = 'Loading...';
    }

    const response = await fetch(`${ENDPOINTS.PRICE_GRAPH_SWEEPS}?limit=50`);
    if (!response.ok) {
        throw new Error(`HTTP ${response.status}: failed to load price graph sweeps`);
    }

    const responseText = await response.text();
    const data = safeParseJSON(responseText, {});
    const sweeps = Array.isArray(data) ? data : Array.isArray(data.items) ? data.items : [];

    elements.priceGraphTable.innerHTML = '';

    if (!sweeps.length) {
        const emptyRow = document.createElement('tr');
        emptyRow.innerHTML = `<td colspan="7" class="text-center py-4 text-muted">No sweeps found.</td>`;
        elements.priceGraphTable.appendChild(emptyRow);
        if (elements.sweepsStatus) elements.sweepsStatus.textContent = '';
        return;
    }

    sweeps.forEach(sweep => {
        const row = document.createElement('tr');
        row.dataset.sweepId = sweep.id;

        const statusBadge = formatSweepStatusBadge(sweep.status);
        const coverage = `${sweep.origin_count || 0} × ${sweep.destination_count || 0}`;
        const tripWindow = formatTripLengthRange(sweep.trip_length_min, sweep.trip_length_max);
        const updatedAt = sweep.updated_at ? new Date(sweep.updated_at).toLocaleString() : '—';

        row.innerHTML = `
            <td>${sweep.id}</td>
            <td>${statusBadge}</td>
            <td>${coverage}</td>
            <td>${tripWindow}</td>
            <td>${sweep.error_count || 0}</td>
            <td>${updatedAt}</td>
            <td class="text-end">
                <div class="btn-group btn-group-sm">
                    <button class="btn btn-outline-primary" onclick="viewPriceGraphResults(${sweep.id})">
                        <i class="bi bi-eye"></i>
                    </button>
                </div>
            </td>
        `;

        elements.priceGraphTable.appendChild(row);
    });

    if (elements.sweepsStatus) {
        const timestamp = new Date().toLocaleTimeString();
        elements.sweepsStatus.textContent = `Last updated ${timestamp}`;
    }

    if (refreshResults && currentSweepId) {
        const stillExists = sweeps.some(sweep => sweep.id === currentSweepId);
        if (stillExists) {
            await viewPriceGraphResults(currentSweepId, true);
        }
    }
}

// View results for a specific sweep
async function viewPriceGraphResults(sweepId, silent = false) {
    currentSweepId = sweepId;

    if (elements.priceGraphResultsMeta && !silent) {
        elements.priceGraphResultsMeta.textContent = `Loading results for sweep #${sweepId}...`;
    }
    if (elements.exportResultsBtn) {
        elements.exportResultsBtn.disabled = true;
    }

    highlightSelectedSweepRow(sweepId);

    try {
        const response = await fetch(`${ENDPOINTS.PRICE_GRAPH_SWEEPS}/${sweepId}`);
        if (!response.ok) {
            const errorText = await response.text();
            const errorBody = safeParseJSON(errorText, {});
            throw new Error(errorBody.error || `Failed to load sweep results (HTTP ${response.status})`);
        }

        const responseText = await response.text();
        const data = safeParseJSON(responseText, {});
        const summary = data?.sweep || {};
        const results = Array.isArray(data?.results) ? data.results : [];

        currentSweepResults = results;

        renderPriceGraphResults(summary, results);

        if (elements.priceGraphResultsMeta) {
            const statusLabel = summary.status ? summary.status.replace(/_/g, ' ') : 'unknown';
            const currency = (summary.currency || (results[0]?.currency) || 'USD').toUpperCase();
            const message = `Sweep #${summary.id || sweepId} • ${results.length} fares • ${currency.toUpperCase()} • ${statusLabel}`;
            elements.priceGraphResultsMeta.textContent = message;
        }

        if (elements.exportResultsBtn) {
            elements.exportResultsBtn.disabled = results.length === 0;
        }

        if (!silent) {
            showAlert(`Loaded results for sweep #${sweepId}`, 'info');
        }
    } catch (error) {
        console.error(`Error loading results for sweep ${sweepId}:`, error);
        if (elements.priceGraphResultsMeta) {
            elements.priceGraphResultsMeta.textContent = `Failed to load sweep #${sweepId}: ${error.message}`;
        }
        showAlert(`Error loading sweep results: ${error.message}`, 'danger');
    }
}

// Render price graph result rows
function renderPriceGraphResults(summary, results) {
    if (!elements.priceGraphResultsTable) return;

    elements.priceGraphResultsTable.innerHTML = '';

    if (!results.length) {
        const emptyRow = document.createElement('tr');
        emptyRow.innerHTML = `<td colspan="5" class="text-center py-4 text-muted">No results stored for this sweep yet.</td>`;
        elements.priceGraphResultsTable.appendChild(emptyRow);
        return;
    }

    const formatter = buildCurrencyFormatter(summary?.currency || results[0]?.currency || 'USD');

    results.forEach(result => {
        const row = document.createElement('tr');
        const priceValue = Number(result.price);
        row.innerHTML = `
            <td><strong>${result.origin}</strong> → <strong>${result.destination}</strong></td>
            <td>${formatDate(result.departure_date)}</td>
            <td>${formatDate(result.return_date)}</td>
            <td>${result.trip_length ?? '—'}</td>
            <td class="text-end">${Number.isFinite(priceValue) ? formatter(priceValue) : '—'}</td>
        `;
        elements.priceGraphResultsTable.appendChild(row);
    });
}

// Export current sweep results to CSV
function exportPriceGraphResults() {
    if (!currentSweepResults.length || !currentSweepId) return;

    const header = ['sweep_id', 'origin', 'destination', 'departure_date', 'return_date', 'trip_length', 'price', 'currency', 'queried_at'];
    const rows = currentSweepResults.map(result => [
        result.sweep_id || currentSweepId,
        result.origin || '',
        result.destination || '',
        formatDateISO(result.departure_date),
        formatDateISO(result.return_date),
        result.trip_length ?? '',
        result.price ?? '',
        result.currency || '',
        formatDateISO(result.queried_at)
    ]);

    const csvContent = [header, ...rows]
        .map(line => line.map(escapeCsvValue).join(','))
        .join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);

    const link = document.createElement('a');
    link.href = url;
    link.download = `price-graph-sweep-${currentSweepId}.csv`;
    link.click();

    URL.revokeObjectURL(url);
}

// Save a new job
async function saveJob() {
    // Validate form
    if (!elements.newJobForm.checkValidity()) {
        elements.newJobForm.reportValidity();
        return;
    }

    // Get form values
    const time = document.getElementById('time').value;
    const interval = document.getElementById('interval').value;
    const [hour, minute] = time.split(':');
    
    // Generate cron expression based on interval and time
    let cronExpression;
    switch(interval) {
        case 'daily':
            cronExpression = `${minute} ${hour} * * *`;
            break;
        case 'weekly':
            cronExpression = `${minute} ${hour} * * 1`; // Monday
            break;
        case 'monthly':
            cronExpression = `${minute} ${hour} 1 * *`; // First of month
            break;
        default:
            cronExpression = `${minute} ${hour} * * *`; // Default to daily
    }

    const isDynamicDates = document.getElementById('dynamicDates').checked;

    const jobData = {
        name: document.getElementById('jobName').value,
        origin: document.getElementById('origin').value,
        destination: document.getElementById('destination').value,
        trip_type: document.getElementById('tripType').value,
        class: document.getElementById('class').value,
        stops: document.getElementById('stops').value,
        adults: parseInt(document.getElementById('adults').value),
        children: parseInt(document.getElementById('children').value),
        infants_lap: parseInt(document.getElementById('infantsLap').value),
        infants_seat: parseInt(document.getElementById('infantsSeat').value),
        currency: document.getElementById('currency').value,
        interval: interval,
        time: time,
        cron_expression: cronExpression,
        dynamic_dates: isDynamicDates
    };

    // Add date fields based on mode
    if (isDynamicDates) {
        // Dynamic date mode
        jobData.days_from_execution = parseInt(document.getElementById('daysFromExecution').value, 10) || 14;
        jobData.search_window_days = parseInt(document.getElementById('searchWindowDays').value, 10) || 7;
        jobData.trip_length = parseInt(document.getElementById('tripLength').value, 10) || 7;

        // Set placeholder static dates (required by backend but not used when dynamic)
        const futureDate = new Date();
        futureDate.setDate(futureDate.getDate() + 30);
        const dateString = futureDate.toISOString().split('T')[0];
        jobData.date_start = dateString;
        jobData.date_end = dateString;
    } else {
        // Static date mode
        jobData.date_start = document.getElementById('dateStart').value;
        jobData.date_end = document.getElementById('dateEnd').value;

        const returnDateStart = document.getElementById('returnDateStart').value;
        const returnDateEnd = document.getElementById('returnDateEnd').value;
        if (returnDateStart) {
            jobData.return_date_start = returnDateStart;
        }
        if (returnDateEnd) {
            jobData.return_date_end = returnDateEnd;
        }
    }

    // Remove dynamic-only fields when not in dynamic mode
    if (!isDynamicDates) {
        delete jobData.days_from_execution;
        delete jobData.search_window_days;
        delete jobData.trip_length;
    }

    try {
        const response = await fetch(ENDPOINTS.JOBS, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(jobData)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to create job');
        }

        // Close modal and refresh data
        const modal = bootstrap.Modal.getInstance(document.getElementById('newJobModal'));
        modal.hide();

        // Reset form
        elements.newJobForm.reset();

        // Show success message
        showAlert('Job created successfully!', 'success');

        // Refresh all data to update jobs, queue status, and workers
        await refreshData();
    } catch (error) {
        console.error('Error creating job:', error);
        showAlert(`Error creating job: ${error.message}`, 'danger');
    }
}

// Run a job
async function runJob(jobId) {
    try {
        const response = await fetch(`${ENDPOINTS.JOBS}/${jobId}/run`, {
            method: 'POST'
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to run job');
        }
        
        showAlert('Job started successfully!', 'success');
        await refreshData();
    } catch (error) {
        console.error('Error running job:', error);
        showAlert(`Error running job: ${error.message}`, 'danger');
    }
}

// Enable a job
async function enableJob(jobId) {
    try {
        const response = await fetch(`${ENDPOINTS.JOBS}/${jobId}/enable`, {
            method: 'POST'
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to enable job');
        }
        
        showAlert('Job enabled successfully!', 'success');
        await loadJobs();
    } catch (error) {
        console.error('Error enabling job:', error);
        showAlert(`Error enabling job: ${error.message}`, 'danger');
    }
}

// Disable a job
async function disableJob(jobId) {
    try {
        const response = await fetch(`${ENDPOINTS.JOBS}/${jobId}/disable`, {
            method: 'POST'
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to disable job');
        }
        
        showAlert('Job disabled successfully!', 'success');
        await loadJobs();
    } catch (error) {
        console.error('Error disabling job:', error);
        showAlert(`Error disabling job: ${error.message}`, 'danger');
    }
}

// Delete a job
async function deleteJob(jobId) {
    if (!confirm('Are you sure you want to delete this job?')) {
        return;
    }
    
    try {
        const response = await fetch(`${ENDPOINTS.JOBS}/${jobId}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to delete job');
        }
        
        showAlert('Job deleted successfully!', 'success');
        await loadJobs();
    } catch (error) {
        console.error('Error deleting job:', error);
        showAlert(`Error deleting job: ${error.message}`, 'danger');
    }
}

// Helper function to format duration
function formatDuration(seconds) {
    if (!seconds) return 'N/A';
    
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    
    return `${hours}h ${minutes}m ${secs}s`;
}

// Show an alert message
function showAlert(message, type = 'info') {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
    alertDiv.role = 'alert';

    const messageSpan = document.createElement('span');
    messageSpan.textContent = message;
    alertDiv.appendChild(messageSpan);

    const closeBtn = document.createElement('button');
    closeBtn.type = 'button';
    closeBtn.className = 'btn-close';
    closeBtn.setAttribute('data-bs-dismiss', 'alert');
    closeBtn.setAttribute('aria-label', 'Close');
    alertDiv.appendChild(closeBtn);
    
    // Insert at the top of the main content
    const main = document.querySelector('main');
    main.insertBefore(alertDiv, main.firstChild);
    
    // Auto-dismiss after 5 seconds
    setTimeout(() => {
        alertDiv.classList.remove('show');
        setTimeout(() => alertDiv.remove(), 150);
    }, 5000);
}

// Update cron expression preview
function updateCronPreview() {
    const timeInput = document.getElementById('time');
    const intervalInput = document.getElementById('interval');
    const cronPreview = document.getElementById('cronPreview');
    
    if (!timeInput || !intervalInput || !cronPreview) return;
    
    const time = timeInput.value;
    const interval = intervalInput.value;
    
    if (!time) return;
    
    const [hour, minute] = time.split(':');
    
    let cronExpression;
    switch(interval) {
        case 'daily':
            cronExpression = `${minute} ${hour} * * *`;
            break;
        case 'weekly':
            cronExpression = `${minute} ${hour} * * 1`; // Monday
            break;
        case 'monthly':
            cronExpression = `${minute} ${hour} 1 * *`; // First of month
            break;
        default:
            cronExpression = `${minute} ${hour} * * *`;
    }
    
    cronPreview.textContent = cronExpression;
}

// Toggle between static and dynamic date modes
function toggleDateMode() {
    const dynamicDatesCheckbox = document.getElementById('dynamicDates');
    const staticDateFields = document.getElementById('staticDateFields');
    const dynamicDateFields = document.getElementById('dynamicDateFields');
    const dateStartInput = document.getElementById('dateStart');
    const dateEndInput = document.getElementById('dateEnd');
    const daysFromExecutionInput = document.getElementById('daysFromExecution');
    const searchWindowDaysInput = document.getElementById('searchWindowDays');
    
    if (!dynamicDatesCheckbox || !staticDateFields || !dynamicDateFields) return;
    
    const isDynamic = dynamicDatesCheckbox.checked;
    
    if (isDynamic) {
        // Show dynamic fields, hide static fields
        staticDateFields.style.display = 'none';
        dynamicDateFields.style.display = 'block';
        
        // Remove required attribute from static date fields
        if (dateStartInput) dateStartInput.removeAttribute('required');
        if (dateEndInput) dateEndInput.removeAttribute('required');
        
        // Add required attribute to dynamic fields
        if (daysFromExecutionInput) daysFromExecutionInput.setAttribute('required', 'required');
        if (searchWindowDaysInput) searchWindowDaysInput.setAttribute('required', 'required');
    } else {
        // Show static fields, hide dynamic fields
        staticDateFields.style.display = 'block';
        dynamicDateFields.style.display = 'none';
        
        // Add required attribute to static date fields
        if (dateStartInput) dateStartInput.setAttribute('required', 'required');
        if (dateEndInput) dateEndInput.setAttribute('required', 'required');
        
        // Remove required attribute from dynamic fields
        if (daysFromExecutionInput) daysFromExecutionInput.removeAttribute('required');
        if (searchWindowDaysInput) searchWindowDaysInput.removeAttribute('required');
    }
}

function setButtonLoading(button, isLoading) {
    if (!button) return;
    if (isLoading) {
        if (!button.dataset.originalHtml) {
            button.dataset.originalHtml = button.innerHTML;
        }
        button.disabled = true;
        button.innerHTML = `<span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>Processing...`;
    } else {
        button.disabled = false;
        if (button.dataset.originalHtml) {
            button.innerHTML = button.dataset.originalHtml;
        }
    }
}

function parseAirportList(input) {
    return input
        .split(',')
        .map(code => code.trim().toUpperCase())
        .filter(code => code.length > 0);
}

function parseNumberList(input) {
    return input
        .split(',')
        .map(value => parseInt(value.trim(), 10))
        .filter(value => !Number.isNaN(value) && Number.isFinite(value));
}

function formatTripLengthRange(minVal, maxVal) {
    if (minVal == null && maxVal == null) return '—';
    if (minVal == null) return `${maxVal}`;
    if (maxVal == null) return `${minVal}`;
    if (minVal === maxVal) return `${minVal}`;
    return `${minVal} - ${maxVal}`;
}

function formatDate(value) {
    if (!value) return '—';
    try {
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) {
            return '—';
        }
        return date.toLocaleDateString();
    } catch (err) {
        return '—';
    }
}

function formatDateISO(value) {
    if (!value) return '';
    try {
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) {
            return '';
        }
        return date.toISOString();
    } catch (err) {
        return '';
    }
}

function formatSweepStatusBadge(status) {
    const normalized = (status || 'unknown').toLowerCase();
    const badgeClasses = {
        queued: 'bg-secondary',
        running: 'bg-info text-dark',
        completed: 'bg-success',
        completed_with_errors: 'bg-warning text-dark',
        failed: 'bg-danger',
        unknown: 'bg-secondary'
    };
    const cssClass = badgeClasses[normalized] || 'bg-secondary';
    const label = normalized.replace(/_/g, ' ');
    return `<span class="badge ${cssClass} text-uppercase">${label}</span>`;
}

function highlightSelectedSweepRow(sweepId) {
    if (!elements.priceGraphTable) return;
    const rows = elements.priceGraphTable.querySelectorAll('tr');
    rows.forEach(row => {
        if (row.dataset && row.dataset.sweepId) {
            row.classList.toggle('table-active', row.dataset.sweepId === String(sweepId));
        } else {
            row.classList.remove('table-active');
        }
    });
}

function buildCurrencyFormatter(currencyCode) {
    try {
        const formatter = new Intl.NumberFormat(undefined, {
            style: 'currency',
            currency: (currencyCode || 'USD').toUpperCase(),
            maximumFractionDigits: 2
        });
        return value => formatter.format(value);
    } catch (err) {
        return value => value.toFixed(2);
    }
}

function escapeCsvValue(value) {
    if (value == null) return '';
    const stringValue = String(value);
    if (/[",\n]/.test(stringValue)) {
        return `"${stringValue.replace(/"/g, '""')}"`;
    }
    return stringValue;
}

// Utility function to safely parse JSON responses that might be malformed
function safeParseJSON(responseText, fallbackValue = null) {
    try {
        // First, try normal JSON parsing
        return JSON.parse(responseText);
    } catch (e) {
        console.warn('Initial JSON parse failed, trying fallback methods:', e.message);
        
        // Try to handle duplicate JSON objects like {"a":1}{"b":2}
        try {
            const jsonObjects = responseText.match(/\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}/g);
            if (jsonObjects && jsonObjects.length > 0) {
                // Use the last valid JSON object
                const lastObject = jsonObjects[jsonObjects.length - 1];
                return JSON.parse(lastObject);
            }
        } catch (e2) {
            console.warn('Duplicate object parsing failed:', e2.message);
        }
        
        // Try to handle array format like [][{data}]
        try {
            const arrayMatch = responseText.match(/\]\[(.+)\]$/);
            if (arrayMatch) {
                return JSON.parse('[' + arrayMatch[1] + ']');
            }
        } catch (e3) {
            console.warn('Array format parsing failed:', e3.message);
        }
        
        // Try to extract any valid JSON from the response
        try {
            const jsonMatch = responseText.match(/(\[.*\]|\{.*\})/);
            if (jsonMatch) {
                return JSON.parse(jsonMatch[1]);
            }
        } catch (e4) {
            console.warn('Pattern matching failed:', e4.message);
        }
        
        console.error('All JSON parsing methods failed for response:', responseText);
        return fallbackValue;
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', initAdminPanel);
