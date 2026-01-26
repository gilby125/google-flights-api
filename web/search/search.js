// Flight Search JavaScript

// API endpoints
const API_BASE = "/api";
const ENDPOINTS = {
  SEARCH: `${API_BASE}/search`,
  PRICE_HISTORY: `${API_BASE}/price-history`,
  AIRPORTS: `${API_BASE}/airports`,
  ADMIN_BULK_JOBS: "/api/v1/admin/bulk-jobs",
};

const AIRPORT_CACHE_KEY = "flight-search-airports-cache";
const AIRPORT_CACHE_TTL_MS = 24 * 60 * 60 * 1000; // 24 hours

// DOM elements
const elements = {
  searchForm: document.getElementById("searchForm"),
  advancedBtn: document.getElementById("advancedBtn"),
  advancedOptions: document.getElementById("advancedOptions"),
  loadingIndicator: document.getElementById("loadingIndicator"),
  priceGraphCard: document.getElementById("priceGraphCard"),
  priceGraph: document.getElementById("priceGraph"),
  resultsCard: document.getElementById("resultsCard"),
  flightResults: document.getElementById("flightResults"),
  sortResults: document.getElementById("sortResults"),
  submitButton: document.querySelector("#searchForm button[type='submit']"),
  recurringBtn: document.getElementById("recurringBtn"),
  recurringOptions: document.getElementById("recurringOptions"),
  createRecurringBtn: document.getElementById("createRecurringBtn"),
  recurringStatus: document.getElementById("recurringStatus"),
};

const inputs = {
  origin: document.getElementById("origin"),
  destination: document.getElementById("destination"),
  departureDate: document.getElementById("departure_date"),
  returnDate: document.getElementById("return_date"),
  tripType: document.getElementById("trip_type"),
  travelClass: document.getElementById("class"),
  stops: document.getElementById("stops"),
  adults: document.getElementById("adults"),
  children: document.getElementById("children"),
  infantsLap: document.getElementById("infantsLap"),
  infantsSeat: document.getElementById("infantsSeat"),
  currency: document.getElementById("currency"),
  recurringName: document.getElementById("recurringName"),
  recurringInterval: document.getElementById("recurringInterval"),
  recurringTime: document.getElementById("recurringTime"),
  recurringDynamicDates: document.getElementById("recurringDynamicDates"),
  recurringDaysFromExecution: document.getElementById(
    "recurringDaysFromExecution",
  ),
  recurringSearchWindowDays: document.getElementById(
    "recurringSearchWindowDays",
  ),
  recurringTripLength: document.getElementById("recurringTripLength"),
};

// Chart instance
let priceChart = null;
let currentOffers = [];
let currentSearchParams = null;
let currentRoutes = null;
let currentActiveRouteIndex = 0;
let routePaneByIndex = new Map();
let advancedOptionsVisible =
  !!elements.advancedOptions &&
  getComputedStyle(elements.advancedOptions).display !== "none";

