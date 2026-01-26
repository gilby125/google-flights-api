package flights

import (
	"strings"
	"testing"
	"time"
)

func TestFlightString(t *testing.T) {
	// Use correct field names based on flights/types.go
	f := Flight{
		DepTime:        time.Date(2025, 3, 30, 15, 30, 0, 0, time.UTC),
		ArrTime:        time.Date(2025, 3, 30, 18, 45, 0, 0, time.UTC),
		DepAirportCode: "JFK",
		ArrAirportCode: "LAX",
		FlightNumber:   "AA123",
		AirlineName:    "American Airlines",
		DepCity:        "New York",
		ArrCity:        "Los Angeles",
		DepAirportName: "John F. Kennedy International Airport",
		ArrAirportName: "Los Angeles International Airport",
		Duration:       3*time.Hour + 15*time.Minute,
		Airplane:       "Boeing 737",
		Legroom:        "Average",
	}

	// Adjust expected substrings based on the actual String() method output and correct fields
	expectedSubstrings := []string{
		"DepAirportCode: JFK",
		"DepAirportName: John F. Kennedy International Airport",
		"DepCity: New York",
		"ArrAirportName: Los Angeles International Airport",
		"ArrAirportCode: LAX",
		"ArrCity: Los Angeles",
		"DepTime: 2025-03-30 15:30:00", // Check actual format
		"ArrTime: 2025-03-30 18:45:00", // Check actual format
		"Duration: 3h15m0s",            // Check actual format
		"Airplane: Boeing 737",
		"FlightNumber: AA123",
		"AirlineName: American Airlines",
		"Legroom: Average",
	}
	result := f.String()

	for _, substr := range expectedSubstrings {
		if !strings.Contains(result, substr) {
			t.Errorf("String() output missing expected substring %q", substr)
		}
	}
}

func TestOfferString(t *testing.T) {
	// Use correct field names based on flights/types.go
	o := Offer{
		StartDate:  time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
		ReturnDate: time.Date(2025, 4, 17, 0, 0, 0, 0, time.UTC),
		Price:      499.99,
	}

	output := o.String()
	// Check against the actual String() format: "{YYYY-MM-DD YYYY-MM-DD Price}"
	expected := "{2025-04-10 2025-04-17 499}" // Price is formatted as int in String()
	if output != expected {
		t.Errorf("Offer String() incorrect. Got: %q, Want: %q", output, expected)
	}
}

func TestFullOfferString(t *testing.T) {
	// Use correct field names based on flights/types.go
	// FullOffer embeds Offer, so it has StartDate, ReturnDate, Price
	fo := FullOffer{
		Offer: Offer{
			StartDate:  time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			ReturnDate: time.Date(2025, 5, 8, 0, 0, 0, 0, time.UTC),
			Price:      799.99,
		},
		Flight: []Flight{
			{DepAirportCode: "LHR", ArrAirportCode: "CDG", FlightNumber: "BA308"},
		},
		SrcAirportCode: "LHR",
		DstAirportCode: "CDG",
		SrcCity:        "London",
		DstCity:        "Paris",
		FlightDuration: 1*time.Hour + 15*time.Minute,
	}

	output := fo.String()
	// Check for substrings based on the actual String() method
	expectedSubstrings := []string{
		"StartDate: 2025-05-01",  // Check format
		"ReturnDate: 2025-05-08", // Check format
		"Price: 799",             // Price formatted as int
		"Flight: [",              // Start of flight details
		"DepAirportCode: LHR",
		"ArrAirportCode: CDG",
		"FlightNumber: BA308",
		"SrcAirportCode: LHR",
		"DstAirportCode: CDG",
		"SrcCity: London",
		"DstCity: Paris",
		"FlightDuration: 1h15m0s", // Check format
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(output, substr) {
			t.Errorf("FullOffer String() output missing expected substring %q in %q", substr, output)
		}
	}
}

const wrongAirportCode = "wrong"

func testValidateOffersArgs(t *testing.T, args Args, wantErr string) {
	gotErr := args.ValidateOffersArgs()

	if gotErr == nil {
		t.Fatalf("Validate call, should result in error args: %v", args)
	} else if gotErr.Error() != wantErr {
		t.Fatalf(`Wrong error want: "%s", got: "%s"`, wantErr, gotErr.Error())
	}
}

func testValidatePriceGraphArgs(t *testing.T, args PriceGraphArgs, wantErr string) {
	gotErr := args.Validate()

	if gotErr == nil {
		t.Fatalf("Validate call should result in error args: %v", args)
	} else if gotErr.Error() != wantErr {
		t.Fatalf(`Wrong error want: "%s", got: "%s"`, wantErr, gotErr.Error())
	}
}

func testValidatePriceGraphArgsOK(t *testing.T, args PriceGraphArgs) PriceGraphArgs {
	got := args
	if err := got.Validate(); err != nil {
		t.Fatalf("Validate call should not fail, args: %v, err: %v", args, err)
	}
	return got
}

func testValidateURLArg(t *testing.T, args Args, wantErr string) {
	gotErr := args.ValidateURLArgs()

	if gotErr == nil {
		t.Fatalf("Validate call should result in error args: %v", args)
	} else if gotErr.Error() != wantErr {
		t.Fatalf(`Wrong error want: "%s", got: "%s"`, wantErr, gotErr.Error())
	}
}

