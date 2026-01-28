// Admin Panel JavaScript

// HTML escaping utility to prevent XSS
function escapeHtml(unsafe) {
  if (unsafe == null) return "";
  return String(unsafe)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// API endpoints - use relative paths to work with any host
const API_BASE = "/api/v1";
const ENDPOINTS = {
  JOBS: `${API_BASE}/admin/jobs`,
  WORKERS: `${API_BASE}/admin/workers`,
  QUEUE: `${API_BASE}/admin/queue`,
  AIRPORTS: `${API_BASE}/airports`,
  AIRLINES: `${API_BASE}/airlines`,
  REGIONS: `${API_BASE}/regions`,
  AIRLINE_GROUPS: `${API_BASE}/airline-groups`,
  PRICE_GRAPH_SWEEPS: `${API_BASE}/admin/price-graph-sweeps`,
  CONTINUOUS_SWEEP: `${API_BASE}/admin/continuous-sweep`,
};

// DOM elements
const elements = {
  jobsCount: document.getElementById("jobsCount"),
  workersCount: document.getElementById("workersCount"),
  queueSize: document.getElementById("queueSize"),
  searchesCount: document.getElementById("searchesCount"),
  jobsTable: document.getElementById("jobsTable"),
  workersTable: document.getElementById("workersTable"),
  queueTable: document.getElementById("queueTable"),
  refreshBtn: document.getElementById("refreshBtn"),
  saveJobBtn: document.getElementById("saveJobBtn"),
  newJobForm: document.getElementById("newJobForm"),
  priceGraphForm: document.getElementById("priceGraphForm"),
  priceGraphTable: document.getElementById("priceGraphTable"),
  priceGraphResultsTable: document.getElementById("priceGraphResultsTable"),
  priceGraphResultsMeta: document.getElementById("priceGraphResultsMeta"),
  priceGraphEmpty: document.getElementById("priceGraphEmpty"),
  refreshSweepsBtn: document.getElementById("refreshSweepsBtn"),
  sweepsStatus: document.getElementById("sweepsStatus"),
  exportResultsBtn: document.getElementById("exportResultsBtn"),
};

// State for price graph results
let currentSweepId = null;
let currentSweepResults = [];

// Initialize the admin panel
async function initAdminPanel() {
  // Add event listeners
  elements.refreshBtn.addEventListener("click", refreshData);
  elements.saveJobBtn.addEventListener("click", saveJob);
  if (elements.queueTable) {
    elements.queueTable.addEventListener("click", (event) => {
      const button = event.target.closest("button[data-action]");
      if (!button) return;

      const action = button.getAttribute("data-action");
      const queueName = button.getAttribute("data-queue");
      const pending =
        parseInt(button.getAttribute("data-pending") || "0", 10) || 0;
      const processing =
        parseInt(button.getAttribute("data-processing") || "0", 10) || 0;
      const failed =
        parseInt(button.getAttribute("data-failed") || "0", 10) || 0;

      const handler = (() => {
        switch (action) {
          case "clear-queue":
            return () => clearQueue(queueName, pending);
          case "drain-queue":
            return () => drainQueue(queueName, pending, processing);
          case "clear-processing":
            return () => clearProcessing(queueName, processing);
          case "clear-failed":
            return () => clearFailed(queueName, failed);
          case "retry-failed":
            return () => retryFailed(queueName, failed);
          default:
            return null;
        }
      })();

      if (!handler) return;
      handler().catch((err) => {
        console.error("Queue action error:", err);
        showAlert(`Queue action error: ${err.message}`, "danger");
      });
    });
  }

  if (elements.workersTable) {
    elements.workersTable.addEventListener("click", (event) => {
      const button = event.target.closest("button[data-action]");
      if (!button) return;
      const action = button.getAttribute("data-action");
      if (action !== "cancel-queue-job") return;

      const queueName = button.getAttribute("data-queue");
      const jobID = button.getAttribute("data-job");
      cancelQueueJob(queueName, jobID).catch((err) => {
        console.error("Cancel job error:", err);
        showAlert(`Cancel job error: ${err.message}`, "danger");
      });
    });
  }
  if (elements.refreshSweepsBtn) {
    elements.refreshSweepsBtn.addEventListener("click", () =>
      loadPriceGraphSweeps(true),
    );
  }
  if (elements.priceGraphForm) {
    elements.priceGraphForm.addEventListener("submit", submitPriceGraphSweep);
  }
  if (elements.exportResultsBtn) {
    elements.exportResultsBtn.addEventListener(
      "click",
      exportPriceGraphResults,
    );
  }

  // Initialize Bootstrap modals
  const newJobModalEl = document.getElementById("newJobModal");
  if (newJobModalEl) {
    new bootstrap.Modal(newJobModalEl);
  }

  // Set default dates (30 days from now)
  const futureDate = new Date();
  futureDate.setDate(futureDate.getDate() + 30);
  const dateString = futureDate.toISOString().split("T")[0];

  const dateStartEl = document.getElementById("dateStart");
  const dateEndEl = document.getElementById("dateEnd");
  if (dateStartEl) dateStartEl.value = dateString;
  if (dateEndEl) dateEndEl.value = dateString;

  const pgDepartureFrom = document.getElementById("pgDepartureFrom");
  const pgDepartureTo = document.getElementById("pgDepartureTo");
  if (pgDepartureFrom && pgDepartureTo) {
    const departStart = new Date();
    departStart.setDate(departStart.getDate() + 21);
    const departEnd = new Date(departStart);
    departEnd.setDate(departEnd.getDate() + 7);
    pgDepartureFrom.value = departStart.toISOString().split("T")[0];
    pgDepartureTo.value = departEnd.toISOString().split("T")[0];
  }

  // Initialize cron preview
  updateCronPreview();

  // Initialize macro token helpers (regions / airline-groups)
  initMacroTokenHelpers();

  // Load initial data
  await refreshData();

  // Set up real-time updates via Server-Sent Events for workers
  initEventSource();

  // Keep periodic refresh for non-worker data (jobs, queue, sweeps)
  setInterval(refreshNonWorkerData, 30000);
}

// Refresh non-worker data (jobs, queue, counts, sweeps) - workers are updated via SSE
async function refreshNonWorkerData() {
  try {
    await Promise.allSettled([
      loadJobs().catch((err) => {
        console.error("Error loading jobs:", err);
        if (elements.jobsCount) elements.jobsCount.textContent = "?";
      }),
      loadQueueStatus().catch((err) => {
        console.error("Error loading queue:", err);
        if (elements.queueSize) elements.queueSize.textContent = "?";
      }),
      loadSearchCounts().catch((err) => {
        console.error("Error loading search counts:", err);
        if (elements.searchesCount) elements.searchesCount.textContent = "?";
      }),
      loadPriceGraphSweeps().catch((err) => {
        console.error("Error loading price graph sweeps:", err);
        if (elements.sweepsStatus)
          elements.sweepsStatus.textContent = "Failed to load sweeps";
      }),
    ]);
  } catch (error) {
    console.error("Error refreshing non-worker data:", error);
  }
}

// Refresh all data
async function refreshData() {
  try {
    console.log("Refreshing admin panel data...");
    // Load data with error handling, don't fail entire refresh if one fails
    const loadPromises = [
      loadJobs().catch((err) => {
        console.error("Error loading jobs:", err);
        elements.jobsCount.textContent = "?";
      }),
      loadWorkers().catch((err) => {
        console.error("Error loading workers:", err);
        elements.workersCount.textContent = "?";
      }),
      loadQueueStatus().catch((err) => {
        console.error("Error loading queue:", err);
        elements.queueSize.textContent = "?";
      }),
      loadSearchCounts().catch((err) => {
        console.error("Error loading search counts:", err);
        elements.searchesCount.textContent = "?";
      }),
      loadPriceGraphSweeps().catch((err) => {
        console.error("Error loading price graph sweeps:", err);
        if (elements.sweepsStatus)
          elements.sweepsStatus.textContent = "Failed to load sweeps";
      }),
    ];

    // Wait for all to complete (success or failure)
    await Promise.allSettled(loadPromises);
    console.log("Data refresh completed");
  } catch (error) {
    console.error("Error refreshing data:", error);
    showAlert(
      "Some data could not be loaded. Check console for details.",
      "warning",
    );
  }
}

// Initialize Server-Sent Events for real-time updates
function initEventSource() {
  const eventSource = new EventSource(`${API_BASE}/admin/events`);

  eventSource.addEventListener("worker-status", (e) => {
    try {
      const workers = JSON.parse(e.data);
      updateWorkersUI(workers);
    } catch (error) {
      console.error("Error parsing worker-status event:", error);
    }
  });

  eventSource.onerror = (error) => {
    console.warn("SSE connection error, will auto-reconnect:", error);
    // EventSource automatically reconnects
  };

  eventSource.onopen = () => {
    console.log("SSE connection established");
  };

  // Store reference for cleanup if needed
  window.adminEventSource = eventSource;
}

// Update workers UI with new data (extracted from loadWorkers)
function updateWorkersUI(workers) {
  if (!elements.workersTable) return;

  // Convert worker status object to workers array if needed
  if (!Array.isArray(workers)) {
    if (workers.status === "running") {
      workers = Array.from({ length: 5 }, (_, i) => ({
        id: i + 1,
        status: "active",
        current_job: null,
        processed_jobs: 0,
        uptime: 0,
      }));
    } else {
      workers = [];
    }
  }

  // Update workers count (treat processing as active)
  const activeWorkers = workers.filter(
    (w) => w.status === "active" || w.status === "processing",
  ).length;
  if (elements.workersCount) {
    elements.workersCount.textContent = activeWorkers;
  }

  // Clear table
  elements.workersTable.innerHTML = "";

  // Add workers to table
  workers.forEach((worker) => {
    const row = document.createElement("tr");
    const workerId = Number.isInteger(worker.id)
      ? worker.id
      : escapeHtml(worker.id);
    const workerStatus = escapeHtml(worker.status);
    const currentJobRaw = worker.current_job || "";
    const currentJob = escapeHtml(currentJobRaw || "None");
    const processedJobs = Number.isInteger(worker.processed_jobs)
      ? worker.processed_jobs
      : 0;

    const source = escapeHtml(worker.source || "");
    const hostname = escapeHtml(worker.hostname || "");
    const concurrency = Number.isInteger(worker.concurrency)
      ? worker.concurrency
      : null;
    const heartbeatAgeSeconds = Number.isInteger(worker.heartbeat_age_seconds)
      ? worker.heartbeat_age_seconds
      : null;

    let idCell = `${workerId}`;
    if (source === "remote") {
      idCell += ' <span class="badge bg-info ms-1">remote</span>';
    }
    const metaParts = [];
    if (hostname) metaParts.push(hostname);
    if (concurrency != null) metaParts.push(`${concurrency}x`);
    if (heartbeatAgeSeconds != null)
      metaParts.push(`${heartbeatAgeSeconds}s ago`);
    if (metaParts.length) {
      idCell += `<br><small class="text-muted">${metaParts.join(" • ")}</small>`;
    }

    let cancelButton = "";
    if (currentJobRaw && String(worker.status) === "processing") {
      const parts = String(currentJobRaw).split(":");
      if (parts.length >= 2) {
        const queueName = parts[0];
        const jobID = parts.slice(1).join(":");
        cancelButton = `
          <button
            type="button"
            class="btn btn-sm btn-outline-danger"
            data-action="cancel-queue-job"
            data-queue="${escapeHtml(queueName)}"
            data-job="${escapeHtml(jobID)}"
            title="Request cancellation (best-effort)">
            <i class="bi bi-stop-fill me-1"></i>Cancel
          </button>
        `;
      }
    }
    row.innerHTML = `
            <td>${idCell}</td>
            <td>
                <span class="badge ${worker.status === "active" || worker.status === "processing" ? "bg-success" : "bg-secondary"}">
                    ${workerStatus}
                </span>
            </td>
            <td>${currentJob}</td>
            <td class="text-end">${cancelButton}</td>
            <td>${processedJobs}</td>
            <td>${formatDuration(worker.uptime)}</td>
        `;
    elements.workersTable.appendChild(row);
  });
}

// Load jobs data
async function loadJobs() {
  const response = await fetch(ENDPOINTS.JOBS);
  if (!response.ok) throw new Error("Failed to load jobs");

  // Handle potential duplicate response format using safe parsing
  const responseText = await response.text();
  const jobs = safeParseJSON(responseText, []);

  // Ensure we have an array
  if (!Array.isArray(jobs)) {
    console.warn("Jobs response is not an array:", jobs);
    return;
  }

  // Update jobs count
  elements.jobsCount.textContent = jobs.length;

  // Clear table
  elements.jobsTable.innerHTML = "";

  // Add jobs to table
  jobs.forEach((job) => {
    const row = document.createElement("tr");

    // Get job details safely and escape for HTML
    const origin = escapeHtml(job.details?.origin || job.origin || "N/A");
    const destination = escapeHtml(
      job.details?.destination || job.destination || "N/A",
    );
    const schedule = escapeHtml(job.cron_expression || "N/A");
    const jobName = escapeHtml(job.name || "Unnamed Job");
    const isDynamic = job.details?.dynamic_dates || false;

    // Format route with dynamic indicator
    let routeDisplay = `${origin} → ${destination}`;
    if (isDynamic) {
      const daysFromExecution = escapeHtml(
        job.details?.days_from_execution || "N/A",
      );
      const searchWindowDays = escapeHtml(
        job.details?.search_window_days || "N/A",
      );
      routeDisplay += `<br><small class="text-info"><i class="bi bi-arrow-clockwise"></i> Dynamic: +${daysFromExecution}d (${searchWindowDays}d window)</small>`;
    }

    // Validate job.id is a number to prevent injection in onclick handlers
    const jobId = Number.isInteger(job.id) ? job.id : 0;

    row.innerHTML = `
            <td>${jobId || "N/A"}</td>
            <td>${jobName}</td>
            <td><code>${schedule}</code></td>
            <td>${routeDisplay}</td>
            <td>
                <span class="badge ${job.enabled ? "bg-success" : "bg-secondary"}">
                    ${job.enabled ? "Enabled" : "Disabled"}
                </span>
                ${isDynamic ? '<br><span class="badge bg-info mt-1">Dynamic</span>' : ""}
            </td>
            <td>${job.last_run ? new Date(job.last_run).toLocaleString() : "Never"}</td>
            <td>
                <div class="btn-group btn-group-sm">
                    <button class="btn btn-outline-primary" onclick="runJob(${jobId})" title="Run Job">
                        <i class="bi bi-play-fill"></i>
                    </button>
                    <button class="btn btn-outline-${job.enabled ? "warning" : "success"}"
                        onclick="${job.enabled ? "disableJob" : "enableJob"}(${jobId})"
                        title="${job.enabled ? "Disable" : "Enable"} Job">
                        <i class="bi bi-${job.enabled ? "pause-fill" : "check-lg"}"></i>
                    </button>
                    <button class="btn btn-outline-danger" onclick="deleteJob(${jobId})" title="Delete Job">
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
  if (!response.ok) throw new Error("Failed to load workers");

  const responseText = await response.text();
  const workers = safeParseJSON(responseText, []);
  updateWorkersUI(workers);
}

// Load queue status
async function loadQueueStatus() {
  const response = await fetch(ENDPOINTS.QUEUE);
  if (!response.ok) throw new Error("Failed to load queue status");

  // Handle potential duplicate response format using safe parsing
  const responseText = await response.text();
  const queues = safeParseJSON(responseText, {});

  // Ensure we have an object
  if (typeof queues !== "object" || Array.isArray(queues)) {
    console.warn("Queue response is not an object:", queues);
    return;
  }

  // Update queue size
  let totalPending = 0;
  Object.values(queues).forEach((q) => (totalPending += q.pending || 0));
  elements.queueSize.textContent = totalPending;

  // Clear table
  elements.queueTable.innerHTML = "";

  // Add queues to table
  Object.entries(queues).forEach(([name, stats]) => {
    const row = document.createElement("tr");
    const queueName = escapeHtml(name);
    const pending = Number.isInteger(stats.pending) ? stats.pending : 0;
    const processing = Number.isInteger(stats.processing)
      ? stats.processing
      : 0;
    const completed = Number.isInteger(stats.completed) ? stats.completed : 0;
    const failed = Number.isInteger(stats.failed) ? stats.failed : 0;

    const clearDisabled = pending === 0;
    const clearProcessingDisabled = processing === 0;
    const clearFailedDisabled = failed === 0;
    const retryFailedDisabled = failed === 0;
    const drainDisabled = pending === 0 && processing === 0;
    row.innerHTML = `
            <td>${queueName}</td>
            <td>${pending}</td>
            <td>${processing}</td>
            <td>${completed}</td>
            <td>${failed}</td>
            <td class="text-end">
                <button
                    type="button"
                    class="btn btn-sm btn-outline-danger me-1"
                    data-action="drain-queue"
                    data-queue="${queueName}"
                    data-pending="${pending}"
                    data-processing="${processing}"
                    ${drainDisabled ? "disabled" : ""}
                    title="${drainDisabled ? "Nothing to drain" : "Cancel in-flight jobs and clear pending jobs"}">
                    <i class="bi bi-stop-fill me-1"></i>Drain
                </button>
                <button
                    type="button"
                    class="btn btn-sm btn-outline-danger me-1"
                    data-action="clear-queue"
                    data-queue="${queueName}"
                    data-pending="${pending}"
                    ${clearDisabled ? "disabled" : ""}
                    title="${clearDisabled ? "No pending jobs to clear" : "Clear pending jobs"}">
                    <i class="bi bi-trash me-1"></i>Clear
                </button>
                <button
                    type="button"
                    class="btn btn-sm btn-outline-danger me-1"
                    data-action="clear-processing"
                    data-queue="${queueName}"
                    data-processing="${processing}"
                    ${clearProcessingDisabled ? "disabled" : ""}
                    title="${clearProcessingDisabled ? "No processing jobs to clear" : "Clear processing jobs (force)"}">
                    <i class="bi bi-x-octagon me-1"></i>Clear Processing
                </button>
                <button
                    type="button"
                    class="btn btn-sm btn-outline-danger me-1"
                    data-action="clear-failed"
                    data-queue="${queueName}"
                    data-failed="${failed}"
                    ${clearFailedDisabled ? "disabled" : ""}
                    title="${clearFailedDisabled ? "No failed jobs to clear" : "Clear failed jobs"}">
                    <i class="bi bi-trash3 me-1"></i>Clear Failed
                </button>
                <button
                    type="button"
                    class="btn btn-sm btn-outline-primary"
                    data-action="retry-failed"
                    data-queue="${queueName}"
                    data-failed="${failed}"
                    ${retryFailedDisabled ? "disabled" : ""}
                    title="${retryFailedDisabled ? "No failed jobs to retry" : "Retry failed jobs"}">
                    <i class="bi bi-arrow-repeat me-1"></i>Retry Failed
                </button>
            </td>
        `;
    elements.queueTable.appendChild(row);
  });
}

async function clearQueue(queueName, pending) {
  if (!queueName) return;

  const message =
    pending > 0
      ? `Clear ${pending} pending job(s) from "${queueName}"? This does not stop in-flight processing jobs.`
      : `Clear pending jobs from "${queueName}"?`;
  if (!confirm(message)) return;

  const url = `${ENDPOINTS.QUEUE}/${encodeURIComponent(queueName)}/clear`;
  const response = await fetch(url, { method: "POST" });
  const responseText = await response.text();
  const body = safeParseJSON(responseText, {});

  if (!response.ok) {
    const errorMessage =
      body && body.error ? body.error : `HTTP ${response.status}`;
    throw new Error(errorMessage);
  }

  const cleared = typeof body.cleared === "number" ? body.cleared : pending;
  showAlert(
    `Cleared ${cleared} pending job(s) from "${queueName}".`,
    "success",
  );
  await loadQueueStatus();
}

async function drainQueue(queueName, pending, processing) {
  if (!queueName) return;

  const message =
    pending > 0 || processing > 0
      ? `Drain "${queueName}"? This requests cancellation for ${processing} in-flight job(s) and clears ${pending} pending job(s).`
      : `Drain "${queueName}"?`;
  if (!confirm(message)) return;

  const url = `${ENDPOINTS.QUEUE}/${encodeURIComponent(queueName)}/drain`;
  const response = await fetch(url, { method: "POST" });
  const responseText = await response.text();
  const body = safeParseJSON(responseText, {});

  if (!response.ok) {
    const errorMessage =
      body && body.error ? body.error : `HTTP ${response.status}`;
    throw new Error(errorMessage);
  }

  showAlert(`Drained "${queueName}".`, "success");
  await loadQueueStatus();
  await loadWorkers();
}

async function cancelQueueJob(queueName, jobID) {
  if (!queueName || !jobID) return;

  if (
    !confirm(
      `Cancel job ${jobID} in "${queueName}"? This is best-effort: workers will stop when they next observe the cancel flag.`,
    )
  ) {
    return;
  }

  const url = `${ENDPOINTS.QUEUE}/${encodeURIComponent(queueName)}/jobs/${encodeURIComponent(jobID)}/cancel`;
  const response = await fetch(url, { method: "POST" });
  const responseText = await response.text();
  const body = safeParseJSON(responseText, {});
  if (!response.ok) {
    const errorMessage =
      body && body.error ? body.error : `HTTP ${response.status}`;
    throw new Error(errorMessage);
  }

  showAlert(`Cancel requested for ${jobID}.`, "success");
  await loadWorkers();
  await loadQueueStatus();
}

async function clearProcessing(queueName, processing) {
  if (!queueName) return;

  const message =
    processing > 0
      ? `Force-clear ${processing} processing job(s) from "${queueName}"? This will drop "in-flight" bookkeeping and ack/delete stream entries when possible.`
      : `Force-clear processing jobs from "${queueName}"?`;
  if (!confirm(message)) return;

  const url = `${ENDPOINTS.QUEUE}/${encodeURIComponent(queueName)}/clear-processing`;
  const response = await fetch(url, { method: "POST" });
  const responseText = await response.text();
  const body = safeParseJSON(responseText, {});

  if (!response.ok) {
    const errorMessage =
      body && body.error ? body.error : `HTTP ${response.status}`;
    throw new Error(errorMessage);
  }

  const cleared =
    typeof body.cleared === "number" ? body.cleared : processing;
  showAlert(
    `Cleared ${cleared} processing job(s) from "${queueName}".`,
    "success",
  );
  await loadQueueStatus();
}

async function clearFailed(queueName, failed) {
  if (!queueName) return;

  const message =
    failed > 0
      ? `Clear ${failed} failed job(s) from "${queueName}"?`
      : `Clear failed jobs from "${queueName}"?`;
  if (!confirm(message)) return;

  const url = `${ENDPOINTS.QUEUE}/${encodeURIComponent(queueName)}/clear-failed`;
  const response = await fetch(url, { method: "POST" });
  const responseText = await response.text();
  const body = safeParseJSON(responseText, {});

  if (!response.ok) {
    const errorMessage =
      body && body.error ? body.error : `HTTP ${response.status}`;
    throw new Error(errorMessage);
  }

  const cleared = typeof body.cleared === "number" ? body.cleared : failed;
  showAlert(`Cleared ${cleared} failed job(s) from "${queueName}".`, "success");
  await loadQueueStatus();
}

async function retryFailed(queueName, failed) {
  if (!queueName) return;

  const defaultLimit = Math.min(200, failed || 200);
  const limitInput = prompt(
    `Retry how many failed job(s) from "${queueName}"?`,
    String(defaultLimit),
  );
  if (limitInput == null) return;
  const limit = parseInt(limitInput, 10);
  if (!Number.isFinite(limit) || limit <= 0) {
    throw new Error("Invalid retry limit");
  }

  const message = `Retry up to ${limit} failed job(s) from "${queueName}"?`;
  if (!confirm(message)) return;

  const url = `${ENDPOINTS.QUEUE}/${encodeURIComponent(queueName)}/retry-failed?limit=${encodeURIComponent(
    String(limit),
  )}`;
  const response = await fetch(url, { method: "POST" });
  const responseText = await response.text();
  const body = safeParseJSON(responseText, {});

  if (!response.ok) {
    const errorMessage =
      body && body.error ? body.error : `HTTP ${response.status}`;
    throw new Error(errorMessage);
  }

  const retried = typeof body.retried === "number" ? body.retried : 0;
  showAlert(`Retried ${retried} failed job(s) from "${queueName}".`, "success");
  await loadQueueStatus();
}

// Load search counts with retry logic
async function loadSearchCounts() {
  let retries = 3;
  while (retries > 0) {
    try {
      const response = await fetch(`${API_BASE}/search?limit=0`, {
        timeout: 10000,
        headers: {
          Accept: "application/json",
        },
      });
      if (!response.ok)
        throw new Error(
          `HTTP ${response.status}: Failed to load search counts`,
        );

      const data = await response.json();
      elements.searchesCount.textContent = data.total || 0;
      return; // Success, exit retry loop
    } catch (error) {
      console.error(`Error loading search counts (${4 - retries}/3):`, error);
      retries--;
      if (retries > 0) {
        // Wait before retry
        await new Promise((resolve) => setTimeout(resolve, 1000));
      } else {
        // Set fallback value on final failure
        elements.searchesCount.textContent = "?";
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

  const submitBtn = document.getElementById("submitPriceGraphBtn");
  setButtonLoading(submitBtn, true);

  try {
    const origins = parseAirportList(
      document.getElementById("pgOrigins")?.value || "",
    );
    const destinations = parseAirportList(
      document.getElementById("pgDestinations")?.value || "",
    );

    if (!origins.length || !destinations.length) {
      throw new Error("At least one origin and destination are required");
    }

    const departureFrom = document.getElementById("pgDepartureFrom")?.value;
    const departureTo = document.getElementById("pgDepartureTo")?.value;
    if (!departureFrom || !departureTo) {
      throw new Error("Departure window is required");
    }

    const tripLengths = parseNumberList(
      document.getElementById("pgTripLengths")?.value || "",
    );

    const selectedClasses = [
      document.getElementById("pgClassEconomy"),
      document.getElementById("pgClassPremiumEconomy"),
      document.getElementById("pgClassBusiness"),
      document.getElementById("pgClassFirst"),
    ]
      .filter((el) => el && el.checked)
      .map((el) => el.value);

    if (!selectedClasses.length) {
      throw new Error("Select at least one cabin class");
    }

    const payload = {
      origins,
      destinations,
      departure_date_from: `${departureFrom}T00:00:00Z`,
      departure_date_to: `${departureTo}T00:00:00Z`,
      trip_lengths: tripLengths.length ? tripLengths : undefined,
      trip_type: document.getElementById("pgTripType")?.value || "round_trip",
      classes: selectedClasses,
      stops: document.getElementById("pgStops")?.value || "any",
      adults: parseInt(document.getElementById("pgAdults")?.value || "1", 10),
      children: parseInt(
        document.getElementById("pgChildren")?.value || "0",
        10,
      ),
      infants_lap: parseInt(
        document.getElementById("pgInfantsLap")?.value || "0",
        10,
      ),
      infants_seat: parseInt(
        document.getElementById("pgInfantsSeat")?.value || "0",
        10,
      ),
      currency: (document.getElementById("pgCurrency")?.value || "USD")
        .trim()
        .toUpperCase(),
    };

    const rateLimitRaw = document.getElementById("pgRateLimit")?.value;
    const rateLimit = parseInt(rateLimitRaw || "", 10);
    if (!Number.isNaN(rateLimit) && rateLimit >= 0) {
      payload.rate_limit_millis = rateLimit;
    }

    // Remove undefined keys
    Object.keys(payload).forEach((key) => {
      if (payload[key] === undefined || payload[key] === null) {
        delete payload[key];
      }
    });

    const response = await fetch(ENDPOINTS.PRICE_GRAPH_SWEEPS, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errorText = await response.text();
      const errorBody = safeParseJSON(errorText, {});
      throw new Error(
        errorBody.error || `Failed to enqueue sweep (HTTP ${response.status})`,
      );
    }

    const dataText = await response.text();
    const data = safeParseJSON(dataText, {});
    const sweepId = data?.sweep_id;

    showAlert("Price graph sweep enqueued.", "success");

    await loadPriceGraphSweeps(true);

    if (sweepId) {
      await viewPriceGraphResults(sweepId, true);
      if (elements.priceGraphResultsMeta) {
        elements.priceGraphResultsMeta.textContent = `Sweep #${sweepId} queued. Results will populate as the worker progresses.`;
      }
    }
  } catch (error) {
    console.error("Error enqueueing price graph sweep:", error);
    showAlert(`Error enqueueing sweep: ${error.message}`, "danger");
  } finally {
    setButtonLoading(submitBtn, false);
  }
}

