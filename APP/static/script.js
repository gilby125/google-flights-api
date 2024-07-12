document.getElementById('searchForm').addEventListener('submit', function(e) {
    e.preventDefault();
    
    const formData = new FormData(this);
    const searchData = {
        srcCities: formData.get('srcCities').split(',').map(city => city.trim()),
        dstCities: formData.get('dstCities').split(',').map(city => city.trim()),
        startDate: formData.get('startDate'),
        endDate: formData.get('endDate'),
        tripLength: parseInt(formData.get('tripLength')),
        airlines: formData.get('airlines').split(',').map(airline => airline.trim()),
        travelClass: formData.get('travelClass')
    };

    fetch('/api/search', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify(searchData),
})
.then(response => response.json())
.then(data => {
    if (data.message) {
        document.getElementById('results').innerHTML = data.message;
    } else {
        document.getElementById('results').innerHTML = JSON.stringify(data, null, 2);
    }
})
.catch((error) => {
    console.error('Error:', error);
    document.getElementById('results').innerHTML = 'An error occurred while processing your request.';
});

