// Flight Search JavaScript

// API endpoints
const API_BASE = "/api";
const ENDPOINTS = {
  SEARCH: `${API_BASE}/search`,
  BULK_SEARCH: `${API_BASE}/bulk-search`,
  PRICE_HISTORY: `${API_BASE}/price-history`,
  AIRPORTS: `${API_BASE}/airports`,
  AIRLINES: `${API_BASE}/airlines`,
};

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
};

// Chart instance
let priceChart = null;

// Initialize the search page
function initSearchPage() {
  // Initialize date pickers
  flatpickr("#departure_date", {
    minDate: "today",
    dateFormat: "Y-m-d",
  });

  flatpickr("#return_date", {
    minDate: "today",
    dateFormat: "Y-m-d",
  });

  // Toggle advanced options
  elements.advancedBtn.addEventListener("click", () => {
    elements.advancedOptions.style.display =
      elements.advancedOptions.style.display === "none" ? "block" : "none";
  });

  // Handle form submission
  elements.searchForm.addEventListener("submit", handleSearch);

  // Handle sort change
  elements.sortResults.addEventListener("change", sortFlightResults);

  // Load airports for autocomplete
  loadAirports();
}

// Load airports for autocomplete
// Add this function after initSearchPage
async function loadAirports() {
  try {
    const response = await fetch(ENDPOINTS.AIRPORTS);
    if (!response.ok) {
      throw new Error("Failed to load airports");
    }

    const airports = await response.json();

    // Create datalist for airports
    const originDatalist = document.createElement("datalist");
    originDatalist.id = "originList";

    const destDatalist = document.createElement("datalist");
    destDatalist.id = "destinationList";

    // Add options to datalists
    airports.forEach((airport) => {
      const option = document.createElement("option");
      option.value = `${airport.code} - ${airport.name}, ${airport.city}`;

      originDatalist.appendChild(option.cloneNode(true));
      destDatalist.appendChild(option);
    });

    // Add datalists to the document
    document.body.appendChild(originDatalist);
    document.body.appendChild(destDatalist);

    // Connect inputs to datalists
    document.getElementById("origin").setAttribute("list", "originList");
    document
      .getElementById("destination")
      .setAttribute("list", "destinationList");
  } catch (error) {
    console.error("Error loading airports:", error);
    showAlert("Failed to load airport data", "warning");
  }
}

// Add this function for price history
async function loadPriceHistory(origin, destination) {
  try {
    const response = await fetch(
      `${ENDPOINTS.PRICE_HISTORY}?origin=${origin}&destination=${destination}`,
    );
    if (!response.ok) {
      throw new Error("Failed to load price history");
    }

    const priceHistory = await response.json();

    // Show price graph card
    elements.priceGraphCard.style.display = "block";

    // Create price graph
    const ctx = document.getElementById("priceGraph").getContext("2d");
    new Chart(ctx, {
      type: "line",
      data: {
        labels: priceHistory.map((item) => item.date),
        datasets: [
          {
            label: "Price History",
            data: priceHistory.map((item) => item.price),
            borderColor: "rgb(75, 192, 192)",
            tension: 0.1,
          },
        ],
      },
      options: {
        responsive: true,
        scales: {
          y: {
            beginAtZero: false,
          },
        },
      },
    });
  } catch (error) {
    console.error("Error loading price history:", error);
    // Don't show the price graph card if there's an error
    elements.priceGraphCard.style.display = "none";
  }
}

// Add this function for creating flight cards
function createFlightCard(offer) {
  const card = document.createElement("div");
  card.className = "flight-card";

  // Get outbound segment (first segment)
  const outbound = offer.segments[0];

  // Format times
  const departureTime = new Date(outbound.departure_time);
  const arrivalTime = new Date(outbound.arrival_time);

  // Format duration
  const hours = Math.floor(offer.total_duration / 60);
  const minutes = offer.total_duration % 60;
  const durationText = `${hours}h ${minutes}m`;

  // Create card content
  card.innerHTML = `
        <div class="row align-items-center">
            <div class="col-md-2">
                <div class="d-flex align-items-center">
                    <img src="https://via.placeholder.com/32" class="airline-logo me-2" alt="${outbound.airline}">
                    <div>
                        <div>${outbound.airline}</div>
                        <small>${outbound.flight_number}</small>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="flight-time">${departureTime.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</div>
                <div>${outbound.departure_airport}</div>
            </div>
            <div class="col-md-2 text-center">
                <div class="flight-duration">${durationText}</div>
                <div class="flight-path">
                    <i class="bi bi-arrow-right"></i>
                </div>
                <div>${offer.segments.length > 1 ? "Round Trip" : "One Way"}</div>
            </div>
            <div class="col-md-3">
                <div class="flight-time">${arrivalTime.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</div>
                <div>${outbound.arrival_airport}</div>
            </div>
            <div class="col-md-2 text-end">
                <div class="flight-price">${offer.currency} ${offer.price.toFixed(2)}</div>
                <button class="btn btn-sm btn-outline-primary mt-2">Select</button>
            </div>
        </div>
    `;

  return card;
}

