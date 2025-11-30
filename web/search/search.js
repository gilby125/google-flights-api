// Flight Search JavaScript

// API endpoints
const API_BASE = "/api";
const ENDPOINTS = {
  SEARCH: `${API_BASE}/search`,
  PRICE_HISTORY: `${API_BASE}/price-history`,
  AIRPORTS: `${API_BASE}/airports`,
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
};

// Chart instance
let priceChart = null;
let currentOffers = [];
let advancedOptionsVisible =
  !!elements.advancedOptions &&
  getComputedStyle(elements.advancedOptions).display !== "none";

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

  if (elements.sortResults) {
    elements.sortResults.addEventListener("change", sortFlightResults);
  }

  loadAirports();
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
  elements.loadingIndicator.setAttribute("aria-busy", isLoading ? "true" : "false");
  elements.loadingIndicator.setAttribute("aria-hidden", isLoading ? "false" : "true");
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
      throw new Error("Please fill in origin, destination, and departure date.");
    }

    const origin = originInput.value.split(" - ")[0].trim().toUpperCase();
    const destination = destinationInput.value
      .split(" - ")[0]
      .trim()
      .toUpperCase();
    const departureDate = departureInput.value;
    const returnDate = returnInput ? returnInput.value : "";
    const tripType = tripTypeInput ? tripTypeInput.value : "round_trip";
    const travelClass = classInput ? classInput.value : "economy";
    const stops = stopsInput ? stopsInput.value : "any";

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

    // Load price history for the route
    await loadPriceHistory(origin, destination);

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

  // Sort and display results
  sortFlightResults();
}

// Sort flight results based on selected criteria
function sortFlightResults() {
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
function createFlightCard(offer) {
  const card = document.createElement("div");
  card.className = "flight-card";

  // Format departure and arrival times
  const departureSegment = offer.segments[0];
  const returnSegment = offer.segments.find((s) => s.is_return);

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
    const tripType = returnSegment ? "round_trip" : "one_way";
    googleFlightsUrl = createGoogleFlightsUrl(
      departureSegment.departure_airport,
      departureSegment.arrival_airport,
      departureTime,
      tripType,
    );
  }

  const airlineCode = departureSegment.airline_code || "";
  const airlineLogo = airlineCode
    ? `https://www.gstatic.com/flights/airline_logos/70px/${airlineCode}.png`
    : "https://via.placeholder.com/70x40?text=Air";

  // Create card content
  card.innerHTML = `
        <div class="row align-items-center">
            <div class="col-md-2">
                <img src="${airlineLogo}" 
                     alt="${airlineCode || "Airline"}" class="airline-logo">
                <div>${airlineCode || "Unknown"} ${departureSegment.flight_number || ""}</div>
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
                <div class="flight-price">${offer.currency} ${offer.price.toFixed(2)}</div>
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
