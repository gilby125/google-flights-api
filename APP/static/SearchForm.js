import React, { useState } from 'react';
import { TextField, Select, MenuItem, Slider, Button, InputLabel, FormControl } from '@material-ui/core';
import DateRangePicker from '@material-ui/lab/DateRangePicker';
import AdapterDateFns from '@material-ui/lab/AdapterDateFns';
import LocalizationProvider from '@material-ui/lab/LocalizationProvider';
import './SearchForm.css';

function SearchForm({ onSearch }) {
  const [formData, setFormData] = useState({
    srcCities: [],
    dstCities: [],
    dateRange: [null, null],
    tripLength: 7,
    airlines: [],
    travelClass: 'Economy',
  });

  const handleChange = (name, value) => {
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    onSearch(formData);
  };

  return (
    <form onSubmit={handleSubmit} className="search-form">
      <FormControl fullWidth margin="normal">
        <InputLabel>Source Cities</InputLabel>
        <Select
          multiple
          value={formData.srcCities}
          onChange={(e) => handleChange('srcCities', e.target.value)}
        >
          <MenuItem value="San Francisco">San Francisco</MenuItem>
          <MenuItem value="San Jose">San Jose</MenuItem>
        </Select>
      </FormControl>
      <FormControl fullWidth margin="normal">
        <InputLabel>Destination Cities</InputLabel>
        <Select
          multiple
          value={formData.dstCities}
          onChange={(e) => handleChange('dstCities', e.target.value)}
        >
          <MenuItem value="New York">New York</MenuItem>
          <MenuItem value="Philadelphia">Philadelphia</MenuItem>
          <MenuItem value="Washington">Washington</MenuItem>
        </Select>
      </FormControl>
      <LocalizationProvider dateAdapter={AdapterDateFns}>
        <DateRangePicker
          startText="Start Date"
          endText="End Date"
          value={formData.dateRange}
          onChange={(newValue) => handleChange('dateRange', newValue)}
          renderInput={(startProps, endProps) => (
            <>
              <TextField {...startProps} fullWidth margin="normal" />
              <TextField {...endProps} fullWidth margin="normal" />
            </>
          )}
        />
      </LocalizationProvider>
      <FormControl fullWidth margin="normal">
        <InputLabel>Trip Length</InputLabel>
        <Slider
          value={formData.tripLength}
          onChange={(e, newValue) => handleChange('tripLength', newValue)}
          aria-labelledby="trip-length-slider"
          valueLabelDisplay="auto"
          step={1}
          marks
          min={1}
          max={30}
        />
      </FormControl>
      <FormControl fullWidth margin="normal">
        <InputLabel>Airlines</InputLabel>
        <Select
          multiple
          value={formData.airlines}
          onChange={(e) => handleChange('airlines', e.target.value)}
        >
          <MenuItem value="United">United</MenuItem>
          <MenuItem value="American">American</MenuItem>
        </Select>
      </FormControl>
      <FormControl fullWidth margin="normal">
        <InputLabel>Travel Class</InputLabel>
        <Select
          value={formData.travelClass}
          onChange={(e) => handleChange('travelClass', e.target.value)}
        >
          <MenuItem value="Economy">Economy</MenuItem>
          <MenuItem value="Business">Business</MenuItem>
          <MenuItem value="First">First</MenuItem>
        </Select>
      </FormControl>
      <Button type="submit" variant="contained" color="primary" fullWidth>
        Search Flights
      </Button>
    </form>
  );
}

export default SearchForm;