// Load recent price graph sweeps
async function loadPriceGraphSweeps(refreshResults = false) {
  if (!elements.priceGraphTable) return;

  if (elements.sweepsStatus) {
    elements.sweepsStatus.textContent = "Loading...";
  }

  const response = await fetch(`${ENDPOINTS.PRICE_GRAPH_SWEEPS}?limit=50`);
  if (!response.ok) {
    throw new Error(
      `HTTP ${response.status}: failed to load price graph sweeps`,
    );
  }

  const responseText = await response.text();
  const data = safeParseJSON(responseText, {});
  const sweeps = Array.isArray(data)
    ? data
    : Array.isArray(data.items)
      ? data.items
      : [];

  elements.priceGraphTable.innerHTML = "";

  if (!sweeps.length) {
    const emptyRow = document.createElement("tr");
    emptyRow.innerHTML = `<td colspan="7" class="text-center py-4 text-muted">No sweeps found.</td>`;
    elements.priceGraphTable.appendChild(emptyRow);
    if (elements.sweepsStatus) elements.sweepsStatus.textContent = "";
    return;
  }

  sweeps.forEach((sweep) => {
    const row = document.createElement("tr");
    row.dataset.sweepId = sweep.id;

    const statusBadge = formatSweepStatusBadge(sweep.status);
    const coverage = `${sweep.origin_count || 0} × ${sweep.destination_count || 0}`;
    const tripWindow = formatTripLengthRange(
      sweep.trip_length_min,
      sweep.trip_length_max,
    );
    const updatedAt = sweep.updated_at
      ? new Date(sweep.updated_at).toLocaleString()
      : "—";

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
    const stillExists = sweeps.some((sweep) => sweep.id === currentSweepId);
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
      throw new Error(
        errorBody.error ||
          `Failed to load sweep results (HTTP ${response.status})`,
      );
    }

    const responseText = await response.text();
    const data = safeParseJSON(responseText, {});
    const summary = data?.sweep || {};
    const results = Array.isArray(data?.results) ? data.results : [];

    currentSweepResults = results;

    renderPriceGraphResults(summary, results);

    if (elements.priceGraphResultsMeta) {
      const statusLabel = summary.status
        ? summary.status.replace(/_/g, " ")
        : "unknown";
      const currency = (
        summary.currency ||
        results[0]?.currency ||
        "USD"
      ).toUpperCase();
      const message = `Sweep #${summary.id || sweepId} • ${results.length} fares • ${currency.toUpperCase()} • ${statusLabel}`;
      elements.priceGraphResultsMeta.textContent = message;
    }

    if (elements.exportResultsBtn) {
      elements.exportResultsBtn.disabled = results.length === 0;
    }

    if (!silent) {
      showAlert(`Loaded results for sweep #${sweepId}`, "info");
    }
  } catch (error) {
    console.error(`Error loading results for sweep ${sweepId}:`, error);
    if (elements.priceGraphResultsMeta) {
      elements.priceGraphResultsMeta.textContent = `Failed to load sweep #${sweepId}: ${error.message}`;
    }
    showAlert(`Error loading sweep results: ${error.message}`, "danger");
  }
}

