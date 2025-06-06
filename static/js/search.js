// ... existing code ...

// Make sure the event listener is properly attached to the search button
document.addEventListener('DOMContentLoaded', function() {
    const searchButton = document.getElementById('search-button');
    if (searchButton) {
        searchButton.addEventListener('click', searchFlights);
    } else {
        console.error("Search button not found in the DOM");
    }
});

// Ensure the searchFlights function is properly implemented
function searchFlights() {
    console.log("Search flights function called");
    
    // Get form values
    const origin = document.getElementById('origin').value;
    const destination = document.getElementById('destination').value;
    const departDate = document.getElementById('depart-date').value;
    const returnDate = document.getElementById('return-date').value;
    
    // Validate inputs
    if (!origin || !destination || !departDate) {
        alert("Please fill in all required fields (origin, destination, and departure date)");
        return;
    }
    
    // Show loading indicator
    const resultsContainer = document.getElementById('search-results');
    if (resultsContainer) {
        resultsContainer.innerHTML = '<div class="loading">Searching for flights...</div>';
    }
    
    // Prepare search parameters
    const searchParams = {
        origin: origin,
        destination: destination,
        departure_date: departDate,
        return_date: returnDate,
        adults: 1,
        trip_type: returnDate ? 'round_trip' : 'one_way',
        class: 'economy',
        stops: 'any',
        currency: 'USD'
    };
    
    console.log('Sending search request:', searchParams);
    
    // Make API request
    fetch('/api/search', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(searchParams)
    })
    .then(response => {
        console.log('Response status:', response.status);
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    })
    .then(data => {
        console.log('Search results:', data);
        displaySearchResults(data);
    })
    .catch(error => {
        console.error('Error during search:', error);
        if (resultsContainer) {
            resultsContainer.innerHTML = '<div class="error">An error occurred while searching for flights. Please try again.</div>';
        }
    });
}

// Function to display search results
function displaySearchResults(data) {
    const resultsContainer = document.getElementById('search-results');
    if (!resultsContainer) return;
    
    if (!data || !data.offers || data.offers.length === 0) {
        resultsContainer.innerHTML = '<div class="no-results">No flights found. Please try different search criteria.</div>';
        return;
    }
    
    let resultsHTML = '<div class="results-container">';
    
    data.offers.forEach(offer => {
        // Get the first segment for basic info
        const firstSegment = offer.segments && offer.segments.length > 0 ? offer.segments[0] : {};
        
        // Create Google Flights URL if we have departure information
        let googleFlightsUrl = '#';
        if (firstSegment.departure_airport && firstSegment.arrival_airport && firstSegment.departure_time) {
            const departureDate = new Date(firstSegment.departure_time);
            const formattedDate = departureDate.toISOString().split('T')[0];
            const origin = encodeURIComponent(firstSegment.departure_airport);
            const destination = encodeURIComponent(firstSegment.arrival_airport);
            googleFlightsUrl = `https://www.google.com/travel/flights?q=Flights%20from%20${origin}%20to%20${destination}%20on%20${formattedDate}`;
        }
        
        resultsHTML += `
            <div class="flight-card">
                <div class="flight-header">
                    <span class="airline">${firstSegment.airline || 'Unknown Airline'}</span>
                    <span class="price">$${offer.price} ${offer.currency}</span>
                </div>
                <div class="flight-details">
                    <div class="departure">
                        <div class="time">${firstSegment.departure_time ? new Date(firstSegment.departure_time).toLocaleTimeString() : 'N/A'}</div>
                        <div class="airport">${firstSegment.departure_airport || 'N/A'}</div>
                    </div>
                    <div class="flight-duration">
                        <div class="duration">${Math.floor(offer.total_duration / 60)}h ${offer.total_duration % 60}m</div>
                        <div class="flight-line">——————→</div>
                    </div>
                    <div class="arrival">
                        <div class="time">${firstSegment.arrival_time ? new Date(firstSegment.arrival_time).toLocaleTimeString() : 'N/A'}</div>
                        <div class="airport">${firstSegment.arrival_airport || 'N/A'}</div>
                    </div>
                </div>
                <div class="flight-actions">
                    <a href="${googleFlightsUrl}" target="_blank" class="google-flights-link">View on Google Flights</a>
                </div>
            </div>
        `;
    });
    
    resultsHTML += '</div>';
    resultsContainer.innerHTML = resultsHTML;
}
// ... existing code ...