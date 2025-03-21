// Admin Panel JavaScript

// API endpoints
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

    // Load initial data
    await refreshData();

    // Set up auto-refresh every 30 seconds
    setInterval(refreshData, 30000);
}

// Refresh all data
async function refreshData() {
    try {
        await Promise.all([
            loadJobs(),
            loadWorkers(),
            loadQueueStatus(),
            loadSearchCounts()
        ]);
    } catch (error) {
        console.error('Error refreshing data:', error);
        showAlert('Error refreshing data. Please try again.', 'danger');
    }
}

// Load jobs data
async function loadJobs() {
    const response = await fetch(ENDPOINTS.JOBS);
    if (!response.ok) throw new Error('Failed to load jobs');

    const jobs = await response.json();

    // Update jobs count
    elements.jobsCount.textContent = jobs.length;

    // Clear table
    elements.jobsTable.innerHTML = '';

    // Add jobs to table
    jobs.forEach(job => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${job.id}</td>
            <td>${job.name}</td>
            <td>${job.friendly_schedule}</td>
            <td>${job.origin} → ${job.destination}</td>
            <td>
                <span class="badge ${job.enabled ? 'bg-success' : 'bg-secondary'}">
                    ${job.enabled ? 'Enabled' : 'Disabled'}
                </span>
            </td>
            <td>${job.last_run ? new Date(job.last_run).toLocaleString() : 'Never'}</td>
            <td>
                <div class="btn-group btn-group-sm">
                    <button class="btn btn-outline-primary" onclick="runJob(${job.id})">
                        <i class="bi bi-play-fill"></i>
                    </button>
                    <button class="btn btn-outline-${job.enabled ? 'warning' : 'success'}"
                        onclick="${job.enabled ? 'disableJob' : 'enableJob'}(${job.id})">
                        <i class="bi bi-${job.enabled ? 'pause-fill' : 'check-lg'}"></i>
                    </button>
                    <button class="btn btn-outline-danger" onclick="deleteJob(${job.id})">
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
    
    const workers = await response.json();
    
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
    
    const queues = await response.json();
    
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

// Load search counts
async function loadSearchCounts() {
    try {
        const response = await fetch(`${API_BASE}/search?limit=0`);
        if (!response.ok) throw new Error('Failed to load search counts');
        
        const data = await response.json();
        elements.searchesCount.textContent = data.total || 0;
    } catch (error) {
        console.error('Error loading search counts:', error);
        // Don't show an alert for this non-critical error
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
    const jobData = {
        name: document.getElementById('jobName').value,
        origin: document.getElementById('origin').value,
        destination: document.getElementById('destination').value,
        date_start: document.getElementById('dateStart').value,
        date_end: document.getElementById('dateEnd').value,
        return_date_start: document.getElementById('returnDateStart').value || null,
        return_date_end: document.getElementById('returnDateEnd').value || null,
        trip_type: document.getElementById('tripType').value,
        class: document.getElementById('class').value,
        stops: document.getElementById('stops').value,
        adults: parseInt(document.getElementById('adults').value),
        children: parseInt(document.getElementById('children').value),
        infants_lap: parseInt(document.getElementById('infantsLap').value),
        infants_seat: parseInt(document.getElementById('infantsSeat').value),
        currency: document.getElementById('currency').value,
        interval: document.getElementById('interval').value,
        time: document.getElementById('time').value,
        friendly_schedule: ''
    };

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
    alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    `;
    
    // Insert at the top of the main content
    const main = document.querySelector('main');
    main.insertBefore(alertDiv, main.firstChild);
    
    // Auto-dismiss after 5 seconds
    setTimeout(() => {
        alertDiv.classList.remove('show');
        setTimeout(() => alertDiv.remove(), 150);
    }, 5000);
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', initAdminPanel);
