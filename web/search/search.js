// Flight Search JavaScript

// API endpoints
const API_BASE = '/api/v1';
const ENDPOINTS = {
    SEARCH: `${API_BASE}/search`,
    BULK_SEARCH: `${API_BASE}/bulk-search`,
    PRICE_HISTORY: `${API_BASE}/price-history`,
    AIRPORTS: `${API_BASE}/airports`,
    AIRLINES: `${API_BASE}/airlines`
};

// DOM elements
const elements = {
    searchForm: document.getElementById('searchForm'),
    advancedBtn: document.getElementById('advancedBtn'),
    advancedOptions: document.getElementById('advancedOptions'),
    loadingIndicator: document.getElementById('loadingIndicator'),
    priceGraphCard: document.getElementById('priceGraphCard'),
    priceGraph: document.getElementById('priceGraph'),
    resultsCard: document.getElementById('resultsCard'),
    flightResults: document.getElementById('flightResults'),
    sortResults: document.getElementById('sortResults')
};

// Chart instance
let priceChart = null;

// Initialize the search page
function initSearchPage() {
    // Initialize date pickers
    flatpickr('#departureDate', {
        minDate: 'today',
        dateFormat: 'Y-m-d'
    });
    
    flatpickr('#returnDate', {
        minDate: 'today',
        dateFormat: 'Y-m-d'
    });
    
    // Toggle advanced options
    elements.advancedBtn.addEventListener('click', () => {
        elements.advancedOptions.style.display = 
            elements.advancedOptions.style.display === 'none' ? 'block' : 'none';
    });
    
    // Handle form submission
    elements.searchForm.addEventListener('submit', handleSearch);
    
    // Handle sort change
    elements.sortResults.addEventListener('change', sortFlightResults);
    
    // Load airports for autocomplete
    loadAirports();
}

// Load airports for autocomplete
async function loadAirports() {
    try {
        const response = await fetch(ENDPOINTS.AIRPORTS);
        if (!response.ok) throw new Error('Failed to load airports');
        
        const airports = await response.json();
        
        // Implement autocomplete for origin and destination fields
        // This is a simplified implementation - a real one would use a proper autocomplete library
        const airportOptions = airports.map(airport => `${airport.code} - ${airport.name}, ${airport.city}`);
        
        // Add datalist for autocomplete
        const originDatalist = document.createElement('datalist');
        originDatalist.id = 'airportList';
        
        airportOptions.forEach(option => {
            const optionEl = document.createElement('option');
            optionEl.value = option;
            originDatalist.appendChild(optionEl);
        });
        
        document.body.appendChild(originDatalist);
        
        // Connect datalist to inputs
        document.getElementById('origin').setAttribute('list', 'airportList');
        document.getElementById('destination').setAttribute('list', 'airportList');
        
    } catch (error) {
        console.error('Error loading airports:', error);
    }
}

// Handle search form submission
async function handleSearch(event) {
    event.preventDefault();
    
    // Show loading indicator
    elements.loadingIndicator.style.display = 'block';
    elements.priceGraphCard.style.display = 'none';
    elements.resultsCard.style.display = 'none';
    
    // Get form values
    const origin = document.getElementById('origin').value.split(' - ')[0].trim();
    const destination = document.getElementById('destination').value.split(' - ')[0].trim();
    const departureDate = document.getElementById('departureDate').value;
    const returnDate = document.getElementById('returnDate').value;
    const tripType = document.getElementById('tripType').value;
    const travelClass = document.getElementById('class').value;
    const stops = document.getElementById('stops').value;
    
    // Get advanced options if visible
    let adults = 1;
    let children = 0;
    let infantsLap = 0;
    let infantsSeat = 0;
    let currency = 'USD';
    
    if (elements.advancedOptions.style.display === 'block') {
        adults = parseInt(document.getElementById('adults').value);
        children = parseInt(document.getElementById('children').value);
        infantsLap = parseInt(document.getElementById('infantsLap').value);
        infantsSeat = parseInt(document.getElementById('infantsSeat').value);
        currency = document.getElementById('currency').value;
    } else {
        // Use the simple passengers field
        adults = parseInt(document.getElementById('passengers').value);
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
        currency: currency
    };
    
    // Add return date if round trip
    if (tripType === 'round_trip' && returnDate) {
        searchData.return_date = returnDate;
    }
    
    try {
        // Send search request
        const response = await fetch(ENDPOINTS.SEARCH, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(searchData)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Search failed');
        }
        
        const searchResult = await response.json();
        
        // Load price history for the route
        await loadPriceHistory(origin, destination);
        
        // Display search results
        displaySearchResults(searchResult);
        
    } catch (error) {
        console.error('Error searching flights:', error);
        showAlert(`Error searching flights: ${error.message}`, 'danger');
    } finally {
        // Hide loading indicator
        elements.loadingIndicator.style.display = 'none';
    }
}