func TestValidateOffersArgs(t *testing.T) {
	args := Args{SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{}, DstAirports: []string{}}
	testValidateOffersArgs(t, args, "dst locations: number of locations should be at least 1, specified: 0")

	args = Args{SrcCities: []string{}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{}}
	testValidateOffersArgs(t, args, "src locations: number of locations should be at least 1, specified: 0")

	args = Args{SrcCities: []string{"abc"}, SrcAirports: []string{wrongAirportCode}, DstCities: []string{"abc"}, DstAirports: []string{}}
	testValidateOffersArgs(t, args, "src airport 'wrong' is not an airport code")

	args = Args{SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{wrongAirportCode}}
	testValidateOffersArgs(t, args, "dst airport 'wrong' is not an airport code")

	args = Args{
		SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{},
		Date:       time.Now().AddDate(0, 0, 3),
		ReturnDate: time.Now().AddDate(0, 0, 1),
	}
	testValidateOffersArgs(t, args, "returnDate is before date")

	args = Args{
		SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{},
		Date:       time.Now().AddDate(0, 0, -1),
		ReturnDate: time.Now().AddDate(0, 0, 1),
	}
	testValidateOffersArgs(t, args, "date is before today's date")
}

func TestValidatePriceGraphArgs(t *testing.T) {
	args := PriceGraphArgs{SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{}, DstAirports: []string{}}
	testValidatePriceGraphArgs(t, args, "dst locations: number of locations should be at least 1, specified: 0")

	args = PriceGraphArgs{SrcCities: []string{}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{}}
	testValidatePriceGraphArgs(t, args, "src locations: number of locations should be at least 1, specified: 0")

	args = PriceGraphArgs{SrcCities: []string{"abc"}, SrcAirports: []string{wrongAirportCode}, DstCities: []string{"abc"}, DstAirports: []string{}}
	testValidatePriceGraphArgs(t, args, "src airport 'wrong' is not an airport code")

	args = PriceGraphArgs{SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{wrongAirportCode}}
	testValidatePriceGraphArgs(t, args, "dst airport 'wrong' is not an airport code")

	args = PriceGraphArgs{
		SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{},
		RangeStartDate: time.Now().AddDate(0, 0, 5),
		RangeEndDate:   time.Now().AddDate(0, 0, 170),
	}
	testValidatePriceGraphArgs(t, args, "number of days between dates is larger than 161, 165")

	args = PriceGraphArgs{
		SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{},
		RangeStartDate: time.Now().AddDate(0, 0, 2),
		RangeEndDate:   time.Now().AddDate(0, 0, 2),
	}
	validated := testValidatePriceGraphArgsOK(t, args)
	if !validated.RangeEndDate.After(validated.RangeStartDate) {
		t.Fatalf("expected RangeEndDate to be auto-adjusted, got start=%v end=%v", validated.RangeStartDate, validated.RangeEndDate)
	}

	args = PriceGraphArgs{
		SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{},
		RangeStartDate: time.Now().AddDate(0, 0, 5),
		RangeEndDate:   time.Now().AddDate(0, 0, 2),
	}
	testValidatePriceGraphArgs(t, args, "rangeEndDate is before rangeStartDate")

	args = PriceGraphArgs{
		SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{},
		RangeStartDate: time.Now().AddDate(0, 0, -1),
		RangeEndDate:   time.Now().AddDate(0, 0, 2),
	}
	testValidatePriceGraphArgs(t, args, "rangeStartDate is before today's date")
}

func TestValidateRangeDate_AllowsToday(t *testing.T) {
	prev := timeNow
	t.Cleanup(func() { timeNow = prev })

	loc := time.FixedZone("test", 2*60*60)
	timeNow = func() time.Time {
		return time.Date(2026, 1, 26, 19, 37, 0, 0, loc)
	}

	rangeStartDate := time.Date(2026, 1, 26, 0, 0, 0, 0, loc)
	rangeEndDate := time.Date(2026, 1, 27, 0, 0, 0, 0, loc)
	if err := validateRangeDate(rangeStartDate, rangeEndDate); err != nil {
		t.Fatalf("expected start date on today's date to be allowed, got error: %v", err)
	}
}

func TestValidateURLArgs(t *testing.T) {
	args := Args{SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{}, DstAirports: []string{}}
	testValidateURLArg(t, args, "dst locations: number of locations should be at least 1, specified: 0")

	args = Args{SrcCities: []string{}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{}}
	testValidateURLArg(t, args, "src locations: number of locations should be at least 1, specified: 0")

	args = Args{SrcCities: []string{"abc"}, SrcAirports: []string{wrongAirportCode}, DstCities: []string{"abc"}, DstAirports: []string{}}
	testValidateURLArg(t, args, "src airport 'wrong' is not an airport code")

	args = Args{SrcCities: []string{"abc"}, SrcAirports: []string{}, DstCities: []string{"abc"}, DstAirports: []string{wrongAirportCode}}
	testValidateURLArg(t, args, "dst airport 'wrong' is not an airport code")
}
