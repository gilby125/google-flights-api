// Flight Search JavaScript

// API endpoints
const API_BASE = "/api";
const ENDPOINTS = {
  SEARCH: `${API_BASE}/search`,
  PRICE_HISTORY_V1: "/api/v1/price-history",
  BULK_SEARCH_V1: "/api/v1/bulk-search",
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
  priceHistoryEmpty: document.getElementById("priceHistoryEmpty"),
  googlePriceGraphCard: document.getElementById("googlePriceGraphCard"),
  googlePriceGraph: document.getElementById("googlePriceGraph"),
  googlePriceGraphError: document.getElementById("googlePriceGraphError"),
  resultsCard: document.getElementById("resultsCard"),
  flightResults: document.getElementById("flightResults"),
  sortResults: document.getElementById("sortResults"),
  submitButton: document.querySelector("#searchForm button[type='submit']"),
  asyncBulkStatus: document.getElementById("asyncBulkStatus"),
  asyncBulkStatusText: document.getElementById("asyncBulkStatusText"),
  asyncBulkStatusCounts: document.getElementById("asyncBulkStatusCounts"),
  asyncBulkProgressBar: document.getElementById("asyncBulkProgressBar"),
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
  includePriceGraph: document.getElementById("includePriceGraph"),
  priceGraphMode: document.getElementById("priceGraphMode"),
  priceGraphWindowDays: document.getElementById("priceGraphWindowDays"),
  priceGraphTripLengthDays: document.getElementById("priceGraphTripLengthDays"),
  debugBatches: document.getElementById("debugBatches"),
  excludeLowCost: document.getElementById("excludeLowCost"),
  includeStar: document.getElementById("includeStar"),
  includeOneworld: document.getElementById("includeOneworld"),
  includeSkyTeam: document.getElementById("includeSkyTeam"),
  googleCarriers: document.getElementById("googleCarriers"),
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
let googlePriceGraphChart = null;
let activeRoutePriceChart = null;
let activeRouteGoogleChart = null;
let currentOffers = [];
let currentSearchParams = null;
let currentRoutes = null;
let currentActiveRouteIndex = 0;
let routePaneByIndex = new Map();
let advancedOptionsVisible =
  !!elements.advancedOptions &&
  getComputedStyle(elements.advancedOptions).display !== "none";
let activeBulkSearch = null;

function parseCarrierTokens(input) {
  const raw = String(input || "")
    .trim()
    .toUpperCase();
  if (!raw) return [];
  return raw
    .split(/[,\s]+/)
    .map((t) => t.trim())
    .filter((t) => /^[A-Z0-9_]+$/.test(t));
}

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

const REGION_ALIAS_TO_TOKEN = {
  AFRICA: "REGION:AFRICA",
  ASIA: "REGION:ASIA",
  CARIBBEAN: "REGION:CARIBBEAN",
  EUROPE: "REGION:EUROPE",
  MIDDLEEAST: "REGION:MIDDLE_EAST",
  MIDEAST: "REGION:MIDDLE_EAST",
  NORTHAMERICA: "REGION:NORTH_AMERICA",
  OCEANIA: "REGION:OCEANIA",
  SOUTHAMERICA: "REGION:SOUTH_AMERICA",
  WORLD: "REGION:WORLD",
  WORLDALL: "REGION:WORLD_ALL",
};

function normalizeRegionAlias(value) {
  const upper = String(value || "")
    .trim()
    .toUpperCase();
  const withoutPrefix = upper.startsWith("REGION:")
    ? upper.slice("REGION:".length)
    : upper;

  // Keep only A-Z0-9 and normalize separators away to allow "NORTH AMERICA" / "NORTH_AMERICA".
  return withoutPrefix.replace(/[^A-Z0-9]+/g, "");
}

function canonicalizeRegionToken(value) {
  const token = String(value || "")
    .trim()
    .toUpperCase();
  if (!token) return "";

  if (token.startsWith("REGION:")) {
    // Accept canonical REGION:* tokens.
    const key = normalizeRegionAlias(token);
    const mapped = REGION_ALIAS_TO_TOKEN[key];
    return mapped || token;
  }

  const key = normalizeRegionAlias(token);
  return REGION_ALIAS_TO_TOKEN[key] || "";
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
    let token = normalizeAirportToken(part);
    if (!token) continue;

    if (!/^[A-Z0-9]{3}$/.test(token)) {
      const regionToken = canonicalizeRegionToken(token);
      if (!regionToken) continue;
      token = regionToken;
    }

    if (seen.has(token)) continue;
    seen.add(token);
    out.push(token);
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
    throw new Error("No valid origin airport/region tokens found.");
  }
  if (!destinations.length) {
    throw new Error("No valid destination airport/region tokens found.");
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

  if (inputs.includePriceGraph) {
    inputs.includePriceGraph.addEventListener(
      "change",
      setPriceGraphFieldVisibility,
    );
  }
  if (inputs.priceGraphMode) {
    inputs.priceGraphMode.addEventListener(
      "change",
      setPriceGraphFieldVisibility,
    );
  }
  setPriceGraphFieldVisibility();

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

function setPriceGraphFieldVisibility() {
  const enabled = !!inputs.includePriceGraph?.checked;
  document.querySelectorAll(".priceGraphOption").forEach((el) => {
    el.style.display = enabled ? "block" : "none";
  });

  const mode = String(inputs.priceGraphMode?.value || "around_date");
  const showOpen = enabled && mode === "open_dates";
  document.querySelectorAll(".priceGraphOptionOpen").forEach((el) => {
    el.style.display = showOpen ? "block" : "none";
  });
}

function addDaysToDateString(dateStr, days) {
  const input = String(dateStr || "").trim();
  if (!input) return "";
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(input);
  if (!match) return "";

  const year = parseInt(match[1], 10);
  const monthIndex = parseInt(match[2], 10) - 1;
  const day = parseInt(match[3], 10);
  const date = new Date(Date.UTC(year, monthIndex, day));
  if (Number.isNaN(date.getTime())) return "";

  date.setUTCDate(date.getUTCDate() + (parseInt(days, 10) || 0));
  const outYear = date.getUTCFullYear();
  const outMonth = String(date.getUTCMonth() + 1).padStart(2, "0");
  const outDay = String(date.getUTCDate()).padStart(2, "0");
  return `${outYear}-${outMonth}-${outDay}`;
}

function toYyyyMmDd(value) {
  if (!value) return "";
  const raw = String(value).trim();
  if (/^\d{4}-\d{2}-\d{2}/.test(raw)) return raw.slice(0, 10);
  const parsed = new Date(raw);
  if (Number.isNaN(parsed.getTime())) return "";
  const year = parsed.getUTCFullYear();
  const month = String(parsed.getUTCMonth() + 1).padStart(2, "0");
  const day = String(parsed.getUTCDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function normalizePriceHistorySeries(priceHistory) {
  const history = Array.isArray(priceHistory?.history)
    ? priceHistory.history
    : Array.isArray(priceHistory)
      ? priceHistory
      : null;
  if (!history || history.length === 0) {
    return { dates: [], prices: [] };
  }

  // Normalize to one point per day (min price) so the chart isn't noisy/duplicated.
  const minByDate = new Map();
  for (const item of history) {
    const date = toYyyyMmDd(item?.date);
    if (!date) continue;
    const price = Number(item?.price);
    if (!Number.isFinite(price) || price <= 0) continue;

    const current = minByDate.get(date);
    if (current == null || price < current) {
      minByDate.set(date, price);
    }
  }

  const dates = Array.from(minByDate.keys()).sort();
  const prices = dates.map((date) => minByDate.get(date));
  return { dates, prices };
}

function renderPriceHistoryChart(canvas, priceHistory) {
  if (!canvas) return null;
  if (typeof ApexCharts !== "function") {
    console.error("ApexCharts is not available on the page.");
    return null;
  }

  const { dates, prices } = normalizePriceHistorySeries(priceHistory);
  if (dates.length === 0) return null;

  canvas.innerHTML = "";

  const formatShortDate = (dateStr) => {
    const raw = String(dateStr || "").trim();
    if (!raw) return "";
    const parsed = new Date(`${raw}T00:00:00Z`);
    if (Number.isNaN(parsed.getTime())) return raw;
    try {
      return new Intl.DateTimeFormat(undefined, {
        month: "numeric",
        day: "numeric",
      }).format(parsed);
    } catch {
      return raw;
    }
  };

  const currency = String(priceHistory?.currency || inputs.currency?.value || "")
    .trim()
    .toUpperCase();
  const formatMoney = (value) => {
    const num = typeof value === "number" ? value : Number(value);
    if (!Number.isFinite(num)) return "";
    if (!currency) return Math.round(value).toString();
    try {
      return new Intl.NumberFormat(undefined, {
        style: "currency",
        currency,
        maximumFractionDigits: 0,
      }).format(num);
    } catch {
      return `${currency} ${Math.round(num)}`;
    }
  };

  const chart = new ApexCharts(canvas, {
    chart: {
      type: "area",
      height: "100%",
      toolbar: { show: false },
      zoom: { enabled: false },
      animations: { enabled: true },
    },
    series: [{ name: "Price", data: prices }],
    stroke: { curve: "smooth", width: 2 },
    fill: {
      type: "gradient",
      gradient: { shadeIntensity: 0.3, opacityFrom: 0.35, opacityTo: 0.05 },
    },
    markers: { size: 3, hover: { size: 5 } },
    dataLabels: { enabled: false },
    xaxis: {
      type: "category",
      categories: dates,
      tickAmount: Math.min(2, dates.length),
      labels: {
        rotate: 0,
        hideOverlappingLabels: true,
        style: { fontSize: "10px" },
        formatter: (value) => {
          const raw = String(value || "").trim();
          if (!raw) return "";
          if (raw === dates[0] || raw === dates[dates.length - 1]) {
            return formatShortDate(raw);
          }
          return "";
        },
      },
    },
    yaxis: {
      labels: {
        formatter: (value) => formatMoney(value),
      },
    },
    tooltip: {
      x: {
        formatter: (value) => String(value || ""),
      },
      y: {
        formatter: (value) => formatMoney(value),
      },
    },
    grid: { borderColor: "rgba(15, 23, 42, 0.12)" },
    colors: ["#20c997"],
  });

  chart.render();
  return chart;
}

function renderGooglePriceGraphChart(canvas, priceGraph, errorEl) {
  if (!canvas) return null;
  if (typeof ApexCharts !== "function") {
    console.error("ApexCharts is not available on the page.");
    return null;
  }

  if (priceGraph?.error) {
    if (errorEl) {
      errorEl.textContent = String(priceGraph.error);
      errorEl.classList.remove("d-none");
    }
    return null;
  }

  if (priceGraph?.filter_warning && errorEl) {
    errorEl.textContent = String(priceGraph.filter_warning);
    errorEl.classList.remove("d-none");
  }

  const points = Array.isArray(priceGraph?.points) ? priceGraph.points : null;
  if (!points || points.length === 0) {
    return null;
  }

  if (errorEl && !priceGraph?.filter_warning) {
    errorEl.classList.add("d-none");
    errorEl.textContent = "";
  }

  canvas.innerHTML = "";

  const formatShortDate = (dateStr) => {
    const raw = String(dateStr || "").trim();
    if (!raw) return "";
    const parsed = new Date(`${raw}T00:00:00Z`);
    if (Number.isNaN(parsed.getTime())) return raw;
    try {
      return new Intl.DateTimeFormat(undefined, {
        month: "numeric",
        day: "numeric",
      }).format(parsed);
    } catch {
      return raw;
    }
  };

  const currency = String(priceGraph?.currency || inputs.currency?.value || "")
    .trim()
    .toUpperCase();
  const formatMoney = (value) => {
    const num = typeof value === "number" ? value : Number(value);
    if (!Number.isFinite(num)) return "";
    if (!currency) return Math.round(value).toString();
    try {
      return new Intl.NumberFormat(undefined, {
        style: "currency",
        currency,
        maximumFractionDigits: 0,
      }).format(num);
    } catch {
      return `${currency} ${Math.round(num)}`;
    }
  };

  const categories = points.map((p) => String(p?.departure_date || "").trim());
  const prices = points.map((p) => {
    const value = Number(p?.price);
    return Number.isFinite(value) && value > 0 ? value : null;
  });
  const urls = points.map((p) => String(p?.google_flights_url || "").trim());

  const chartTitle = `${priceGraph?.origin || ""} → ${priceGraph?.destination || ""}`.trim();

  const options = {
    chart: {
      type: "area",
      height: "100%",
      toolbar: { show: false },
      zoom: { enabled: false },
      animations: { enabled: true },
      events: {
        dataPointMouseEnter: () => {
          canvas.style.cursor = "pointer";
        },
        dataPointMouseLeave: () => {
          canvas.style.cursor = "default";
        },
        dataPointSelection: (_event, _chartContext, config) => {
          const index = config?.dataPointIndex;
          if (typeof index !== "number" || index < 0) return;
          const url = urls?.[index];
          if (url) {
            window.open(url, "_blank", "noopener");
          }
        },
      },
    },
    series: [{ name: "Price", data: prices }],
    stroke: { curve: "smooth", width: 2 },
    fill: {
      type: "gradient",
      gradient: { shadeIntensity: 0.25, opacityFrom: 0.35, opacityTo: 0.05 },
    },
    markers: { size: 3, hover: { size: 6 } },
    dataLabels: { enabled: false },
    xaxis: {
      type: "category",
      categories,
      tickAmount: Math.min(2, categories.length),
      labels: {
        rotate: 0,
        hideOverlappingLabels: true,
        style: { fontSize: "10px" },
        formatter: (value) => {
          const raw = String(value || "").trim();
          if (!raw) return "";
          if (raw === categories[0] || raw === categories[categories.length - 1]) {
            return formatShortDate(raw);
          }
          return "";
        },
      },
    },
    yaxis: {
      labels: {
        formatter: (value) => formatMoney(value),
      },
    },
    tooltip: {
      x: {
        formatter: (value) => String(value || ""),
      },
      y: {
        formatter: (value) => formatMoney(value),
      },
    },
    grid: { borderColor: "rgba(15, 23, 42, 0.12)" },
    colors: ["#0d6efd"],
  };
  if (chartTitle) {
    options.title = { text: chartTitle, style: { fontSize: "13px" } };
  }

  const chart = new ApexCharts(canvas, options);

  chart.render();
  return chart;
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
  activeBulkSearch = null;
  if (elements.priceGraphCard) elements.priceGraphCard.style.display = "none";
  if (elements.priceHistoryEmpty) {
    elements.priceHistoryEmpty.classList.add("d-none");
    elements.priceHistoryEmpty.textContent =
      "No price history collected for this route yet.";
  }
  if (elements.googlePriceGraphCard)
    elements.googlePriceGraphCard.style.display = "none";
  if (elements.googlePriceGraphError) {
    elements.googlePriceGraphError.classList.add("d-none");
    elements.googlePriceGraphError.textContent = "";
  }
  if (elements.resultsCard) elements.resultsCard.style.display = "none";
  if (elements.asyncBulkStatus) elements.asyncBulkStatus.classList.add("d-none");

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

    const carrierTokens = parseCarrierTokens(inputs.googleCarriers?.value);
    if (carrierTokens.length > 0) {
      searchData.carriers = carrierTokens;
    }

    if (inputs.debugBatches?.checked) {
      searchData.debug_batches = true;
    }

    const excludeGroups = [];
    if (inputs.excludeLowCost?.checked) {
      excludeGroups.push("GROUP:LOW_COST");
    }
    if (excludeGroups.length) {
      searchData.exclude_airline_groups = excludeGroups;
    }

    const includeGroups = [];
    if (inputs.includeStar?.checked) includeGroups.push("GROUP:STAR_ALLIANCE");
    if (inputs.includeOneworld?.checked) includeGroups.push("GROUP:ONEWORLD");
    if (inputs.includeSkyTeam?.checked) includeGroups.push("GROUP:SKYTEAM");
    if (includeGroups.length) {
      searchData.include_airline_groups = includeGroups;
    }

    if (tripType === "round_trip" && returnDate) {
      searchData.return_date = returnDate;
    }

    const includePriceGraph = !!inputs.includePriceGraph?.checked;
    if (includePriceGraph) {
      searchData.include_price_graph = true;

      const windowDays = parseInt(
        inputs.priceGraphWindowDays?.value || "30",
        10,
      );
      if (Number.isFinite(windowDays) && windowDays > 0) {
        searchData.price_graph_window_days = windowDays;
      }

      const mode = String(inputs.priceGraphMode?.value || "around_date");
      if (mode === "open_dates") {
        const rangeStart = departureDate;
        const rangeEnd = addDaysToDateString(
          departureDate,
          windowDays || 30,
        );
        if (rangeStart && rangeEnd) {
          searchData.price_graph_departure_date_from = rangeStart;
          searchData.price_graph_departure_date_to = rangeEnd;
        }

        let tripLengthDays = 0;
        if (tripType === "round_trip" && returnDate) {
          const dep = new Date(`${departureDate}T00:00:00Z`);
          const ret = new Date(`${returnDate}T00:00:00Z`);
          tripLengthDays = Math.round(
            (ret - dep) / (24 * 60 * 60 * 1000),
          );
        }
        if (!tripLengthDays) {
          tripLengthDays = parseInt(
            inputs.priceGraphTripLengthDays?.value || "7",
            10,
          );
        }
        if (Number.isFinite(tripLengthDays) && tripLengthDays >= 0) {
          searchData.price_graph_trip_length_days = tripLengthDays;
        }
      }
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
      let errorData = null;
      try {
        errorData = await response.json();
      } catch {}

      const message = errorData?.error || `Search failed (${response.status})`;
      if (isDirectSearchTooLargeError(message)) {
        await startBulkSearchFallback(parsedRoutes, searchData, message);
        return;
      }

      throw new Error(message);
    }

    const searchResult = await response.json();
    console.log("Search results:", searchResult);

    await applySearchResultToPage(searchResult, origin, destination);
  } catch (error) {
    console.error("Error searching flights:", error);
    showAlert(`Error searching flights: ${error.message}`, "danger");
  } finally {
    setLoadingState(false);
  }
}

function isDirectSearchTooLargeError(message) {
  const msg = String(message || "").toLowerCase();
  return (
    msg.includes("too many airports after expansion") ||
    msg.includes("too many batched requests")
  );
}

function updateAsyncBulkStatus(status) {
  if (
    !elements.asyncBulkStatus ||
    !elements.asyncBulkStatusText ||
    !elements.asyncBulkStatusCounts ||
    !elements.asyncBulkProgressBar
  ) {
    return;
  }

  const total = Number(status?.total_searches) || 0;
  const completed = Number(status?.completed) || 0;
  const state = String(status?.status || "").trim() || "queued";
  const pct = total > 0 ? Math.min(100, Math.round((completed / total) * 100)) : 0;

  elements.asyncBulkStatus.classList.remove("d-none");
  elements.asyncBulkStatusText.textContent =
    state === "completed" || state === "completed_with_errors"
      ? "Bulk search completed."
      : state === "failed"
        ? "Bulk search failed."
        : "Bulk search running in background…";
  elements.asyncBulkStatusCounts.textContent =
    total > 0 ? `${completed}/${total} routes processed` : "";

  elements.asyncBulkProgressBar.style.width = `${pct}%`;
  elements.asyncBulkProgressBar.setAttribute("aria-valuenow", String(pct));
  elements.asyncBulkProgressBar.textContent = pct > 0 ? `${pct}%` : "";

  elements.asyncBulkProgressBar.classList.toggle(
    "bg-success",
    state === "completed" || state === "completed_with_errors",
  );
  elements.asyncBulkProgressBar.classList.toggle("bg-danger", state === "failed");
}

async function startBulkSearchFallback(parsedRoutes, baseSearchData, reasonMessage) {
  const dep = toYyyyMmDd(baseSearchData?.departure_date);
  const ret = toYyyyMmDd(baseSearchData?.return_date);
  const tripType = String(baseSearchData?.trip_type || "round_trip");

  const payload = {
    origins: parsedRoutes.origins,
    destinations: parsedRoutes.destinations,
    departure_date_from: dep,
    departure_date_to: dep,
    adults: baseSearchData?.adults || 1,
    children: baseSearchData?.children || 0,
    infants_lap: baseSearchData?.infants_lap || 0,
    infants_seat: baseSearchData?.infants_seat || 0,
    trip_type: tripType,
    class: baseSearchData?.class || "economy",
    stops: baseSearchData?.stops || "any",
    currency: String(baseSearchData?.currency || "USD").toUpperCase(),
  };
  if (Array.isArray(baseSearchData?.carriers) && baseSearchData.carriers.length > 0) {
    payload.carriers = baseSearchData.carriers;
  }
  if (Array.isArray(baseSearchData?.include_airline_groups)) {
    payload.include_airline_groups = baseSearchData.include_airline_groups;
  }
  if (Array.isArray(baseSearchData?.exclude_airline_groups)) {
    payload.exclude_airline_groups = baseSearchData.exclude_airline_groups;
  }

  if (tripType === "round_trip") {
    const depDate = new Date(`${dep}T00:00:00Z`);
    const retDate = new Date(`${ret}T00:00:00Z`);
    const diffDays = Math.round((retDate - depDate) / (24 * 60 * 60 * 1000));
    if (Number.isFinite(diffDays) && diffDays > 0) {
      payload.trip_length = diffDays;
    } else {
      throw new Error("Round trip bulk search requires a valid return date.");
    }
  }

  showAlert(
    `${escapeHtml(reasonMessage)} Switching to async bulk search (same page).`,
    "info",
  );

  const response = await fetch(ENDPOINTS.BULK_SEARCH_V1, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    let message = `Failed to start bulk search (${response.status})`;
    try {
      const err = await response.json();
      message = err?.error || message;
    } catch {}
    throw new Error(message);
  }

  const created = await response.json();
  const bulkSearchId = created?.bulk_search_id;
  if (!bulkSearchId) {
    throw new Error("Bulk search did not return a bulk_search_id.");
  }

  activeBulkSearch = {
    bulkSearchId,
    baseSearchData,
    lastResultsKey: "",
    timer: null,
  };

  await pollBulkSearchOnce();
  if (activeBulkSearch?.timer) clearInterval(activeBulkSearch.timer);
  activeBulkSearch.timer = setInterval(pollBulkSearchOnce, 2000);
}

async function pollBulkSearchOnce() {
  if (!activeBulkSearch) return;

  const bulkSearchId = activeBulkSearch.bulkSearchId;
  const response = await fetch(
    `${ENDPOINTS.BULK_SEARCH_V1}/${encodeURIComponent(bulkSearchId)}?page=1&per_page=60`,
  );
  if (!response.ok) return;

  const status = await response.json();
  updateAsyncBulkStatus(status);

  const results = Array.isArray(status?.results) ? status.results : [];
  const key = JSON.stringify(
    results.map((r) => [r?.origin, r?.destination, r?.price, r?.departure_date, r?.return_date]),
  );
  if (key !== activeBulkSearch.lastResultsKey) {
    activeBulkSearch.lastResultsKey = key;
    applyBulkSearchResultsToPage(status);
  }

  const state = String(status?.status || "");
  if (state === "completed" || state === "completed_with_errors" || state === "failed") {
    if (activeBulkSearch?.timer) clearInterval(activeBulkSearch.timer);
    if (state === "failed") {
      showAlert("Bulk search failed. Try narrowing your inputs.", "danger");
    }
  }
}

function applyBulkSearchResultsToPage(status) {
  const results = Array.isArray(status?.results) ? status.results : [];
  const base = activeBulkSearch?.baseSearchData || {};

  const routes = results
    .map((row) => {
      const origin = String(row?.origin || "").trim().toUpperCase();
      const destination = String(row?.destination || "").trim().toUpperCase();
      const price = Number(row?.price);
      const currency = String(row?.currency || base.currency || "USD").toUpperCase();
      const departureDate = toYyyyMmDd(row?.departure_date || base.departure_date);
      const returnDate = toYyyyMmDd(row?.return_date || base.return_date);

      if (!origin || !destination) return null;
      return {
        origin,
        destination,
        offers: null,
        loading: false,
        error: null,
        summary: {
          price: Number.isFinite(price) ? price : null,
          currency,
          departure_date: departureDate,
          return_date: returnDate,
          airline_code: row?.airline_code || null,
          duration: row?.duration || null,
        },
        // Keep base params; per-route dates are read from summary when lazy-loading.
        search_params: base,
      };
    })
    .filter(Boolean);

  const pseudoSearchResult = {
    routes,
    search_params: base,
  };

  displaySearchResults(pseudoSearchResult);
}

async function applySearchResultToPage(searchResult, originRaw, destinationRaw) {
  displayGooglePriceGraph(searchResult?.price_graph);

  // Load price history for the route (single or cheapest route).
  let priceHistoryOrigin = splitAirportList(originRaw)[0] || "";
  let priceHistoryDestination = splitAirportList(destinationRaw)[0] || "";
  if (searchResult?.cheapest?.origin && searchResult?.cheapest?.destination) {
    priceHistoryOrigin = searchResult.cheapest.origin;
    priceHistoryDestination = searchResult.cheapest.destination;
  }
  const isIata = (value) => /^[A-Z0-9]{3}$/.test(String(value || "").trim());
  if (isIata(priceHistoryOrigin) && isIata(priceHistoryDestination)) {
    const params = searchResult?.search_params || {};
    await loadPriceHistory(priceHistoryOrigin, priceHistoryDestination, {
      includeGroups: params?.include_airline_groups,
      excludeGroups: params?.exclude_airline_groups,
    });
  }

  displaySearchResults(searchResult);
}

// Load price history for a route
async function loadPriceHistory(origin, destination, filters = null) {
  try {
    const originEncoded = encodeURIComponent(origin);
    const destinationEncoded = encodeURIComponent(destination);

    const query = new URLSearchParams();
    if (Array.isArray(filters?.includeGroups)) {
      for (const token of filters.includeGroups) {
        if (!token) continue;
        query.append("include_airline_groups", token);
      }
    }
    if (Array.isArray(filters?.excludeGroups)) {
      for (const token of filters.excludeGroups) {
        if (!token) continue;
        query.append("exclude_airline_groups", token);
      }
    }
    const qs = query.toString();

    const response = await fetch(
      `${ENDPOINTS.PRICE_HISTORY_V1}/${originEncoded}/${destinationEncoded}${qs ? `?${qs}` : ""}`,
    );

    if (!response.ok) throw new Error(`Price history unavailable (${response.status})`);

    const priceHistory = await response.json();
    displayPriceGraph(priceHistory);
  } catch (error) {
    console.error("Error loading price history:", error);
    if (elements.priceGraphCard) elements.priceGraphCard.style.display = "block";
    if (elements.priceHistoryEmpty) {
      elements.priceHistoryEmpty.textContent =
        "No price history collected for this route yet.";
      elements.priceHistoryEmpty.classList.remove("d-none");
    }
  }
}

// Display price graph
function displayPriceGraph(priceHistory) {
  if (!elements.priceGraphCard || !priceHistory) return;

  if (elements.priceHistoryEmpty) {
    elements.priceHistoryEmpty.classList.add("d-none");
    elements.priceHistoryEmpty.textContent =
      "No price history collected for this route yet.";
  }

  const { dates, prices } = normalizePriceHistorySeries(priceHistory);
  if (dates.length === 0) {
    elements.priceGraphCard.style.display = "block";
    if (elements.priceHistoryEmpty) {
      elements.priceHistoryEmpty.classList.remove("d-none");
    }
    return;
  }

  elements.priceGraphCard.style.display = "block";

  // Destroy existing chart if it exists
  if (priceChart) {
    priceChart.destroy();
  }

  priceChart = renderPriceHistoryChart(elements.priceGraph, priceHistory);
}

function displayGooglePriceGraph(priceGraph) {
  if (!elements.googlePriceGraphCard || !elements.googlePriceGraph) return;

  const points = Array.isArray(priceGraph?.points) ? priceGraph.points : null;
  if (!points || points.length === 0 || priceGraph?.error) {
    elements.googlePriceGraphCard.style.display = "none";
    return;
  }

  if (elements.googlePriceGraphError) {
    elements.googlePriceGraphError.classList.add("d-none");
    elements.googlePriceGraphError.textContent = "";
  }

  elements.googlePriceGraphCard.style.display = "block";

  if (googlePriceGraphChart) {
    googlePriceGraphChart.destroy();
  }

  googlePriceGraphChart = renderGooglePriceGraphChart(
    elements.googlePriceGraph,
    priceGraph,
    elements.googlePriceGraphError,
  );
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
      if (
        !offer ||
        typeof offer.price !== "number" ||
        !Number.isFinite(offer.price) ||
        offer.price <= 0
      ) {
        continue;
      }
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

  // Multi-route: render graphs per-route in the active tab, not the global cards.
  if (elements.priceGraphCard) elements.priceGraphCard.style.display = "none";
  if (elements.googlePriceGraphCard)
    elements.googlePriceGraphCard.style.display = "none";

  elements.flightResults.innerHTML = "";

  const batchSummary = searchResult?.batch_summary || null;
  if (batchSummary && typeof batchSummary === "object") {
    const totalBatches = Number(batchSummary.total_batches) || 0;
    const failedBatches = Number(batchSummary.failed_batches) || 0;
    const emptyBatches = Number(batchSummary.empty_batches) || 0;
    if (totalBatches > 0) {
      const alert = document.createElement("div");
      alert.className =
        failedBatches > 0
          ? "alert alert-warning mb-3"
          : emptyBatches > 0
            ? "alert alert-info mb-3"
            : "alert alert-success mb-3";
      alert.innerHTML = `
        <div class="fw-semibold mb-1">Batch diagnostics</div>
        <div class="small">
          ${totalBatches} batches • ${failedBatches} failed • ${emptyBatches} returned 0 offers.
          ${
            inputs.debugBatches?.checked
              ? "See console for batch_debug."
              : "Enable “Debug multi-route batches” to include per-batch details."
          }
        </div>
      `;
      elements.flightResults.appendChild(alert);
    }

    if (inputs.debugBatches?.checked && Array.isArray(searchResult?.batch_debug)) {
      console.log("Batch debug:", searchResult.batch_debug);
    }
  }

  const routeMinPriceForSort = (route) => {
    if (!route) return Number.POSITIVE_INFINITY;
    if (Array.isArray(route.offers) && route.offers.length > 0) {
      let min = Number.POSITIVE_INFINITY;
      for (const offer of route.offers) {
        const price =
          typeof offer?.price === "number" && Number.isFinite(offer.price)
            ? offer.price
            : null;
        if (price != null && price > 0) {
          min = Math.min(min, price);
        }
      }
      if (Number.isFinite(min) && min > 0) return min;
    }
    const summaryPrice =
      typeof route?.summary?.price === "number" && Number.isFinite(route.summary.price)
        ? route.summary.price
        : null;
    if (summaryPrice != null && summaryPrice > 0) return summaryPrice;
    return Number.POSITIVE_INFINITY;
  };

  currentRoutes = Array.isArray(searchResult.routes) ? searchResult.routes : [];
  currentRoutes.sort((a, b) => {
    const priceA = routeMinPriceForSort(a);
    const priceB = routeMinPriceForSort(b);
    if (priceA !== priceB) return priceA - priceB;
    const originA = String(a?.origin || "");
    const originB = String(b?.origin || "");
    if (originA !== originB) return originA.localeCompare(originB);
    const destA = String(a?.destination || "");
    const destB = String(b?.destination || "");
    return destA.localeCompare(destB);
  });
  let cheapest = null;
  if (
    searchResult?.cheapest?.offer &&
    typeof searchResult?.cheapest?.origin === "string" &&
    typeof searchResult?.cheapest?.destination === "string" &&
    typeof searchResult?.cheapest?.offer?.price === "number" &&
    Number.isFinite(searchResult.cheapest.offer.price) &&
    searchResult.cheapest.offer.price > 0
  ) {
    cheapest = {
      origin: searchResult.cheapest.origin,
      destination: searchResult.cheapest.destination,
      offer: searchResult.cheapest.offer,
      searchParams: null,
    };
  } else {
    cheapest = computeCheapestFromRoutes(currentRoutes);
  }

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
  } else if (activeBulkSearch) {
    const note = document.createElement("div");
    note.className = "alert alert-info mb-3";
    note.textContent =
      "Select a route tab to load full offers. Results are arriving as the bulk search runs.";
    elements.flightResults.appendChild(note);
  }

  if (!currentRoutes.length) {
    elements.flightResults.innerHTML +=
      '<div class="alert alert-info">No flights found matching your criteria.</div>';
    return;
  }

  const tabsNav = document.createElement("ul");
  tabsNav.className = "nav nav-tabs";

  const tabsContent = document.createElement("div");
  tabsContent.className =
    "tab-content border border-top-0 rounded-bottom p-3 bg-white";

  const formatCurrencyShort = (currency, amount) => {
    const code = String(currency || "")
      .trim()
      .toUpperCase();
    if (!code || typeof amount !== "number" || !Number.isFinite(amount)) {
      return "";
    }
    try {
      return new Intl.NumberFormat(undefined, {
        style: "currency",
        currency: code,
        maximumFractionDigits: 0,
      }).format(amount);
    } catch {
      return `${code} ${Math.round(amount)}`;
    }
  };

  currentRoutes.forEach((route, index) => {
    const origin = route?.origin || "";
    const destination = route?.destination || "";
    const tabId = `route-tab-${index}`;
    const paneId = `route-pane-${index}`;
    const routeCurrency =
      route?.search_params?.currency || currentSearchParams?.currency || "USD";
    const routeMinPrice = Array.isArray(route?.offers)
      ? route.offers.reduce((min, offer) => {
          const price =
            typeof offer?.price === "number" &&
            Number.isFinite(offer.price) &&
            offer.price > 0
              ? offer.price
              : null;
          if (price == null) return min;
          if (min == null) return price;
          return Math.min(min, price);
        }, null)
      : typeof route?.summary?.price === "number" && Number.isFinite(route.summary.price)
        ? route.summary.price
        : null;
    const priceSuffix =
      typeof routeMinPrice === "number" && Number.isFinite(routeMinPrice)
        ? ` ${formatCurrencyShort(routeCurrency, routeMinPrice)}`
        : "";

    const li = document.createElement("li");
    li.className = "nav-item";
    li.role = "presentation";

    const btn = document.createElement("button");
    btn.className = `nav-link${index === 0 ? " active" : ""}`;
    btn.type = "button";
    btn.id = tabId;
    btn.setAttribute("aria-controls", paneId);
    btn.setAttribute("aria-selected", index === 0 ? "true" : "false");
    btn.textContent = `${origin}→${destination}${priceSuffix}`;
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
      ensureRouteGraphsLoaded(index);
      ensureRouteLoaded(index);
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

  if (activeBulkSearch && currentRoutes.length > 0) {
    // Auto-load the first route's offers so the page shows real itineraries quickly.
    ensureRouteLoaded(0);
    ensureRouteGraphsLoaded(0);
  }
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

  if (route.loading) {
    pane.innerHTML = `
      <div class="d-flex align-items-center gap-2 text-muted">
        <div class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></div>
        <div>Loading offers for ${escapeHtml(route.origin)} → ${escapeHtml(route.destination)}…</div>
      </div>
    `;
    return;
  }

  if (!Array.isArray(route.offers)) {
    const summary = route.summary || {};
    const priceText =
      typeof summary.price === "number" && Number.isFinite(summary.price)
        ? `${escapeHtml(summary.currency || "")} ${Math.round(summary.price)}`
        : "Price unknown";
    const dateText = summary.departure_date
      ? summary.return_date
        ? `${escapeHtml(summary.departure_date)} → ${escapeHtml(summary.return_date)}`
        : `${escapeHtml(summary.departure_date)}`
      : "";

    const container = document.createElement("div");
    container.innerHTML = `
      <div class="mb-2 text-muted small">Summary</div>
      <div class="mb-3">
        <div><strong>${escapeHtml(route.origin)} → ${escapeHtml(route.destination)}</strong></div>
        <div class="text-muted small">${escapeHtml(dateText)}</div>
        <div class="mt-1"><strong>${priceText}</strong></div>
      </div>
      <button type="button" class="btn btn-primary btn-sm" id="loadOffersBtn">
        Load offers for this route
      </button>
    `;

    pane.appendChild(container);
    const btn = container.querySelector("#loadOffersBtn");
    if (btn) {
      btn.addEventListener("click", () => ensureRouteLoaded(currentActiveRouteIndex));
    }
    return;
  }

  const offers = route.offers.slice();
  if (!offers.length) {
    pane.innerHTML =
      '<div class="alert alert-info mb-0">No flights found matching your criteria.</div>';
    return;
  }

  // Per-route graphs (loaded on tab click)
  const graphsRow = document.createElement("div");
  graphsRow.className = "row g-3 mb-3";
  graphsRow.innerHTML = `
    <div class="col-12 col-lg-6">
      <div class="card h-100">
        <div class="card-header bg-light">
          <i class="bi bi-graph-up-arrow me-2 text-primary"></i>
          Price History
        </div>
        <div class="card-body">
          <div class="routePriceHistoryEmpty alert alert-info d-none mb-3" role="alert">
            No price history collected for this route yet.
          </div>
          <div class="chart-container" style="height: 220px;">
            <div class="routePriceHistoryChart apex-chart" role="img" aria-label="Flight price history"></div>
          </div>
        </div>
      </div>
    </div>
    <div class="col-12 col-lg-6">
      <div class="card h-100">
        <div class="card-header bg-light">
          <i class="bi bi-graph-up me-2 text-primary"></i>
          Google Price Graph
        </div>
        <div class="card-body">
          <div class="routeGooglePriceGraphError alert alert-warning d-none mb-3" role="alert"></div>
          <div class="routeGooglePriceGraphEmpty alert alert-info d-none mb-3" role="alert">
            Enable “Show Google price graph” and re-load this route to fetch the calendar graph.
          </div>
          <div class="chart-container" style="height: 220px;">
            <div class="routeGooglePriceGraphChart apex-chart" role="img" aria-label="Google price graph"></div>
          </div>
          <div class="text-muted small mt-2">
            Tip: click a point to open it in Google Flights.
          </div>
        </div>
      </div>
    </div>
  `;

  pane.appendChild(graphsRow);

  const priceHistoryCanvas = graphsRow.querySelector(".routePriceHistoryChart");
  const priceHistoryEmpty = graphsRow.querySelector(".routePriceHistoryEmpty");
  const googleCanvas = graphsRow.querySelector(".routeGooglePriceGraphChart");
  const googleEmpty = graphsRow.querySelector(".routeGooglePriceGraphEmpty");
  const googleError = graphsRow.querySelector(".routeGooglePriceGraphError");

  if (activeRoutePriceChart) {
    activeRoutePriceChart.destroy();
    activeRoutePriceChart = null;
  }
  if (activeRouteGoogleChart) {
    activeRouteGoogleChart.destroy();
    activeRouteGoogleChart = null;
  }

  if (route.priceHistoryLoading) {
    if (priceHistoryEmpty) {
      priceHistoryEmpty.textContent = "Loading price history…";
      priceHistoryEmpty.classList.remove("d-none");
    }
  } else {
    const chart = renderPriceHistoryChart(priceHistoryCanvas, route.priceHistory);
    if (!chart && priceHistoryEmpty) {
      priceHistoryEmpty.classList.remove("d-none");
    }
    activeRoutePriceChart = chart;
  }

  const includeGoogleGraph =
    !!route?.search_params?.include_price_graph ||
    !!currentSearchParams?.include_price_graph ||
    !!inputs.includePriceGraph?.checked;

  if (!includeGoogleGraph) {
    if (googleEmpty) googleEmpty.classList.remove("d-none");
  } else if (route.googlePriceGraphLoading) {
    if (googleEmpty) {
      googleEmpty.textContent = "Loading Google price graph…";
      googleEmpty.classList.remove("d-none");
    }
  } else {
    const chart = renderGooglePriceGraphChart(
      googleCanvas,
      route.google_price_graph,
      googleError,
    );
    if (!chart && googleEmpty) {
      googleEmpty.textContent =
        route.google_price_graph?.error
          ? "Google price graph failed to load."
          : "No Google price graph data returned for this route.";
      googleEmpty.classList.remove("d-none");
    }
    activeRouteGoogleChart = chart;
  }

  const routeCurrency =
    route?.search_params?.currency || currentSearchParams?.currency || "";
  const routeMinPrice = offers.reduce((min, offer) => {
    const price =
      typeof offer?.price === "number" &&
      Number.isFinite(offer.price) &&
      offer.price > 0
        ? offer.price
        : null;
    if (price == null) return min;
    if (min == null) return price;
    return Math.min(min, price);
  }, null);

  const sortBy = elements.sortResults ? elements.sortResults.value : "price";
  switch (sortBy) {
    case "price":
      offers.sort((a, b) => {
        const priceA =
          typeof a?.price === "number" &&
          Number.isFinite(a.price) &&
          a.price > 0
            ? a.price
            : Number.POSITIVE_INFINITY;
        const priceB =
          typeof b?.price === "number" &&
          Number.isFinite(b.price) &&
          b.price > 0
            ? b.price
            : Number.POSITIVE_INFINITY;
        return priceA - priceB;
      });
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
    pane.appendChild(
      createFlightCard(offer, routeParams, { routeCurrency, routeMinPrice }),
    );
  });
}

async function ensureRouteLoaded(routeIndex) {
  if (!activeBulkSearch || !Array.isArray(currentRoutes)) return;
  if (routeIndex < 0 || routeIndex >= currentRoutes.length) return;

  const route = currentRoutes[routeIndex];
  if (!route || route.loading || Array.isArray(route.offers)) return;

  const base = activeBulkSearch.baseSearchData || {};
  const summary = route.summary || {};
  const departureDate = toYyyyMmDd(summary.departure_date || base.departure_date);
  const returnDate = toYyyyMmDd(summary.return_date || base.return_date);

  route.loading = true;
  renderActiveRoutePane();

  try {
    const payload = {
      origin: route.origin,
      destination: route.destination,
      departure_date: departureDate,
      return_date: returnDate,
      trip_type: base.trip_type,
      class: base.class,
      stops: base.stops,
      adults: base.adults,
      children: base.children,
      infants_lap: base.infants_lap,
      infants_seat: base.infants_seat,
      currency: base.currency,
    };
    if (Array.isArray(base.carriers) && base.carriers.length > 0) {
      payload.carriers = base.carriers;
    }
    if (Array.isArray(base.include_airline_groups)) {
      payload.include_airline_groups = base.include_airline_groups;
    }
    if (Array.isArray(base.exclude_airline_groups)) {
      payload.exclude_airline_groups = base.exclude_airline_groups;
    }
    if (base.debug_batches) {
      payload.debug_batches = true;
    }

    if (base.include_price_graph) {
      payload.include_price_graph = true;
      if (base.price_graph_window_days) {
        payload.price_graph_window_days = base.price_graph_window_days;
      }
      if (base.price_graph_departure_date_from) {
        payload.price_graph_departure_date_from = base.price_graph_departure_date_from;
      }
      if (base.price_graph_departure_date_to) {
        payload.price_graph_departure_date_to = base.price_graph_departure_date_to;
      }
      if (base.price_graph_trip_length_days != null) {
        payload.price_graph_trip_length_days = base.price_graph_trip_length_days;
      }
    }

    const response = await fetch(ENDPOINTS.SEARCH, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      let msg = `Failed to load offers (${response.status})`;
      try {
        const err = await response.json();
        msg = err?.error || msg;
      } catch {}
      throw new Error(msg);
    }

    const data = await response.json();
    route.offers = Array.isArray(data?.offers) ? data.offers : [];
    route.search_params = data?.search_params || route.search_params;
    route.loading = false;
    route.error = null;
    route.google_price_graph = data?.price_graph || null;

    if (routeIndex === currentActiveRouteIndex) {
      renderActiveRoutePane();
    }

    await ensureRouteGraphsLoaded(routeIndex);
  } catch (error) {
    route.loading = false;
    route.error = error.message || String(error);
    if (routeIndex === currentActiveRouteIndex) {
      renderActiveRoutePane();
    }
  }
}

async function ensureRouteGraphsLoaded(routeIndex) {
  if (!Array.isArray(currentRoutes)) return;
  if (routeIndex < 0 || routeIndex >= currentRoutes.length) return;

  const route = currentRoutes[routeIndex];
  if (!route || route.priceHistoryLoading || route.googlePriceGraphLoading) return;

  const isIata = (value) => /^[A-Z0-9]{3}$/.test(String(value || "").trim());
  if (!isIata(route.origin) || !isIata(route.destination)) return;

  const base = route.search_params || currentSearchParams || activeBulkSearch?.baseSearchData || {};

  const includeGoogleGraph =
    !!route?.search_params?.include_price_graph ||
    !!currentSearchParams?.include_price_graph ||
    !!inputs.includePriceGraph?.checked;

  // Price history (Neo4j)
  if (route.priceHistory === undefined) {
    route.priceHistoryLoading = true;
    if (routeIndex === currentActiveRouteIndex) renderActiveRoutePane();
    try {
      const query = new URLSearchParams();
      if (Array.isArray(base.include_airline_groups)) {
        for (const token of base.include_airline_groups) {
          if (!token) continue;
          query.append("include_airline_groups", token);
        }
      }
      if (Array.isArray(base.exclude_airline_groups)) {
        for (const token of base.exclude_airline_groups) {
          if (!token) continue;
          query.append("exclude_airline_groups", token);
        }
      }
      const qs = query.toString();

      const resp = await fetch(
        `${ENDPOINTS.PRICE_HISTORY_V1}/${encodeURIComponent(route.origin)}/${encodeURIComponent(route.destination)}${qs ? `?${qs}` : ""}`,
      );
      route.priceHistory = resp.ok ? await resp.json() : null;
    } catch {
      route.priceHistory = null;
    } finally {
      route.priceHistoryLoading = false;
      if (routeIndex === currentActiveRouteIndex) renderActiveRoutePane();
    }
  }

  // Google price graph: fetch on-demand when the user clicks a route tab.
  // For bulk fallback, ensureRouteLoaded populates route.google_price_graph already.
  if (includeGoogleGraph && route.google_price_graph === undefined) {
    route.googlePriceGraphLoading = true;
    if (routeIndex === currentActiveRouteIndex) renderActiveRoutePane();
    try {
      const firstOffer = Array.isArray(route.offers) ? route.offers[0] : null;
      const summary = route.summary || {};
      const departureDate = toYyyyMmDd(
        summary.departure_date ||
          firstOffer?.departure_date ||
          base.departure_date,
      );
      const returnDate = toYyyyMmDd(
        summary.return_date || firstOffer?.return_date || base.return_date,
      );

      const payload = {
        origin: route.origin,
        destination: route.destination,
        departure_date: departureDate,
        return_date: returnDate,
        trip_type: base.trip_type || "round_trip",
        class: base.class || "economy",
        stops: base.stops || "any",
        adults: base.adults || 1,
        children: base.children || 0,
        infants_lap: base.infants_lap || 0,
        infants_seat: base.infants_seat || 0,
        currency: (base.currency || "USD").toUpperCase(),
        include_price_graph: true,
      };
      if (Array.isArray(base.carriers) && base.carriers.length > 0) {
        payload.carriers = base.carriers;
      }
      if (Array.isArray(base.include_airline_groups)) {
        payload.include_airline_groups = base.include_airline_groups;
      }
      if (Array.isArray(base.exclude_airline_groups)) {
        payload.exclude_airline_groups = base.exclude_airline_groups;
      }
      if (base.debug_batches) {
        payload.debug_batches = true;
      }

      if (base.price_graph_window_days) {
        payload.price_graph_window_days = base.price_graph_window_days;
      }
      if (base.price_graph_departure_date_from) {
        payload.price_graph_departure_date_from =
          base.price_graph_departure_date_from;
      }
      if (base.price_graph_departure_date_to) {
        payload.price_graph_departure_date_to = base.price_graph_departure_date_to;
      }
      if (base.price_graph_trip_length_days != null) {
        payload.price_graph_trip_length_days = base.price_graph_trip_length_days;
      }

      const resp = await fetch(ENDPOINTS.SEARCH, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        let msg = `Google price graph unavailable (${resp.status})`;
        try {
          const err = await resp.json();
          msg = err?.error || msg;
        } catch {}
        route.google_price_graph = { error: msg };
      } else {
        const data = await resp.json();
        route.google_price_graph = data?.price_graph || null;
      }
    } finally {
      route.googlePriceGraphLoading = false;
      if (routeIndex === currentActiveRouteIndex) renderActiveRoutePane();
    }
  }
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
      offers.sort((a, b) => {
        const priceA =
          typeof a?.price === "number" &&
          Number.isFinite(a.price) &&
          a.price > 0
            ? a.price
            : Number.POSITIVE_INFINITY;
        const priceB =
          typeof b?.price === "number" &&
          Number.isFinite(b.price) &&
          b.price > 0
            ? b.price
            : Number.POSITIVE_INFINITY;
        return priceA - priceB;
      });
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
function createFlightCard(offer, searchParamsOverride, options) {
  const card = document.createElement("div");
  card.className = "flight-card";

  // Format departure and arrival times
  const segments = Array.isArray(offer.segments) ? offer.segments : [];
  const departureSegment = segments[0];
  const arrivalSegment = segments.length ? segments[segments.length - 1] : null;
  const effectiveSearchParams = searchParamsOverride || currentSearchParams;
  const tripType = effectiveSearchParams?.trip_type || "round_trip";

  const departureTime = departureSegment
    ? new Date(departureSegment.departure_time)
    : new Date(0);
  const arrivalTime = arrivalSegment
    ? new Date(arrivalSegment.arrival_time)
    : new Date(0);

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
    const originAirport = departureSegment?.departure_airport || "";
    const destinationAirport = arrivalSegment?.arrival_airport || "";
    googleFlightsUrl = createGoogleFlightsUrl(
      originAirport,
      destinationAirport,
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

  const flightNumber =
    segments.length > 1
      ? `${departureSegment.flight_number || ""} +${segments.length - 1} more`
      : departureSegment.flight_number || "";

  const routeLabel =
    departureSegment && arrivalSegment
      ? `${departureSegment.departure_airport} - ${arrivalSegment.arrival_airport}`
      : "";

  const airplaneLabel =
    segments.length <= 1
      ? departureSegment.airplane || "Aircraft information unavailable"
      : "Multiple aircraft";

  const stopAirports = getStopAirports(segments);
  const stopDetails =
    stopAirports.length > 0 ? `Stops: ${stopAirports.join(", ")}` : "";

  const airlineGroupBadges = renderAirlineGroupBadges(offer.airline_groups);
  const currency = offer.currency || effectiveSearchParams?.currency || "";
  const hasPrice =
    typeof offer.price === "number" &&
    Number.isFinite(offer.price) &&
    offer.price > 0;
  const fallbackCurrency = options?.routeCurrency || currency;
  const fallbackMinPrice =
    typeof options?.routeMinPrice === "number" &&
    Number.isFinite(options.routeMinPrice) &&
    options.routeMinPrice > 0
      ? options.routeMinPrice
      : null;
  const priceLabel = hasPrice
    ? `${escapeHtml(currency)} ${Number(offer.price).toFixed(2)}`
    : fallbackMinPrice != null
      ? `From ${escapeHtml(fallbackCurrency)} ${Number(
          fallbackMinPrice,
        ).toFixed(2)}`
      : "Price unavailable";

  // Create card content
  card.innerHTML = `
        <div class="row align-items-center">
            <div class="col-md-2">
                <img src="${airlineLogo}"
                     alt="${airlineCode || "Airline"}" class="airline-logo">
                <div>${escapeHtml(airlineCode || "Unknown")} ${escapeHtml(flightNumber)}</div>
                <div class="mt-2">${airlineGroupBadges}</div>
            </div>
            <div class="col-md-3">
                <div class="flight-time">${formattedDepartureTime} - ${formattedArrivalTime}</div>
                <div class="flight-duration">${formattedDuration}</div>
                <div>${escapeHtml(routeLabel)}</div>
            </div>
            <div class="col-md-3">
                <div>${escapeHtml(getStopsLabel(segments))}</div>
                ${
                  stopDetails
                    ? `<div class="text-muted small">${escapeHtml(stopDetails)}</div>`
                    : ""
                }
                <div>${escapeHtml(airplaneLabel)}</div>
            </div>
            <div class="col-md-2 text-end">
                <div class="flight-price">${priceLabel}</div>
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

function getStopAirports(segments) {
  if (!Array.isArray(segments) || segments.length <= 1) return [];
  const out = [];
  const seen = new Set();
  for (let i = 0; i < segments.length - 1; i++) {
    const code = String(segments[i]?.arrival_airport || "")
      .trim()
      .toUpperCase();
    if (!code) continue;
    if (seen.has(code)) continue;
    seen.add(code);
    out.push(code);
  }
  return out;
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