// Add this function for showing alerts
function showAlert(message, type = "info") {
  const alertDiv = document.createElement("div");
  alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
  alertDiv.role = "alert";
  alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    `;

  // Insert at the top of the container
  const container = document.querySelector(".container");
  container.insertBefore(alertDiv, container.firstChild);

  // Auto dismiss after 5 seconds
  setTimeout(() => {
    alertDiv.classList.remove("show");
    setTimeout(() => alertDiv.remove(), 150);
  }, 5000);
}

// Handle search form submission
async function handleSearch(event) {
  event.preventDefault();

  // Show loading indicator
  elements.loadingIndicator.style.display = "block";
  elements.priceGraphCard.style.display = "none";
  elements.resultsCard.style.display = "none";

  // Get form values
  const origin = document.getElementById("origin").value.split(" - ")[0].trim();
  const destination = document
    .getElementById("destination")
    .value.split(" - ")[0]
    .trim();
  const departureDate = document.getElementById("departure_date").value;
  const returnDate = document.getElementById("return_date").value;
  const tripType = document.getElementById("trip_type").value;
  const travelClass = document.getElementById("class")
    ? document.getElementById("class").value
    : "economy";
  const stops = document.getElementById("stops")
    ? document.getElementById("stops").value
    : "any";

  // Get advanced options if visible
  let adults = 1;
  let children = 0;
  let infantsLap = 0;
  let infantsSeat = 0;
  let currency = "USD";

  if (
    elements.advancedOptions &&
    elements.advancedOptions.style.display === "block"
  ) {
    adults = parseInt(document.getElementById("adults").value);
    children = parseInt(document.getElementById("children").value || "0");
    infantsLap = parseInt(document.getElementById("infantsLap").value || "0");
    infantsSeat = parseInt(document.getElementById("infantsSeat").value || "0");
    currency = document.getElementById("currency")
      ? document.getElementById("currency").value
      : "USD";
  } else {
    // Use the simple passengers field
    adults = parseInt(document.getElementById("adults").value);
  }

  // Create search request
  const searchData = {
    origin: origin,
    destination: destination,
    departure_date: departureDate,
    trip_type: tripType,
    class: travelClass,
    stops: stops,
    adults: adults,
    children: children,
    infants_lap: infantsLap,
    infants_seat: infantsSeat,
    currency: currency,
  };

  // Add return date if round trip
  if (tripType === "round_trip" && returnDate) {
    searchData.return_date = returnDate;
  }

  console.log("Sending search request:", searchData);

  try {
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
    // Hide loading indicator
    elements.loadingIndicator.style.display = "none";
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
  // Show price graph card
  elements.priceGraphCard.style.display = "block";

  // Prepare data for chart
  const dates = priceHistory.history.map((item) => item.date);
  const prices = priceHistory.history.map((item) => item.price);

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
  elements.resultsCard.style.display = "block";

  // Clear previous results
  elements.flightResults.innerHTML = "";

  // Check if we have results
  if (!searchResult.offers || searchResult.offers.length === 0) {
    elements.flightResults.innerHTML =
      '<div class="alert alert-info">No flights found matching your criteria.</div>';
    return;
  }

  // Store the offers globally for sorting
  window.flightOffers = searchResult.offers;

  // Sort and display results
  sortFlightResults();
}

// Sort flight results based on selected criteria
function sortFlightResults() {
  const sortBy = elements.sortResults.value;
  const offers = window.flightOffers;

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

  // Create card content
  card.innerHTML = `
        <div class="row align-items-center">
            <div class="col-md-2">
                <img src="https://www.gstatic.com/flights/airline_logos/70px/${departureSegment.airline_code}.png" 
                     alt="${departureSegment.airline_code}" class="airline-logo">
                <div>${departureSegment.airline_code} ${departureSegment.flight_number}</div>
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
