// App.js
import React, { useState } from 'react';
import SearchForm from './components/SearchForm';
import SearchResults from './components/SearchResults';

function App() {
  const [searchResults, setSearchResults] = useState(null);

  const handleSearch = async (searchParams) => {
    try {
      const response = await fetch('/api/search', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(searchParams),
      });
      const data = await response.json();
      setSearchResults(data);
    } catch (error) {
      console.error('Error fetching search results:', error);
    }
  };

  return (
    <div className="App">
      <h1>Flight Search</h1>
      <SearchForm onSearch={handleSearch} />
      {searchResults && <SearchResults results={searchResults} />}
    </div>
  );
}

export default App;

// components/SearchForm.js
import React, { useState } from 'react';

function SearchForm({ onSearch }) {
  const [formData, setFormData] = useState({
    srcCities: [],
    dstCities: [],
    startDate: '',
    endDate: '',
    tripLength: 7,
    airlines: [],
    travelClass: 'Economy',
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    onSearch(formData);
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Add input fields for all search parameters */}
      <button type="submit">Search Flights</button>
    </form>
  );
}

export default SearchForm;

// components/SearchResults.js
import React from 'react';

function SearchResults({ results }) {
  return (
    <div>
      <h2>Search Results</h2>
      {results.map((offer, index) => (
        <div key={index}>
          <h3>{offer.SrcCity} to {offer.DstCity}</h3>
          <p>Price: ${offer.Price}</p>
          {/* Add more details about the offer */}
        </div>
      ))}
    </div>
  );
}

export default SearchResults;
