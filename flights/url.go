package flights

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gilby125/google-flights-api/flights/internal/urlpb"
	"google.golang.org/protobuf/proto"
)

func serializeLocations(locations []string, locationType urlpb.Url_LocationType) []*urlpb.Url_Location {
	locationsRet := make([]*urlpb.Url_Location, 0, len(locations))
	for _, l := range locations {
		// Memory allocation for each element :/
		locationsRet = append(locationsRet, &urlpb.Url_Location{
			Type: locationType,
			Name: l,
		})
	}
	return locationsRet
}

func serializeFlight(
	date time.Time,
	srcCities, srcAirports, dstCities, dstAirports []string,
	stops Stops,
	carriers []string,
) *urlpb.Url_Flight {
	flight := &urlpb.Url_Flight{
		Date:         date.Format(time.DateOnly),
		SrcLocations: append(serializeLocations(srcCities, urlpb.Url_CITY), serializeLocations(srcAirports, urlpb.Url_AIRPORT)...),
		DstLocations: append(serializeLocations(dstCities, urlpb.Url_CITY), serializeLocations(dstAirports, urlpb.Url_AIRPORT)...),
		Stops:        urlpb.Url_Stops(stops).Enum(),
	}
	if len(carriers) > 0 {
		flight.Carriers = append([]string{}, carriers...)
	}
	return flight
}

func serializeFlights(args Args) []*urlpb.Url_Flight {
	if args.TripType == MultiCity {
		flights := make([]*urlpb.Url_Flight, len(args.Segments))
		for i, seg := range args.Segments {
			flights[i] = serializeFlight(seg.Date, seg.SrcCities, seg.SrcAirports, seg.DstCities, seg.DstAirports, args.Stops, args.Carriers)
		}
		return flights
	}
	if args.TripType == OneWay {
		return []*urlpb.Url_Flight{
			serializeFlight(args.Date, args.SrcCities, args.SrcAirports, args.DstCities, args.DstAirports, args.Stops, args.Carriers),
		}
	}
	return []*urlpb.Url_Flight{
		serializeFlight(args.Date, args.SrcCities, args.SrcAirports, args.DstCities, args.DstAirports, args.Stops, args.Carriers),
		serializeFlight(args.ReturnDate, args.DstCities, args.DstAirports, args.SrcCities, args.SrcAirports, args.Stops, args.Carriers),
	}
}

func serializeTravelers(travelers Travelers) []urlpb.Url_Traveler {
	travelersRet := make([]urlpb.Url_Traveler, 0,
		travelers.Adults+travelers.Children+
			travelers.InfantInSeat+travelers.InfantOnLap)

	for i := 0; i < travelers.Adults; i++ {
		travelersRet = append(travelersRet, urlpb.Url_ADULT)
	}
	for i := 0; i < travelers.Children; i++ {
		travelersRet = append(travelersRet, urlpb.Url_CHILD)
	}
	for i := 0; i < travelers.InfantInSeat; i++ {
		travelersRet = append(travelersRet, urlpb.Url_INFANT_IN_SEAT)
	}
	for i := 0; i < travelers.InfantOnLap; i++ {
		travelersRet = append(travelersRet, urlpb.Url_INFANT_ON_LAP)
	}
	return travelersRet
}

// The function serializes arguments to the Google Flight URL. The city names should be provided in the
// language described by args.Lang. The args.Lang language is also used to set the language of the
// website to which the serialized URL leads.
//
// GetPriceGraph returns an error if any of the requests fail or if any of the city names are misspelled.
//
// Requirements are described by the [Args.ValidateURLArgs] function.
func (s *Session) SerializeURL(ctx context.Context, args Args) (string, error) {
	var err error

	if err = args.ValidateURLArgs(); err != nil {
		return "", err
	}

	if args.TripType == MultiCity {
		for i := range args.Segments {
			args.Segments[i].SrcCities, err = s.abbrCities(ctx, args.Segments[i].SrcCities, args.Lang)
			if err != nil {
				return "", fmt.Errorf("could not get abbreviated src cities for segment %d: %v", i, err)
			}

			args.Segments[i].DstCities, err = s.abbrCities(ctx, args.Segments[i].DstCities, args.Lang)
			if err != nil {
				return "", fmt.Errorf("could not get abbreviated dst cities for segment %d: %v", i, err)
			}
		}
	} else {
		args.SrcCities, err = s.abbrCities(ctx, args.SrcCities, args.Lang)
		if err != nil {
			return "", fmt.Errorf("could not get abbreviated src cities: %v", err)
		}

		args.DstCities, err = s.abbrCities(ctx, args.DstCities, args.Lang)
		if err != nil {
			return "", fmt.Errorf("could not get abbreviated dst cities: %v", err)
		}
	}
	urlProto := &urlpb.Url{
		Flight:    serializeFlights(args),
		Travelers: serializeTravelers(args.Travelers),
		Class:     urlpb.Url_Class(args.Class),
		TripType:  serializeTripType(args.TripType),
	}

	tfs, err := proto.Marshal(urlProto)
	if err != nil {
		return "", fmt.Errorf("error during url serialization: %s", err)
	}

	return "https://www.google.com/travel/flights/search" +
		"?tfs=" + base64.RawURLEncoding.EncodeToString(tfs) +
		"&curr=" + args.Currency.String() +
		"&hl=" + args.Lang.String(), nil
}

func serializeTripType(tripType TripType) urlpb.Url_TripType {
	switch tripType {
	case RoundTrip:
		return urlpb.Url_ROUND_TRIP
	case OneWay:
		return urlpb.Url_ONE_WAY
	case MultiCity:
		return urlpb.Url_MULTI_CITY
	default:
		return urlpb.Url_UNSPECIFIED_TRIP
	}
}
