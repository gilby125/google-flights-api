// Admin Panel JavaScript

// API endpoints - use relative paths to work with any host
const API_BASE = '/api/v1';
const ENDPOINTS = {
    JOBS: `${API_BASE}/admin/jobs`,
    WORKERS: `${API_BASE}/admin/workers`,
    QUEUE: `${API_BASE}/admin/queue`,
    AIRPORTS: `${API_BASE}/airports`,
    AIRLINES: `${API_BASE}/airlines`
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
    newJobForm: document.getElementById('newJobForm')
};

// Initialize the admin panel
async function initAdminPanel() {
    // Add event listeners
    elements.refreshBtn.addEventListener('click', refreshData);
    elements.saveJobBtn.addEventListener('click', saveJob);
    
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