// Render price graph result rows
function renderPriceGraphResults(summary, results) {
  if (!elements.priceGraphResultsTable) return;

  elements.priceGraphResultsTable.innerHTML = "";

  if (!results.length) {
    const emptyRow = document.createElement("tr");
    emptyRow.innerHTML = `<td colspan="5" class="text-center py-4 text-muted">No results stored for this sweep yet.</td>`;
    elements.priceGraphResultsTable.appendChild(emptyRow);
    return;
  }

  const formatter = buildCurrencyFormatter(
    summary?.currency || results[0]?.currency || "USD",
  );

  results.forEach((result) => {
    const row = document.createElement("tr");
    const priceValue = Number(result.price);
    const origin = escapeHtml(result.origin);
    const destination = escapeHtml(result.destination);
    row.innerHTML = `
            <td><strong>${origin}</strong> → <strong>${destination}</strong></td>
            <td>${formatDate(result.departure_date)}</td>
            <td>${formatDate(result.return_date)}</td>
            <td>${result.trip_length ?? "—"}</td>
            <td class="text-end">${Number.isFinite(priceValue) ? formatter(priceValue) : "—"}</td>
        `;
    elements.priceGraphResultsTable.appendChild(row);
  });
}

// Export current sweep results to CSV
function exportPriceGraphResults() {
  if (!currentSweepResults.length || !currentSweepId) return;

  const header = [
    "sweep_id",
    "origin",
    "destination",
    "departure_date",
    "return_date",
    "trip_length",
    "price",
    "currency",
    "queried_at",
  ];
  const rows = currentSweepResults.map((result) => [
    result.sweep_id || currentSweepId,
    result.origin || "",
    result.destination || "",
    formatDateISO(result.departure_date),
    formatDateISO(result.return_date),
    result.trip_length ?? "",
    result.price ?? "",
    result.currency || "",
    formatDateISO(result.queried_at),
  ]);

  const csvContent = [header, ...rows]
    .map((line) => line.map(escapeCsvValue).join(","))
    .join("\n");

  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);

  const link = document.createElement("a");
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
  const time = document.getElementById("time").value;
  const interval = document.getElementById("interval").value;
  const [hour, minute] = time.split(":");

  // Generate cron expression based on interval and time
  let cronExpression;
  switch (interval) {
    case "daily":
      cronExpression = `${minute} ${hour} * * *`;
      break;
    case "weekly":
      cronExpression = `${minute} ${hour} * * 1`; // Monday
      break;
    case "monthly":
      cronExpression = `${minute} ${hour} 1 * *`; // First of month
      break;
    default:
      cronExpression = `${minute} ${hour} * * *`; // Default to daily
  }

  const isDynamicDates = document.getElementById("dynamicDates").checked;

  const jobData = {
    name: document.getElementById("jobName").value,
    origin: document.getElementById("origin").value,
    destination: document.getElementById("destination").value,
    trip_type: document.getElementById("tripType").value,
    class: document.getElementById("class").value,
    stops: document.getElementById("stops").value,
    adults: parseInt(document.getElementById("adults").value),
    children: parseInt(document.getElementById("children").value),
    infants_lap: parseInt(document.getElementById("infantsLap").value),
    infants_seat: parseInt(document.getElementById("infantsSeat").value),
    currency: document.getElementById("currency").value,
    interval: interval,
    time: time,
    cron_expression: cronExpression,
    dynamic_dates: isDynamicDates,
  };

  // Add date fields based on mode
  if (isDynamicDates) {
    // Dynamic date mode
    jobData.days_from_execution =
      parseInt(document.getElementById("daysFromExecution").value, 10) || 14;
    jobData.search_window_days =
      parseInt(document.getElementById("searchWindowDays").value, 10) || 7;
    jobData.trip_length =
      parseInt(document.getElementById("tripLength").value, 10) || 7;

    // Set placeholder static dates (required by backend but not used when dynamic)
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 30);
    const dateString = futureDate.toISOString().split("T")[0];
    jobData.date_start = dateString;
    jobData.date_end = dateString;
  } else {
    // Static date mode
    jobData.date_start = document.getElementById("dateStart").value;
    jobData.date_end = document.getElementById("dateEnd").value;

    const returnDateStart = document.getElementById("returnDateStart").value;
    const returnDateEnd = document.getElementById("returnDateEnd").value;
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
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(jobData),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to create job");
    }

    // Close modal and refresh data
    const modal = bootstrap.Modal.getInstance(
      document.getElementById("newJobModal"),
    );
    modal.hide();

    // Reset form
    elements.newJobForm.reset();

    // Show success message
    showAlert("Job created successfully!", "success");

    // Refresh all data to update jobs, queue status, and workers
    await refreshData();
  } catch (error) {
    console.error("Error creating job:", error);
    showAlert(`Error creating job: ${error.message}`, "danger");
  }
}

