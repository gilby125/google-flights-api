import React from 'react';
import { motion } from 'framer-motion';
import { FaPlane, FaCalendar, FaDollarSign } from 'react-icons/fa';
import './SearchResults.css';

function SearchResults({ results, isLoading, progress }) {
  if (isLoading) {
    return (
      <div className="loading-indicator">
        <h3>Searching for the best flight deals...</h3>
        <p>{progress}</p>
      </div>
    );
  }

  return (
    <div className="search-results">
      {results.map((offer, index) => (
        <motion.div
          key={index}
          className="offer-card"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: index * 0.1 }}
        >
          <div className="card-header">
            <h3><FaPlane /> {offer.srcCity} to {offer.dstCity}</h3>
            <span className="price"><FaDollarSign />{offer.price}</span>
          </div>
          <div className="card-body">
            <div className="flight-details">
              <div className="outbound">
                <h4>Outbound</h4>
                <p><FaCalendar /> {offer.startDate}</p>
                <p>{offer.srcCity} → {offer.dstCity}</p>
              </div>
              <div className="return">
                <h4>Return</h4>
                <p><FaCalendar /> {offer.returnDate}</p>
                <p>{offer.dstCity} → {offer.srcCity}</p>
              </div>
            </div>
            <p className="airline">Airline: {offer.airline}</p>
            {offer.url && (
              <a href={offer.url} target="_blank" rel="noopener noreferrer" className="btn btn-primary">
                Book on Google Flights
              </a>
            )}
          </div>
        </motion.div>
      ))}
    </div>
  );
}

export default SearchResults;