function escapeHtml(value) {
  if (value == null) return "";
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/\"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

function extractAirlineCodeFromFlightNumber(flightNumber) {
  const input = String(flightNumber || "")
    .trim()
    .toUpperCase();
  if (input.length < 2) return "";
  const prefix = input.slice(0, 2);
  return /^[A-Z0-9]{2}$/.test(prefix) ? prefix : "";
}

function normalizeAirportToken(value) {
  const trimmed = String(value || "").trim();
  if (!trimmed) return "";
  const firstPart = trimmed.includes(" - ") ? trimmed.split(" - ")[0] : trimmed;
  return firstPart
    .trim()
    .toUpperCase()
    .replace(/^[,;]+|[,;]+$/g, "");
}

function splitAirportList(raw) {
  const input = String(raw || "").trim();
  if (!input) return [];

  const upper = input.toUpperCase().trim();
  const hasSeparators = /[,;|/\\]/.test(upper);
  const spaceSeparatedCodes = /^[A-Z0-9]{3}(?:\s+[A-Z0-9]{3})+$/.test(upper);

  const parts = hasSeparators
    ? upper.split(/[,;|/\\]+/g)
    : spaceSeparatedCodes
      ? upper.split(/\s+/g)
      : [upper];

  const out = [];
  const seen = new Set();
  for (const part of parts) {
    const code = normalizeAirportToken(part);
    if (!/^[A-Z0-9]{3}$/.test(code)) continue;
    if (seen.has(code)) continue;
    seen.add(code);
    out.push(code);
  }
  return out;
}

function parseRouteInputs(originRaw, destinationRaw) {
  const originValue = String(originRaw || "").trim();
  const destinationValue = String(destinationRaw || "").trim();

  if (!originValue) {
    throw new Error("Please fill in origin.");
  }

  let originPart = originValue;
  let destinationPart = destinationValue;

  if (
    !destinationPart &&
    (originPart.includes(">") || originPart.includes("→"))
  ) {
    const parts = originPart.split(/[>→]/);
    if (parts.length !== 2) {
      throw new Error(
        'Invalid format. Use "ORIGINS>DESTINATIONS" (e.g. MKE,MSN>FLL,MIA).',
      );
    }
    originPart = parts[0].trim();
    destinationPart = parts[1].trim();
  }

  if (!destinationPart) {
    throw new Error("Please fill in destination.");
  }

  const origins = splitAirportList(originPart);
  const destinations = splitAirportList(destinationPart);

  if (!origins.length) {
    throw new Error("No valid origin airport codes found.");
  }
  if (!destinations.length) {
    throw new Error("No valid destination airport codes found.");
  }

  return { origins, destinations };
}

function renderAirlineGroupBadges(tokens) {
  if (!Array.isArray(tokens) || tokens.length === 0) return "";

  const labelFor = (token) => {
    switch (token) {
      case "GROUP:LOW_COST":
        return { label: "Low-cost", cls: "bg-success" };
      case "GROUP:STAR_ALLIANCE":
        return { label: "Star Alliance", cls: "bg-primary" };
      case "GROUP:ONEWORLD":
        return { label: "oneworld", cls: "bg-info text-dark" };
      case "GROUP:SKYTEAM":
        return { label: "SkyTeam", cls: "bg-secondary" };
      default:
        return { label: token, cls: "bg-light text-dark border" };
    }
  };

  return tokens
    .map((raw) => {
      const token = String(raw || "")
        .trim()
        .toUpperCase();
      if (!token) return "";
      const meta = labelFor(token);
      return `<span class="badge ${meta.cls} me-1">${escapeHtml(meta.label)}</span>`;
    })
    .filter(Boolean)
    .join("");
}

// Initialize the search page
function initSearchPage() {
  const today = new Date().toISOString().split("T")[0];
  if (inputs.departureDate) {
    inputs.departureDate.min = today;
  }
  if (inputs.returnDate) {
    inputs.returnDate.min = today;
  }

  if (elements.advancedBtn && elements.advancedOptions) {
    elements.advancedBtn.addEventListener("click", () => {
      advancedOptionsVisible = !advancedOptionsVisible;
      elements.advancedOptions.style.display = advancedOptionsVisible
        ? "block"
        : "none";
    });
  }

  if (elements.searchForm) {
    elements.searchForm.addEventListener("submit", handleSearch);
  }

  if (elements.recurringBtn && elements.recurringOptions) {
    elements.recurringBtn.addEventListener("click", () => {
      const visible =
        getComputedStyle(elements.recurringOptions).display !== "none";
      elements.recurringOptions.style.display = visible ? "none" : "block";
      if (elements.recurringStatus) elements.recurringStatus.textContent = "";

      const dynamicEnabled = !!inputs.recurringDynamicDates?.checked;
      setRecurringDynamicFieldVisibility(dynamicEnabled);
    });
  }

  if (inputs.recurringDynamicDates) {
    inputs.recurringDynamicDates.addEventListener("change", (e) => {
      setRecurringDynamicFieldVisibility(!!e.target.checked);
    });
  }

  if (elements.createRecurringBtn) {
    elements.createRecurringBtn.addEventListener(
      "click",
      createRecurringSearch,
    );
  }

  if (elements.sortResults) {
    elements.sortResults.addEventListener("change", sortFlightResults);
  }

  loadAirports();
}

function setRecurringDynamicFieldVisibility(show) {
  document.querySelectorAll(".recurringDynamicField").forEach((el) => {
    el.style.display = show ? "block" : "none";
  });
}

function stopsToApiValue(stopsValue) {
  switch (String(stopsValue || "").trim()) {
    case "non_stop":
      return "nonstop";
    case "1":
    case "one_stop":
      return "one_stop";
    case "2":
    case "two_stops":
      return "two_stops";
    case "two_stops_plus":
      return "two_stops_plus";
    case "any":
    default:
      return "any";
  }
}

function cronFromIntervalAndTime(interval, timeValue) {
  const time = String(timeValue || "07:00");
  const [hour, minute] = time.split(":");
  const safeHour = hour ?? "07";
  const safeMinute = minute ?? "00";

  switch (String(interval || "daily")) {
    case "weekly":
      return `${safeMinute} ${safeHour} * * 1`;
    case "monthly":
      return `${safeMinute} ${safeHour} 1 * *`;
    case "daily":
    default:
      return `${safeMinute} ${safeHour} * * *`;
  }
}

async function createRecurringSearch() {
  try {
    if (!inputs.origin || !inputs.destination || !inputs.departureDate) {
      throw new Error("Origin, destination, and departure date are required.");
    }

    const parsedRoutes = parseRouteInputs(
      inputs.origin.value,
      inputs.destination.value,
    );
    const tripType = inputs.tripType ? inputs.tripType.value : "round_trip";
    const travelClass = inputs.travelClass
      ? inputs.travelClass.value
      : "economy";
    const currency = (inputs.currency?.value || "USD").toUpperCase();

    const stopsRaw = inputs.stops ? inputs.stops.value : "any";
    const stops = stopsToApiValue(stopsRaw);

    let adults = parseInt(inputs.adults?.value || "1", 10);
    let children = parseInt(inputs.children?.value || "0", 10);
    let infantsLap = parseInt(inputs.infantsLap?.value || "0", 10);
    let infantsSeat = parseInt(inputs.infantsSeat?.value || "0", 10);
    if (!Number.isFinite(adults) || adults <= 0) adults = 1;
    if (!Number.isFinite(children) || children < 0) children = 0;
    if (!Number.isFinite(infantsLap) || infantsLap < 0) infantsLap = 0;
    if (!Number.isFinite(infantsSeat) || infantsSeat < 0) infantsSeat = 0;

    const interval = inputs.recurringInterval?.value || "daily";
    const timeValue = inputs.recurringTime?.value || "07:00";
    const cronExpression = cronFromIntervalAndTime(interval, timeValue);

    const dynamicDates = !!inputs.recurringDynamicDates?.checked;

    const nameBase =
      (inputs.recurringName?.value || "").trim() ||
      `Search ${parsedRoutes.origins.join(",")}>${parsedRoutes.destinations.join(",")}`;

    const payload = {
      name: nameBase,
      origins: parsedRoutes.origins,
      destinations: parsedRoutes.destinations,
      trip_type: tripType,
      class: travelClass,
      stops,
      adults,
      children,
      infants_lap: infantsLap,
      infants_seat: infantsSeat,
      currency,
      cron_expression: cronExpression,
      dynamic_dates: dynamicDates,
    };

    if (dynamicDates) {
      const daysFromExecution = parseInt(
        inputs.recurringDaysFromExecution?.value || "14",
        10,
      );
      const searchWindowDays = parseInt(
        inputs.recurringSearchWindowDays?.value || "7",
        10,
      );
      const tripLength = parseInt(inputs.recurringTripLength?.value || "7", 10);

      if (Number.isFinite(daysFromExecution) && daysFromExecution >= 0) {
        payload.days_from_execution = daysFromExecution;
      }
      if (Number.isFinite(searchWindowDays) && searchWindowDays >= 1) {
        payload.search_window_days = searchWindowDays;
      }
      if (
        tripType === "round_trip" &&
        Number.isFinite(tripLength) &&
        tripLength > 0
      ) {
        payload.trip_length = tripLength;
      }

      const placeholder = new Date();
      placeholder.setDate(placeholder.getDate() + 30);
      const dateString = placeholder.toISOString().split("T")[0];
      payload.date_start = dateString;
      payload.date_end = dateString;
    } else {
      const dep = inputs.departureDate.value;
      if (!dep)
        throw new Error("Departure date is required for fixed-date schedules.");
      payload.date_start = dep;
      payload.date_end = dep;

      if (tripType === "round_trip") {
        const ret = inputs.returnDate?.value || "";
        if (ret) {
          payload.return_date_start = ret;
          payload.return_date_end = ret;
        }
      }
    }

    if (elements.createRecurringBtn)
      elements.createRecurringBtn.disabled = true;
    if (elements.recurringStatus) {
      elements.recurringStatus.textContent = "Creating scheduled job(s)…";
    }

    const response = await fetch(ENDPOINTS.ADMIN_BULK_JOBS, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      let message = `Failed to create recurring search (${response.status})`;
      try {
        const errorData = await response.json();
        message = errorData?.error || message;
      } catch {}
      throw new Error(message);
    }

    const result = await response.json();
    const count = Array.isArray(result.jobs) ? result.jobs.length : null;
    if (elements.recurringStatus) {
      elements.recurringStatus.textContent =
        count != null
          ? `Created ${count} scheduled job(s).`
          : "Created scheduled job(s).";
    }
    showAlert(
      count != null
        ? `Created ${count} scheduled job(s).`
        : "Created scheduled job(s).",
      "success",
    );
  } catch (error) {
    console.error("Error creating recurring search:", error);
    showAlert(`Error creating recurring search: ${error.message}`, "danger");
    if (elements.recurringStatus) elements.recurringStatus.textContent = "";
  } finally {
    if (elements.createRecurringBtn)
      elements.createRecurringBtn.disabled = false;
  }
}

// Load airports for autocomplete
// Add this function after initSearchPage
async function loadAirports() {
  try {
    const cached = getCachedAirports();
    if (cached) {
      populateAirportDatalists(cached);
      return;
    }

    const response = await fetch(ENDPOINTS.AIRPORTS);
    if (!response.ok) {
      throw new Error("Failed to load airports");
    }

    const airports = await response.json();
    cacheAirports(airports);
    populateAirportDatalists(airports);
  } catch (error) {
    console.error("Error loading airports:", error);
    showAlert("Failed to load airport data", "warning");
  }
}

function getCachedAirports() {
  try {
    const raw = localStorage.getItem(AIRPORT_CACHE_KEY);
    if (!raw) return null;
    const cached = JSON.parse(raw);
    if (!cached?.data || !cached?.cachedAt) return null;
    if (Date.now() - cached.cachedAt > AIRPORT_CACHE_TTL_MS) {
      localStorage.removeItem(AIRPORT_CACHE_KEY);
      return null;
    }
    return cached.data;
  } catch {
    return null;
  }
}

function cacheAirports(airports) {
  try {
    localStorage.setItem(
      AIRPORT_CACHE_KEY,
      JSON.stringify({ cachedAt: Date.now(), data: airports }),
    );
  } catch (error) {
    console.warn("Failed to cache airports:", error);
  }
}

function populateAirportDatalists(airports) {
  if (!inputs.origin || !inputs.destination || !Array.isArray(airports)) {
    return;
  }

  const ensureList = (id) => {
    let list = document.getElementById(id);
    if (!list) {
      list = document.createElement("datalist");
      list.id = id;
      document.body.appendChild(list);
    } else {
      list.innerHTML = "";
    }
    return list;
  };

  const originDatalist = ensureList("originList");
  const destDatalist = ensureList("destinationList");

  airports.forEach((airport) => {
    const label = `${airport.code} - ${airport.name}, ${airport.city}`;
    const originOption = document.createElement("option");
    originOption.value = label;
    originDatalist.appendChild(originOption);

    const destOption = document.createElement("option");
    destOption.value = label;
    destDatalist.appendChild(destOption);
  });

  inputs.origin.setAttribute("list", "originList");
  inputs.destination.setAttribute("list", "destinationList");
}

// Helper to toggle the loading indicator safely
function setLoadingState(isLoading) {
  if (!elements.loadingIndicator) return;
  elements.loadingIndicator.style.display = isLoading ? "flex" : "none";
  elements.loadingIndicator.setAttribute(
    "aria-busy",
    isLoading ? "true" : "false",
  );
  elements.loadingIndicator.setAttribute(
    "aria-hidden",
    isLoading ? "false" : "true",
  );
  if (elements.submitButton) {
    elements.submitButton.disabled = isLoading;
  }
}

function showAlert(message, type = "info") {
  const alertDiv = document.createElement("div");
  alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
  alertDiv.role = "alert";
  alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    `;

  const container = document.querySelector(".container");
  if (container) {
    container.insertBefore(alertDiv, container.firstChild);
  } else {
    document.body.appendChild(alertDiv);
  }

  setTimeout(() => {
    alertDiv.classList.remove("show");
    setTimeout(() => alertDiv.remove(), 150);
  }, 5000);
}

// Handle search form submission
async function handleSearch(event) {
  event.preventDefault();

  setLoadingState(true);
  if (elements.priceGraphCard) elements.priceGraphCard.style.display = "none";
  if (elements.resultsCard) elements.resultsCard.style.display = "none";

  try {
    const originInput = inputs.origin;
    const destinationInput = inputs.destination;
    const departureInput = inputs.departureDate;
    const returnInput = inputs.returnDate;
    const tripTypeInput = inputs.tripType;
    const classInput = inputs.travelClass;
    const stopsInput = inputs.stops;
    const adultsInput = inputs.adults;
    const childrenInput = inputs.children;
    const infantsLapInput = inputs.infantsLap;
    const infantsSeatInput = inputs.infantsSeat;
    const currencyInput = inputs.currency;

    if (!originInput || !destinationInput || !departureInput) {
      throw new Error(
        "Please fill in origin, destination, and departure date.",
      );
    }

    const parsedRoutes = parseRouteInputs(
      originInput.value,
      destinationInput.value,
    );
    const origin = parsedRoutes.origins.join(",");
    const destination = parsedRoutes.destinations.join(",");
    const departureDate = departureInput.value;
    const returnDate = returnInput ? returnInput.value : "";
    const tripType = tripTypeInput ? tripTypeInput.value : "round_trip";
    const travelClass = classInput ? classInput.value : "economy";
    const stops = stopsToApiValue(stopsInput ? stopsInput.value : "any");

    let adults = 1;
    let children = 0;
    let infantsLap = 0;
    let infantsSeat = 0;
    let currency = "USD";

    if (advancedOptionsVisible) {
      adults = parseInt(adultsInput?.value || "1", 10);
      children = parseInt(childrenInput?.value || "0", 10);
      infantsLap = parseInt(infantsLapInput?.value || "0", 10);
      infantsSeat = parseInt(infantsSeatInput?.value || "0", 10);
      currency = (currencyInput?.value || "USD").toUpperCase();
    } else if (adultsInput) {
      adults = parseInt(adultsInput.value || "1", 10);
    }

    const searchData = {
      origin,
      destination,
      departure_date: departureDate,
      trip_type: tripType,
      class: travelClass,
      stops,
      adults,
      children,
      infants_lap: infantsLap,
      infants_seat: infantsSeat,
      currency,
    };

    if (tripType === "round_trip" && returnDate) {
      searchData.return_date = returnDate;
    }

    console.log("Sending search request:", searchData);

    // Send search request
    const response = await fetch(ENDPOINTS.SEARCH, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(searchData),
    });

    console.log("Response status:", response.status);

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Search failed");
    }

    const searchResult = await response.json();
    console.log("Search results:", searchResult);

    // Load price history for the route (single or cheapest route).
    let priceHistoryOrigin = splitAirportList(origin)[0] || "";
    let priceHistoryDestination = splitAirportList(destination)[0] || "";
    if (searchResult?.cheapest?.origin && searchResult?.cheapest?.destination) {
      priceHistoryOrigin = searchResult.cheapest.origin;
      priceHistoryDestination = searchResult.cheapest.destination;
    }
    if (priceHistoryOrigin && priceHistoryDestination) {
      await loadPriceHistory(priceHistoryOrigin, priceHistoryDestination);
    }

    // Display search results
    displaySearchResults(searchResult);
  } catch (error) {
    console.error("Error searching flights:", error);
    showAlert(`Error searching flights: ${error.message}`, "danger");
  } finally {
    setLoadingState(false);
  }
}

// Load price history for a route
async function loadPriceHistory(origin, destination) {
  try {
    const response = await fetch(
      `${ENDPOINTS.PRICE_HISTORY}?origin=${origin}&destination=${destination}`,
    );
    if (!response.ok) throw new Error("Failed to load price history");

    const priceHistory = await response.json();

    // Display price graph
    displayPriceGraph(priceHistory);
  } catch (error) {
    console.error("Error loading price history:", error);
    // Don't show an alert for this non-critical error
  }
}

// Display price graph
function displayPriceGraph(priceHistory) {
  if (!elements.priceGraphCard || !priceHistory) return;

  const history = Array.isArray(priceHistory.history)
    ? priceHistory.history
    : Array.isArray(priceHistory)
      ? priceHistory
      : null;
  if (!history || history.length === 0) {
    elements.priceGraphCard.style.display = "none";
    return;
  }

  elements.priceGraphCard.style.display = "block";

  // Prepare data for chart
  const dates = history.map((item) => item.date);
  const prices = history.map((item) => item.price);

  // Destroy existing chart if it exists
  if (priceChart) {
    priceChart.destroy();
  }

  // Create new chart
  const ctx = elements.priceGraph.getContext("2d");
  priceChart = new Chart(ctx, {
    type: "line",
    data: {
      labels: dates,
      datasets: [
        {
          label: "Price",
          data: prices,
          borderColor: "rgba(75, 192, 192, 1)",
          backgroundColor: "rgba(75, 192, 192, 0.2)",
          tension: 0.1,
          fill: true,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: false,
          title: {
            display: true,
            text: "Price",
          },
        },
        x: {
          title: {
            display: true,
            text: "Date",
          },
        },
      },
    },
  });
}

// Display search results
function displaySearchResults(searchResult) {
  currentRoutes = null;
  currentActiveRouteIndex = 0;
  routePaneByIndex = new Map();

  if (Array.isArray(searchResult?.routes)) {
    displayMultiRouteResults(searchResult);
    return;
  }

  // Show results card
  if (elements.resultsCard) {
    elements.resultsCard.style.display = "block";
  }

  // Clear previous results
  if (elements.flightResults) {
    elements.flightResults.innerHTML = "";
  }

  // Check if we have results
  if (
    !elements.flightResults ||
    !searchResult.offers ||
    searchResult.offers.length === 0
  ) {
    if (elements.flightResults) {
      elements.flightResults.innerHTML =
        '<div class="alert alert-info">No flights found matching your criteria.</div>';
    }
    return;
  }

  currentOffers = searchResult.offers.slice();
  currentSearchParams = searchResult.search_params || null;

  // Sort and display results
  sortFlightResults();
}

function computeCheapestFromRoutes(routes) {
  let best = null;
  for (const route of routes || []) {
    if (!route || route.error || !Array.isArray(route.offers)) continue;
    for (const offer of route.offers) {
      if (!offer || typeof offer.price !== "number") continue;
      if (!best || offer.price < best.offer.price) {
        best = {
          origin: route.origin,
          destination: route.destination,
          offer,
          searchParams: route.search_params || null,
        };
      }
    }
  }
  return best;
}

function displayMultiRouteResults(searchResult) {
  if (elements.resultsCard) {
    elements.resultsCard.style.display = "block";
  }
  if (!elements.flightResults) return;

  elements.flightResults.innerHTML = "";

  currentRoutes = Array.isArray(searchResult.routes) ? searchResult.routes : [];
  const cheapest =
    searchResult?.cheapest?.offer &&
    typeof searchResult?.cheapest?.origin === "string" &&
    typeof searchResult?.cheapest?.destination === "string"
      ? {
          origin: searchResult.cheapest.origin,
          destination: searchResult.cheapest.destination,
          offer: searchResult.cheapest.offer,
          searchParams: null,
        }
      : computeCheapestFromRoutes(currentRoutes);

  if (cheapest?.offer) {
    const summary = document.createElement("div");
    summary.className = "mb-3";
    summary.innerHTML = `
      <div class="text-muted small mb-2">
        Cheapest overall: <strong>${escapeHtml(cheapest.origin)} → ${escapeHtml(cheapest.destination)}</strong>
      </div>
    `;
    summary.appendChild(
      createFlightCard(cheapest.offer, cheapest.searchParams),
    );
    elements.flightResults.appendChild(summary);
  }

  if (!currentRoutes.length) {
    elements.flightResults.innerHTML +=
      '<div class="alert alert-info">No routes returned for this search.</div>';
    return;
  }

  const tabsNav = document.createElement("ul");
  tabsNav.className = "nav nav-tabs";

  const tabsContent = document.createElement("div");
  tabsContent.className =
    "tab-content border border-top-0 rounded-bottom p-3 bg-white";

  currentRoutes.forEach((route, index) => {
    const origin = route?.origin || "";
    const destination = route?.destination || "";
    const tabId = `route-tab-${index}`;
    const paneId = `route-pane-${index}`;

    const li = document.createElement("li");
    li.className = "nav-item";
    li.role = "presentation";

    const btn = document.createElement("button");
    btn.className = `nav-link${index === 0 ? " active" : ""}`;
    btn.type = "button";
    btn.id = tabId;
    btn.setAttribute("aria-controls", paneId);
    btn.setAttribute("aria-selected", index === 0 ? "true" : "false");
    btn.textContent = `${origin}→${destination}`;
    btn.addEventListener("click", () => {
      currentActiveRouteIndex = index;

      tabsNav.querySelectorAll("button.nav-link").forEach((b) => {
        b.classList.remove("active");
        b.setAttribute("aria-selected", "false");
      });
      btn.classList.add("active");
      btn.setAttribute("aria-selected", "true");

      tabsContent.querySelectorAll(".tab-pane").forEach((pane) => {
        pane.classList.remove("show", "active");
      });
      const pane = routePaneByIndex.get(index);
      if (pane) {
        pane.classList.add("show", "active");
      }

      renderActiveRoutePane();
    });

    li.appendChild(btn);
    tabsNav.appendChild(li);

    const pane = document.createElement("div");
    pane.className = `tab-pane fade${index === 0 ? " show active" : ""}`;
    pane.id = paneId;
    pane.setAttribute("role", "tabpanel");
    pane.setAttribute("aria-labelledby", tabId);
    tabsContent.appendChild(pane);
    routePaneByIndex.set(index, pane);
  });

  elements.flightResults.appendChild(tabsNav);
  elements.flightResults.appendChild(tabsContent);

  renderActiveRoutePane();
}

function renderActiveRoutePane() {
  if (!currentRoutes || !routePaneByIndex) return;
  const pane = routePaneByIndex.get(currentActiveRouteIndex);
  if (!pane) return;

  const route = currentRoutes[currentActiveRouteIndex];
  pane.innerHTML = "";

  if (!route || route.error) {
    const message = route?.error || "No data for this route.";
    pane.innerHTML = `<div class="alert alert-warning mb-0">${escapeHtml(message)}</div>`;
    return;
  }

  const offers = Array.isArray(route.offers) ? route.offers.slice() : [];
  if (!offers.length) {
    pane.innerHTML =
      '<div class="alert alert-info mb-0">No flights found matching your criteria.</div>';
    return;
  }

  const sortBy = elements.sortResults ? elements.sortResults.value : "price";
  switch (sortBy) {
    case "price":
      offers.sort((a, b) => (a.price || 0) - (b.price || 0));
      break;
    case "duration":
      offers.sort((a, b) => (a.total_duration || 0) - (b.total_duration || 0));
      break;
    case "departure":
      offers.sort((a, b) => {
        const dateA = new Date(a?.segments?.[0]?.departure_time || 0);
        const dateB = new Date(b?.segments?.[0]?.departure_time || 0);
        return dateA - dateB;
      });
      break;
  }

  const routeParams = route.search_params || currentSearchParams;
  offers.forEach((offer) => {
    pane.appendChild(createFlightCard(offer, routeParams));
  });
}

// Sort flight results based on selected criteria
function sortFlightResults() {
  if (currentRoutes && currentRoutes.length > 0) {
    renderActiveRoutePane();
    return;
  }

  if (!elements.flightResults) return;
  const sortBy = elements.sortResults ? elements.sortResults.value : "price";
  const offers = currentOffers.slice();

  if (!offers || offers.length === 0) return;

  // Sort offers
  switch (sortBy) {
    case "price":
      offers.sort((a, b) => a.price - b.price);
      break;
    case "duration":
      offers.sort((a, b) => a.total_duration - b.total_duration);
      break;
    case "departure":
      offers.sort((a, b) => {
        const dateA = new Date(a.segments[0].departure_time);
        const dateB = new Date(b.segments[0].departure_time);
        return dateA - dateB;
      });
      break;
  }

  // Clear previous results
  elements.flightResults.innerHTML = "";

  // Display sorted offers
  offers.forEach((offer) => {
    const card = createFlightCard(offer);
    elements.flightResults.appendChild(card);
  });
}

// Create a flight card element
function createFlightCard(offer, searchParamsOverride) {
  const card = document.createElement("div");
  card.className = "flight-card";

  // Format departure and arrival times
  const departureSegment = offer.segments[0];
  const effectiveSearchParams = searchParamsOverride || currentSearchParams;
  const tripType = effectiveSearchParams?.trip_type || "round_trip";

  const departureTime = new Date(departureSegment.departure_time);
  const arrivalTime = new Date(departureSegment.arrival_time);

  const formattedDepartureTime = departureTime.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });
  const formattedArrivalTime = arrivalTime.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  // Format duration
  const hours = Math.floor(offer.total_duration / 60);
  const minutes = offer.total_duration % 60;
  const formattedDuration = `${hours}h ${minutes}m`;

  // Use the Google Flights URL from the backend if available, otherwise fall back to client-side generation
  let googleFlightsUrl = offer.google_flights_url;
  if (!googleFlightsUrl) {
    // Fallback to client-side URL generation if backend URL is not available
    googleFlightsUrl = createGoogleFlightsUrl(
      departureSegment.departure_airport,
      departureSegment.arrival_airport,
      departureTime,
      tripType,
    );
  }

  const airlineCode =
    departureSegment.airline_code ||
    extractAirlineCodeFromFlightNumber(departureSegment.flight_number) ||
    "";
  const airlineLogo = airlineCode
    ? `https://www.gstatic.com/flights/airline_logos/70px/${airlineCode}.png`
    : "https://via.placeholder.com/70x40?text=Air";

  const airlineGroupBadges = renderAirlineGroupBadges(offer.airline_groups);
  const currency = offer.currency || effectiveSearchParams?.currency || "";

  // Create card content
  card.innerHTML = `
        <div class="row align-items-center">
            <div class="col-md-2">
                <img src="${airlineLogo}"
                     alt="${airlineCode || "Airline"}" class="airline-logo">
                <div>${escapeHtml(airlineCode || "Unknown")} ${escapeHtml(departureSegment.flight_number || "")}</div>
                <div class="mt-2">${airlineGroupBadges}</div>
            </div>
            <div class="col-md-3">
                <div class="flight-time">${formattedDepartureTime} - ${formattedArrivalTime}</div>
                <div class="flight-duration">${formattedDuration}</div>
                <div>${departureSegment.departure_airport} - ${departureSegment.arrival_airport}</div>
            </div>
            <div class="col-md-3">
                <div>${getStopsLabel(offer.segments)}</div>
                <div>${departureSegment.airplane || "Aircraft information unavailable"}</div>
            </div>
            <div class="col-md-2 text-end">
                <div class="flight-price">${escapeHtml(currency)} ${Number(offer.price || 0).toFixed(2)}</div>
                <button class="btn btn-sm btn-outline-primary mt-2">Select</button>
                <a href="${googleFlightsUrl}" target="_blank" class="btn btn-sm btn-outline-secondary mt-1">
                    <i class="bi bi-google"></i> View on Google Flights
                </a>
            </div>
        </div>
    `;

  return card;
}

function getStopsLabel(segments) {
  if (segments.length === 1) {
    return "Non-stop";
  } else if (segments.length === 2) {
    return "1 stop";
  } else if (segments.length > 2) {
    return `${segments.length - 1} stops`;
  }
  return "";
}

// Create Google Flights URL based on flight details
function createGoogleFlightsUrl(
  origin,
  destination,
  departureDate,
  tripType = "round_trip",
) {
  // Format the date as YYYY-MM-DD
  const formattedDate = departureDate.toISOString().split("T")[0];

  // Construct the Google Flights URL
  // Format: https://www.google.com/travel/flights?q=Flights%20from%20[ORIGIN]%20to%20[DESTINATION]%20on%20[DATE]
  const encodedOrigin = encodeURIComponent(origin);
  const encodedDestination = encodeURIComponent(destination);
  const encodedDate = encodeURIComponent(formattedDate);

  // Add trip type parameter
  const tripTypeParam = tripType === "one_way" ? "&tfs=oneway" : "";

  return `https://www.google.com/travel/flights?q=Flights%20from%20${encodedOrigin}%20to%20${encodedDestination}%20on%20${encodedDate}${tripTypeParam}`;
}

document.addEventListener("DOMContentLoaded", initSearchPage);