// Run a job
async function runJob(jobId) {
  try {
    const response = await fetch(`${ENDPOINTS.JOBS}/${jobId}/run`, {
      method: "POST",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to run job");
    }

    showAlert("Job started successfully!", "success");
    await refreshData();
  } catch (error) {
    console.error("Error running job:", error);
    showAlert(`Error running job: ${error.message}`, "danger");
  }
}

// Enable a job
async function enableJob(jobId) {
  try {
    const response = await fetch(`${ENDPOINTS.JOBS}/${jobId}/enable`, {
      method: "POST",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to enable job");
    }

    showAlert("Job enabled successfully!", "success");
    await loadJobs();
  } catch (error) {
    console.error("Error enabling job:", error);
    showAlert(`Error enabling job: ${error.message}`, "danger");
  }
}

// Disable a job
async function disableJob(jobId) {
  try {
    const response = await fetch(`${ENDPOINTS.JOBS}/${jobId}/disable`, {
      method: "POST",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to disable job");
    }

    showAlert("Job disabled successfully!", "success");
    await loadJobs();
  } catch (error) {
    console.error("Error disabling job:", error);
    showAlert(`Error disabling job: ${error.message}`, "danger");
  }
}

// Delete a job
async function deleteJob(jobId) {
  if (!confirm("Are you sure you want to delete this job?")) {
    return;
  }

  try {
    const response = await fetch(`${ENDPOINTS.JOBS}/${jobId}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to delete job");
    }

    showAlert("Job deleted successfully!", "success");
    await loadJobs();
  } catch (error) {
    console.error("Error deleting job:", error);
    showAlert(`Error deleting job: ${error.message}`, "danger");
  }
}

// Helper function to format duration
function formatDuration(seconds) {
  if (seconds == null) return "N/A";
  if (seconds < 0) seconds = 0;

  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  return `${hours}h ${minutes}m ${secs}s`;
}

// Show an alert message
function showAlert(message, type = "info") {
  const alertDiv = document.createElement("div");
  alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
  alertDiv.role = "alert";

  const messageSpan = document.createElement("span");
  messageSpan.textContent = message;
  alertDiv.appendChild(messageSpan);

  const closeBtn = document.createElement("button");
  closeBtn.type = "button";
  closeBtn.className = "btn-close";
  closeBtn.setAttribute("data-bs-dismiss", "alert");
  closeBtn.setAttribute("aria-label", "Close");
  alertDiv.appendChild(closeBtn);

  // Insert at the top of the main content
  const main = document.querySelector("main");
  main.insertBefore(alertDiv, main.firstChild);

  // Auto-dismiss after 5 seconds
  setTimeout(() => {
    alertDiv.classList.remove("show");
    setTimeout(() => alertDiv.remove(), 150);
  }, 5000);
}

// Update cron expression preview
function updateCronPreview() {
  const timeInput = document.getElementById("time");
  const intervalInput = document.getElementById("interval");
  const cronPreview = document.getElementById("cronPreview");

  if (!timeInput || !intervalInput || !cronPreview) return;

  const time = timeInput.value;
  const interval = intervalInput.value;

  if (!time) return;

  const [hour, minute] = time.split(":");

  let cronExpression;
  switch (interval) {
    case "daily":
      cronExpression = `${minute} ${hour} * * *`;
      break;
    case "weekly":
      cronExpression = `${minute} ${hour} * * 1`; // Monday
      break;
    case "monthly":
      cronExpression = `${minute} ${hour} 1 * *`; // First of month
      break;
    default:
      cronExpression = `${minute} ${hour} * * *`;
  }

  cronPreview.textContent = cronExpression;
}

// Toggle between static and dynamic date modes
function toggleDateMode() {
  const dynamicDatesCheckbox = document.getElementById("dynamicDates");
  const staticDateFields = document.getElementById("staticDateFields");
  const dynamicDateFields = document.getElementById("dynamicDateFields");
  const dateStartInput = document.getElementById("dateStart");
  const dateEndInput = document.getElementById("dateEnd");
  const daysFromExecutionInput = document.getElementById("daysFromExecution");
  const searchWindowDaysInput = document.getElementById("searchWindowDays");

  if (!dynamicDatesCheckbox || !staticDateFields || !dynamicDateFields) return;

  const isDynamic = dynamicDatesCheckbox.checked;

  if (isDynamic) {
    // Show dynamic fields, hide static fields
    staticDateFields.style.display = "none";
    dynamicDateFields.style.display = "block";

    // Remove required attribute from static date fields
    if (dateStartInput) dateStartInput.removeAttribute("required");
    if (dateEndInput) dateEndInput.removeAttribute("required");

    // Add required attribute to dynamic fields
    if (daysFromExecutionInput)
      daysFromExecutionInput.setAttribute("required", "required");
    if (searchWindowDaysInput)
      searchWindowDaysInput.setAttribute("required", "required");
  } else {
    // Show static fields, hide dynamic fields
    staticDateFields.style.display = "block";
    dynamicDateFields.style.display = "none";

    // Add required attribute to static date fields
    if (dateStartInput) dateStartInput.setAttribute("required", "required");
    if (dateEndInput) dateEndInput.setAttribute("required", "required");

    // Remove required attribute from dynamic fields
    if (daysFromExecutionInput)
      daysFromExecutionInput.removeAttribute("required");
    if (searchWindowDaysInput)
      searchWindowDaysInput.removeAttribute("required");
  }
}

function initMacroTokenHelpers() {
  const originSelect = document.getElementById("pgOriginRegionSelect");
  const originAddBtn = document.getElementById("pgOriginRegionAddBtn");
  const destSelect = document.getElementById("pgDestinationRegionSelect");
  const destAddBtn = document.getElementById("pgDestinationRegionAddBtn");

  if (originAddBtn) {
    originAddBtn.addEventListener("click", () => {
      const token = originSelect?.value || "";
      if (!token) return;
      appendTokenToTextarea("pgOrigins", token);
    });
  }

  if (destAddBtn) {
    destAddBtn.addEventListener("click", () => {
      const token = destSelect?.value || "";
      if (!token) return;
      appendTokenToTextarea("pgDestinations", token);
    });
  }

  // Populate selects asynchronously; UI should still function without it.
  loadMacroMetadata().catch((err) =>
    console.warn("Failed to load macro metadata", err),
  );
}

async function loadMacroMetadata() {
  const originSelect = document.getElementById("pgOriginRegionSelect");
  const originAddBtn = document.getElementById("pgOriginRegionAddBtn");
  const destSelect = document.getElementById("pgDestinationRegionSelect");
  const destAddBtn = document.getElementById("pgDestinationRegionAddBtn");

  if (!originSelect && !destSelect) return;

  const response = await fetch(ENDPOINTS.REGIONS);
  if (!response.ok) {
    throw new Error(`Failed to load regions (HTTP ${response.status})`);
  }

  const regions = safeParseJSON(await response.text(), []);
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

  regions.forEach((region) => {
    if (!region?.token) return;
    const count =
      typeof region.airport_count === "number" ? region.airport_count : null;
    const samples = Array.isArray(region.sample_airports)
      ? region.sample_airports.slice(0, 3)
      : [];
    const labelParts = [region.token];
    if (count != null) labelParts.push(`(${count})`);
    if (samples.length) labelParts.push(`— ${samples.join(", ")}`);

    const option = document.createElement("option");
    option.value = region.token;
    option.textContent = labelParts.join(" ");
    selectEl.appendChild(option);
  });

  // Special token: not returned by /regions because it requires a server-side Postgres-backed expansion.
  const worldAllOption = document.createElement("option");
  worldAllOption.value = "REGION:WORLD_ALL";
  worldAllOption.textContent = "REGION:WORLD_ALL (all airports in DB; large)";
  selectEl.appendChild(worldAllOption);

  if (currentValue) {
    selectEl.value = currentValue;
  }
}

function appendTokenToTextarea(textareaId, token) {
  const textarea = document.getElementById(textareaId);
  if (!textarea) return;

  const normalized = String(token).trim().toUpperCase();
  if (!normalized) return;

  const existing = parseAirportList(textarea.value || "");
  if (existing.includes(normalized)) {
    return;
  }

  const next = existing.concat([normalized]);
  textarea.value = next.join(", ");
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
    .split(/[\n,]/)
    .map((code) => code.trim().toUpperCase())
    .filter((code) => code.length > 0);
}

function parseNumberList(input) {
  const text = String(input || "");
  if (!text.trim()) return [];

  const values = [];
  const re = /(-?\d+)\s*(?:-\s*(-?\d+))?/g;
  let match;
  while ((match = re.exec(text)) !== null) {
    const start = parseInt(match[1], 10);
    if (Number.isNaN(start) || !Number.isFinite(start)) continue;

    if (match[2] != null) {
      const end = parseInt(match[2], 10);
      if (Number.isNaN(end) || !Number.isFinite(end)) continue;

      const step = start <= end ? 1 : -1;
      for (let v = start; v !== end + step; v += step) {
        values.push(v);
      }
    } else {
      values.push(start);
    }
  }

  return Array.from(new Set(values))
    .filter((value) => !Number.isNaN(value) && Number.isFinite(value))
    .sort((a, b) => a - b);
}

function formatTripLengthRange(minVal, maxVal) {
  if (minVal == null && maxVal == null) return "—";
  if (minVal == null) return `${maxVal}`;
  if (maxVal == null) return `${minVal}`;
  if (minVal === maxVal) return `${minVal}`;
  return `${minVal} - ${maxVal}`;
}

function formatDate(value) {
  if (!value) return "—";
  try {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "—";
    }
    return date.toLocaleDateString();
  } catch (err) {
    return "—";
  }
}

function formatDateISO(value) {
  if (!value) return "";
  try {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    return date.toISOString();
  } catch (err) {
    return "";
  }
}

function formatSweepStatusBadge(status) {
  const normalized = (status || "unknown").toLowerCase();
  const badgeClasses = {
    queued: "bg-secondary",
    running: "bg-info text-dark",
    completed: "bg-success",
    completed_with_errors: "bg-warning text-dark",
    failed: "bg-danger",
    unknown: "bg-secondary",
  };
  const cssClass = badgeClasses[normalized] || "bg-secondary";
  const label = escapeHtml(normalized.replace(/_/g, " "));
  return `<span class="badge ${cssClass} text-uppercase">${label}</span>`;
}

function highlightSelectedSweepRow(sweepId) {
  if (!elements.priceGraphTable) return;
  const rows = elements.priceGraphTable.querySelectorAll("tr");
  rows.forEach((row) => {
    if (row.dataset && row.dataset.sweepId) {
      row.classList.toggle(
        "table-active",
        row.dataset.sweepId === String(sweepId),
      );
    } else {
      row.classList.remove("table-active");
    }
  });
}

function buildCurrencyFormatter(currencyCode) {
  try {
    const formatter = new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: (currencyCode || "USD").toUpperCase(),
      maximumFractionDigits: 2,
    });
    return (value) => formatter.format(value);
  } catch (err) {
    return (value) => value.toFixed(2);
  }
}

function escapeCsvValue(value) {
  if (value == null) return "";
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
    console.warn(
      "Initial JSON parse failed, trying fallback methods:",
      e.message,
    );

    // Try to handle duplicate JSON objects like {"a":1}{"b":2}
    try {
      const jsonObjects = responseText.match(
        /\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}/g,
      );
      if (jsonObjects && jsonObjects.length > 0) {
        // Use the last valid JSON object
        const lastObject = jsonObjects[jsonObjects.length - 1];
        return JSON.parse(lastObject);
      }
    } catch (e2) {
      console.warn("Duplicate object parsing failed:", e2.message);
    }

    // Try to handle array format like [][{data}]
    try {
      const arrayMatch = responseText.match(/\]\[(.+)\]$/);
      if (arrayMatch) {
        return JSON.parse("[" + arrayMatch[1] + "]");
      }
    } catch (e3) {
      console.warn("Array format parsing failed:", e3.message);
    }

    // Try to extract any valid JSON from the response
    try {
      const jsonMatch = responseText.match(/(\[.*\]|\{.*\})/);
      if (jsonMatch) {
        return JSON.parse(jsonMatch[1]);
      }
    } catch (e4) {
      console.warn("Pattern matching failed:", e4.message);
    }

    console.error(
      "All JSON parsing methods failed for response:",
      responseText,
    );
    return fallbackValue;
  }
}

// ============================================
// Continuous Sweep Functions
// ============================================

// Interval for auto-refreshing sweep status
let sweepStatusInterval = null;

const SWEEP_CONFIG_FIELD_IDS = [
  "sweepClass",
  "sweepTripLengths",
  "sweepPacingMode",
  "sweepTargetHours",
  "sweepMinDelay",
];

let sweepConfigLastServerValues = {};
const sweepConfigDirtyFields = new Set();

function getSweepConfigServerValuesFromStatus(status) {
  const values = {};
  values.sweepClass = status?.class ?? "";
  values.sweepTripLengths = Array.isArray(status?.trip_lengths)
    ? status.trip_lengths.join(",")
    : "";
  values.sweepPacingMode = status?.pacing_mode ?? "";
  values.sweepTargetHours =
    status?.target_duration_hours != null
      ? String(status.target_duration_hours)
      : "";
  values.sweepMinDelay =
    status?.min_delay_ms != null ? String(status.min_delay_ms) : "";
  return values;
}

function updateSweepConfigDirtyState(fieldId) {
  const el = document.getElementById(fieldId);
  if (!el) return;

  const serverValue = sweepConfigLastServerValues[fieldId];
  if (serverValue == null) {
    sweepConfigDirtyFields.add(fieldId);
    return;
  }

  if (String(el.value) === String(serverValue)) {
    sweepConfigDirtyFields.delete(fieldId);
    return;
  }

  sweepConfigDirtyFields.add(fieldId);
}

function initSweepConfigFormState() {
  SWEEP_CONFIG_FIELD_IDS.forEach((fieldId) => {
    const el = document.getElementById(fieldId);
    if (!el) return;

    const handler = () => updateSweepConfigDirtyState(fieldId);
    el.addEventListener("input", handler);
    el.addEventListener("change", handler);
  });
}

function syncSweepConfigFormFromStatus(status) {
  const nextServerValues = getSweepConfigServerValuesFromStatus(status);
  sweepConfigLastServerValues = nextServerValues;

  // If the user has typed but the server values catch up (or we just loaded status),
  // re-evaluate dirty state so matching fields aren't treated as "locked".
  SWEEP_CONFIG_FIELD_IDS.forEach((fieldId) =>
    updateSweepConfigDirtyState(fieldId),
  );

  SWEEP_CONFIG_FIELD_IDS.forEach((fieldId) => {
    const el = document.getElementById(fieldId);
    if (!el) return;

    const isFocused = document.activeElement === el;
    if (isFocused || sweepConfigDirtyFields.has(fieldId)) {
      return;
    }

    const serverValue = nextServerValues[fieldId];
    if (serverValue == null) return;
    el.value = serverValue;
  });
}

// Initialize continuous sweep controls
function initContinuousSweepControls() {
  const startBtn = document.getElementById("sweepStartBtn");
  const pauseBtn = document.getElementById("sweepPauseBtn");
  const resumeBtn = document.getElementById("sweepResumeBtn");
  const stopBtn = document.getElementById("sweepStopBtn");
  const skipBtn = document.getElementById("sweepSkipBtn");
  const restartBtn = document.getElementById("sweepRestartBtn");
  const refreshStatusBtn = document.getElementById("refreshSweepStatusBtn");
  const refreshStatsBtn = document.getElementById("refreshSweepStatsBtn");
  const refreshResultsBtn = document.getElementById("refreshSweepResultsBtn");
  const configForm = document.getElementById("sweepConfigForm");

  if (startBtn) startBtn.addEventListener("click", startContinuousSweep);
  if (pauseBtn) pauseBtn.addEventListener("click", pauseContinuousSweep);
  if (resumeBtn) resumeBtn.addEventListener("click", resumeContinuousSweep);
  if (stopBtn) stopBtn.addEventListener("click", stopContinuousSweep);
  if (skipBtn) skipBtn.addEventListener("click", skipCurrentRoute);
  if (restartBtn) restartBtn.addEventListener("click", restartCurrentSweep);
  if (refreshStatusBtn)
    refreshStatusBtn.addEventListener("click", loadContinuousSweepStatus);
  if (refreshStatsBtn)
    refreshStatsBtn.addEventListener("click", loadContinuousSweepStats);
  if (refreshResultsBtn)
    refreshResultsBtn.addEventListener("click", loadContinuousSweepResults);
  if (configForm) configForm.addEventListener("submit", updateSweepConfig);

  initSweepConfigFormState();

  // Initial load
  loadContinuousSweepStatus();
  loadContinuousSweepStats();
  loadContinuousSweepResults();

  // Auto-refresh every 5 seconds when sweep is running
  sweepStatusInterval = setInterval(loadContinuousSweepStatus, 5000);
}

// Load continuous sweep status
async function loadContinuousSweepStatus() {
  try {
    const response = await fetch(`${ENDPOINTS.CONTINUOUS_SWEEP}/status`);

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const body = await response.json();
    if (!body || body.initialized === false) {
      showSweepNotReady();
      return;
    }

    const status = body.status ?? body;
    updateSweepStatusUI(status);
  } catch (error) {
    console.error("Error loading sweep status:", error);
  }
}

// Update sweep status UI
function updateSweepStatusUI(status) {
  const notReady = document.getElementById("sweepStatusNotReady");
  const content = document.getElementById("sweepStatusContent");

  if (!status) {
    showSweepNotReady();
    return;
  }

  // Show status content
  if (notReady) notReady.style.display = "none";
  if (content) content.style.display = "block";

  // Update status badge
  const statusBadge = document.getElementById("sweepStatusBadge");
  const statusText = document.getElementById("sweepStatusText");
  if (statusBadge && statusText) {
    const cabinLabel = (() => {
      switch (status.class) {
        case "economy":
          return "Economy";
        case "premium_economy":
          return "Premium Economy";
        case "business":
          return "Business";
        case "first":
          return "First";
        default:
          return null;
      }
    })();
    if (!status.is_running) {
      statusBadge.className = "badge bg-secondary";
      statusBadge.textContent = "Stopped";
      statusText.textContent = cabinLabel
        ? `Sweep is not running • ${cabinLabel}`
        : "Sweep is not running";
    } else if (status.is_paused) {
      statusBadge.className = "badge bg-warning";
      statusBadge.textContent = "Paused";
      statusText.textContent = cabinLabel
        ? `Sweep is paused • ${cabinLabel}`
        : "Sweep is paused";
    } else {
      statusBadge.className = "badge bg-success";
      statusBadge.textContent = "Running";
      const modeLabel =
        status.pacing_mode === "adaptive"
          ? "Adaptive mode"
          : "Fixed delay mode";
      statusText.textContent = cabinLabel
        ? `${modeLabel} • ${cabinLabel}`
        : modeLabel;
    }
  }

  // Update progress bar
  const progressBar = document.getElementById("sweepProgressBar");
  const progressPercent = status.progress_percent || 0;
  if (progressBar) {
    progressBar.style.width = `${progressPercent}%`;
    progressBar.textContent = `${progressPercent.toFixed(1)}%`;
  }

  // Update route index
  const routeIndex = document.getElementById("sweepRouteIndex");
  const totalRoutes = document.getElementById("sweepTotalRoutes");
  if (routeIndex) routeIndex.textContent = status.route_index || 0;
  if (totalRoutes) totalRoutes.textContent = status.total_routes || 0;

  // Update stats
  const sweepNumber = document.getElementById("sweepNumber");
  const queriesPerHour = document.getElementById("sweepQueriesPerHour");
  const completed = document.getElementById("sweepCompleted");
  const errors = document.getElementById("sweepErrors");
  const currentRoute = document.getElementById("sweepCurrentRoute");
  const currentDelay = document.getElementById("sweepCurrentDelay");
  const estCompletion = document.getElementById("sweepEstCompletion");

  if (sweepNumber) sweepNumber.textContent = status.sweep_number || 1;
  if (queriesPerHour)
    queriesPerHour.textContent = (status.queries_per_hour || 0).toFixed(1);
  if (completed) completed.textContent = status.queries_completed || 0;
  if (errors) errors.textContent = status.errors_count || 0;
  if (currentRoute) {
    currentRoute.textContent =
      status.current_origin && status.current_destination
        ? `${status.current_origin} → ${status.current_destination}`
        : "-";
  }
  if (currentDelay)
    currentDelay.textContent = `${status.current_delay_ms || 0}ms`;
  if (estCompletion) {
    if (status.estimated_completion) {
      estCompletion.textContent = new Date(
        status.estimated_completion,
      ).toLocaleString();
    } else {
      estCompletion.textContent = "-";
    }
  }

  // Show last error if present
  const lastError = document.getElementById("sweepLastError");
  const lastErrorText = document.getElementById("sweepLastErrorText");
  if (status.last_error && lastError && lastErrorText) {
    lastError.style.display = "block";
    lastErrorText.textContent = status.last_error;
  } else if (lastError) {
    lastError.style.display = "none";
  }

  // Update button states
  updateSweepButtonStates(status);

  // Sync config inputs without overwriting in-progress edits.
  syncSweepConfigFormFromStatus(status);
}

// Show sweep not ready state
function showSweepNotReady() {
  const notReady = document.getElementById("sweepStatusNotReady");
  const content = document.getElementById("sweepStatusContent");

  if (notReady) notReady.style.display = "block";
  if (content) content.style.display = "none";

  // Reset button states
  const startBtn = document.getElementById("sweepStartBtn");
  const pauseBtn = document.getElementById("sweepPauseBtn");
  const resumeBtn = document.getElementById("sweepResumeBtn");
  const stopBtn = document.getElementById("sweepStopBtn");
  const skipBtn = document.getElementById("sweepSkipBtn");
  const restartBtn = document.getElementById("sweepRestartBtn");

  if (startBtn) startBtn.disabled = false;
  if (pauseBtn) pauseBtn.disabled = true;
  if (resumeBtn) resumeBtn.disabled = true;
  if (stopBtn) stopBtn.disabled = true;
  if (skipBtn) skipBtn.disabled = true;
  if (restartBtn) restartBtn.disabled = true;
}

// Update button states based on sweep status
function updateSweepButtonStates(status) {
  const startBtn = document.getElementById("sweepStartBtn");
  const pauseBtn = document.getElementById("sweepPauseBtn");
  const resumeBtn = document.getElementById("sweepResumeBtn");
  const stopBtn = document.getElementById("sweepStopBtn");
  const skipBtn = document.getElementById("sweepSkipBtn");
  const restartBtn = document.getElementById("sweepRestartBtn");

  const isRunning = status.is_running;
  const isPaused = status.is_paused;

  if (startBtn) startBtn.disabled = isRunning;
  if (pauseBtn) pauseBtn.disabled = !isRunning || isPaused;
  if (resumeBtn) resumeBtn.disabled = !isRunning || !isPaused;
  if (stopBtn) stopBtn.disabled = !isRunning;
  if (skipBtn) skipBtn.disabled = !isRunning || isPaused;
  if (restartBtn) restartBtn.disabled = !isRunning;
}

// Start continuous sweep
async function startContinuousSweep() {
  try {
    setButtonLoading(document.getElementById("sweepStartBtn"), true);

    const response = await fetch(`${ENDPOINTS.CONTINUOUS_SWEEP}/start`, {
      method: "POST",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to start sweep");
    }

    showAlert("Continuous sweep started", "success");
    await loadContinuousSweepStatus();
  } catch (error) {
    console.error("Error starting sweep:", error);
    showAlert(`Error starting sweep: ${error.message}`, "danger");
  } finally {
    setButtonLoading(document.getElementById("sweepStartBtn"), false);
  }
}

// Pause continuous sweep
async function pauseContinuousSweep() {
  try {
    const response = await fetch(`${ENDPOINTS.CONTINUOUS_SWEEP}/pause`, {
      method: "POST",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to pause sweep");
    }

    showAlert("Sweep paused", "info");
    await loadContinuousSweepStatus();
  } catch (error) {
    console.error("Error pausing sweep:", error);
    showAlert(`Error pausing sweep: ${error.message}`, "danger");
  }
}

// Resume continuous sweep
async function resumeContinuousSweep() {
  try {
    const response = await fetch(`${ENDPOINTS.CONTINUOUS_SWEEP}/resume`, {
      method: "POST",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to resume sweep");
    }

    showAlert("Sweep resumed", "success");
    await loadContinuousSweepStatus();
  } catch (error) {
    console.error("Error resuming sweep:", error);
    showAlert(`Error resuming sweep: ${error.message}`, "danger");
  }
}

// Stop continuous sweep
async function stopContinuousSweep() {
  if (
    !confirm(
      "Are you sure you want to stop the continuous sweep? Progress will be saved.",
    )
  ) {
    return;
  }

  try {
    const response = await fetch(`${ENDPOINTS.CONTINUOUS_SWEEP}/stop`, {
      method: "POST",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to stop sweep");
    }

    showAlert("Sweep stopped. Progress has been saved.", "warning");
    await loadContinuousSweepStatus();
  } catch (error) {
    console.error("Error stopping sweep:", error);
    showAlert(`Error stopping sweep: ${error.message}`, "danger");
  }
}

// Skip current route
async function skipCurrentRoute() {
  try {
    const response = await fetch(`${ENDPOINTS.CONTINUOUS_SWEEP}/skip`, {
      method: "POST",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to skip route");
    }

    showAlert("Route skipped", "info");
    await loadContinuousSweepStatus();
  } catch (error) {
    console.error("Error skipping route:", error);
    showAlert(`Error skipping route: ${error.message}`, "danger");
  }
}

// Restart current sweep
async function restartCurrentSweep() {
  if (
    !confirm("Are you sure you want to restart the sweep from the beginning?")
  ) {
    return;
  }

  try {
    const response = await fetch(`${ENDPOINTS.CONTINUOUS_SWEEP}/restart`, {
      method: "POST",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to restart sweep");
    }

    showAlert("Sweep restarted from beginning", "success");
    await loadContinuousSweepStatus();
  } catch (error) {
    console.error("Error restarting sweep:", error);
    showAlert(`Error restarting sweep: ${error.message}`, "danger");
  }
}

// Update sweep configuration
async function updateSweepConfig(event) {
  event.preventDefault();

  const cabinClass = document.getElementById("sweepClass")?.value;
  const tripLengthsRaw =
    document.getElementById("sweepTripLengths")?.value || "";
  const tripLengths = parseNumberList(tripLengthsRaw);
  const pacingMode = document.getElementById("sweepPacingMode")?.value;
  const targetHours = parseInt(
    document.getElementById("sweepTargetHours")?.value || "24",
    10,
  );
  const minDelay = parseInt(
    document.getElementById("sweepMinDelay")?.value || "3000",
    10,
  );

  if (tripLengthsRaw.trim() !== "" && tripLengths.length === 0) {
    showAlert(
      "Trip Lengths must be a list (e.g. 3,5,7) or range (e.g. 3-14).",
      "danger",
    );
    return;
  }

  try {
    const payload = {
      class: cabinClass,
      pacing_mode: pacingMode,
      target_duration_hours: targetHours,
      min_delay_ms: minDelay,
    };
    if (tripLengthsRaw.trim() !== "") {
      payload.trip_lengths = tripLengths;
    }

    const response = await fetch(`${ENDPOINTS.CONTINUOUS_SWEEP}/config`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to update config");
    }

    const data = await response.json();
    sweepConfigDirtyFields.clear();
    if (data?.status) {
      updateSweepStatusUI(data.status);
    } else {
      await loadContinuousSweepStatus();
    }

    showAlert("Configuration updated", "success");
  } catch (error) {
    console.error("Error updating config:", error);
    showAlert(`Error updating config: ${error.message}`, "danger");
  }
}

// Load historical sweep stats
async function loadContinuousSweepStats() {
  const table = document.getElementById("sweepStatsTable");
  if (!table) return;

  try {
    const response = await fetch(
      `${ENDPOINTS.CONTINUOUS_SWEEP}/stats?limit=10`,
    );

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const data = await response.json();
    const stats = data.stats || [];

    table.innerHTML = "";

    if (stats.length === 0) {
      table.innerHTML = `<tr><td colspan="9" class="text-center py-3 text-muted">No historical data yet.</td></tr>`;
      return;
    }

    stats.forEach((stat) => {
      const row = document.createElement("tr");

      const startedAt = stat.started_at
        ? new Date(stat.started_at).toLocaleString()
        : "-";
      const completedAt = stat.completed_at
        ? new Date(stat.completed_at).toLocaleString()
        : "-";
      const duration = stat.total_duration_seconds
        ? formatDuration(stat.total_duration_seconds)
        : "-";
      const avgDelay = stat.avg_delay_ms ? `${stat.avg_delay_ms}ms` : "-";
      const priceRange =
        stat.min_price_found != null && stat.max_price_found != null
          ? `$${stat.min_price_found.toFixed(0)} - $${stat.max_price_found.toFixed(0)}`
          : "-";

      row.innerHTML = `
                <td>${stat.sweep_number || "-"}</td>
                <td>${startedAt}</td>
                <td>${completedAt}</td>
                <td>${duration}</td>
                <td>${stat.total_routes || "-"}</td>
                <td class="text-success">${stat.successful_queries || 0}</td>
                <td class="text-danger">${stat.failed_queries || 0}</td>
                <td>${avgDelay}</td>
                <td>${priceRange}</td>
            `;

      table.appendChild(row);
    });
  } catch (error) {
    console.error("Error loading sweep stats:", error);
    table.innerHTML = `<tr><td colspan="9" class="text-center py-3 text-danger">Failed to load stats</td></tr>`;
  }
}

// Load continuous sweep results
async function loadContinuousSweepResults() {
  const table = document.getElementById("sweepResultsTable");
  const container = document.getElementById("sweepResultsContainer");
  if (!table || !container) return;

  try {
    const response = await fetch(
      `${ENDPOINTS.CONTINUOUS_SWEEP}/results?limit=100`,
    );

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const data = await response.json();
    const results = data.results || [];

    container.style.display = "block";
    table.innerHTML = "";

    if (results.length === 0) {
      table.innerHTML = `<tr><td colspan="7" class="text-center py-3 text-muted">No results captured yet. Start a sweep to begin collecting data.</td></tr>`;
      return;
    }

    results.forEach((result) => {
      const row = document.createElement("tr");

      const origin = escapeHtml(result.origin);
      const destination = escapeHtml(result.destination);
      const departureDate = result.departure_date
        ? new Date(result.departure_date).toLocaleDateString()
        : "-";
      const returnDate = result.return_date
        ? new Date(result.return_date).toLocaleDateString()
        : "-";
      const tripLength = result.trip_length != null ? result.trip_length : "-";
      const price =
        result.price != null ? `$${Number(result.price).toFixed(2)}` : "-";
      const costPerMile =
        result.cost_per_mile != null
          ? `$${Number(result.cost_per_mile).toFixed(4)}`
          : "-";

      row.innerHTML = `
                <td>${origin}</td>
                <td>${destination}</td>
                <td>${departureDate}</td>
                <td>${returnDate}</td>
                <td>${tripLength}</td>
                <td class="fw-bold">${price}</td>
                <td>${costPerMile}</td>
            `;

      table.appendChild(row);
    });
  } catch (error) {
    console.error("Error loading sweep results:", error);
    table.innerHTML = `<tr><td colspan="7" class="text-center py-3 text-danger">Failed to load results</td></tr>`;
  }
}

// Initialize when DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
  initAdminPanel();
  initContinuousSweepControls();
});