// Load price history for a route
async function loadPriceHistory(origin, destination) {
    try {
        const response = await fetch(`${ENDPOINTS.PRICE_HISTORY}/${origin}/${destination}`);
        if (!response.ok) throw new Error('Failed to load price history');
        
        const priceHistory = await response.json();
        
        // Display price graph
        displayPriceGraph(priceHistory);
        
    } catch (error) {
        console.error('Error loading price history:', error);
        // Don't show an alert for this non-critical error
    }
}

// Display price graph
function displayPriceGraph(priceHistory) {
    // Show price graph card
    elements.priceGraphCard.style.display = 'block';
    
    // Prepare data for chart
    const dates = priceHistory.map(item => item.date);
    const prices = priceHistory.map(item => item.price);
    
    // Destroy existing chart if it exists
    if (priceChart) {
        priceChart.destroy();
    }
    
    // Create new chart
    const ctx = elements.priceGraph.getContext('2d');
    priceChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: dates,
            datasets: [{
                label: 'Price',
                data: prices,
                borderColor: 'rgba(75, 192, 192, 1)',
                backgroundColor: 'rgba(75, 192, 192, 0.2)',
                tension: 0.1,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: false,
                    title: {
                        display: true,
                        text: 'Price'
                    }
                },
                x: {
                    title: {
                        display: true,
                        text: 'Date'
                    }
                }
            }
        }
    });
}

// Display search results
function displaySearchResults(searchResult) {
    // Show results card
    elements.resultsCard.style.display = 'block';
    
    // Clear previous results
    elements.flightResults.innerHTML = '';
    
    // Check if we have results
    if (!searchResult.offers || searchResult.offers.length === 0) {
        elements.flightResults.innerHTML = '<div class="alert alert-info">No flights found matching your criteria.</div>';
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
        case 'price':
            offers.sort((a, b) => a.price - b.price);
            break;
        case 'duration':
            offers.sort((a, b) => a.total_duration - b.total_duration);
            break;
        case 'departure':
            offers.sort((a, b) => {
                const dateA = new Date(a.segments[0].departure_time);
                const dateB = new Date(b.segments[0].departure_time);
                return dateA - dateB;
            });
            break;
    }
    
    // Clear previous results
    elements.flightResults.innerHTML = '';
    
    // Display sorted offers
    offers.forEach(offer => {
        const card = createFlightCard(offer);
        elements.flightResults.appendChild(card);
    });
}

// Create a flight card element
function createFlightCard(offer) {
    const card = document.createElement('div');
    card.className = 'flight-card';
    
    // Format departure and arrival times
    const departureSegment = offer.segments[0];
    const returnSegment = offer.segments.find(s => s.is_return);
    
    const departureTime = new Date(departureSegment.departure_time);
    const arrivalTime = new Date(departureSegment.arrival_time);
    
    const formattedDepartureTime = departureTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    const formattedArrivalTime = arrivalTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    
    // Format duration
    const hours = Math.floor(offer.total_duration / 60);
    const minutes = offer.total_duration % 60;
    const formattedDuration = `${hours}h ${minutes}m`;
    
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
                <div>${departureSegment.airplane || 'Aircraft information not available'}</div>
            </div>
            <div class="col-md-2">
                <div class="text-uppercase">${offer.class}</div>
                <div>${offer.legroom || 'Standard legroom'}</div>
            </div>
            <div class="col-md-2 text-end">
                <div class="flight-price">${offer.currency} ${offer.price.toFixed(2)}</div>
                <button class="btn btn-sm btn-primary mt-2">Select</button>
            </div>
        </div>
    `;
    
    // Add return flight info if available
    if (returnSegment) {
        const returnDepartureTime = new Date(returnSegment.departure_time);
        const returnArrivalTime = new Date(returnSegment.arrival_time);
        
        const formattedReturnDepartureTime = returnDepartureTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        const formattedReturnArrivalTime = returnArrivalTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        
        const returnHours = Math.floor(returnSegment.duration / 60);
        const returnMinutes = returnSegment.duration % 60;
        const formattedReturnDuration = `${returnHours}h ${returnMinutes}m`;
        
        const returnInfo = document.createElement('div');
        returnInfo.className = 'row align-items-center mt-3 pt-3 border-top';
        returnInfo.innerHTML = `
            <div class="col-md-2">
                <img src="https://www.gstatic.com/flights/airline_logos/70px/${returnSegment.airline_code}.png" 
                     alt="${returnSegment.airline_code}" class="airline-logo">
                <div>${returnSegment.airline_code} ${returnSegment.flight_number}</div>
            </div>
            <div class="col-md-3">
                <div class="flight-time">${formattedReturnDepartureTime} - ${formattedReturnArrivalTime}</div>
                <div class="flight-duration">${formattedReturnDuration}</div>
                <div>${returnSegment.departure_airport} - ${returnSegment.arrival_airport}</div>
            </div>
            <div class="col-md-3">
                <div>${getStopsLabel([returnSegment])}</div>
                <div>${returnSegment.airplane || 'Aircraft information