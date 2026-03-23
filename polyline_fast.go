package polyline

// This file contains optimized implementations for 2D coordinates using Coord2D.
// The Coord2D type uses fixed-size arrays to eliminate per-coordinate allocations.

// decodeCoordsD2Array is an optimized decoder for 2D coordinates using fixed-size arrays.
// This eliminates per-coordinate allocations, providing 2-3x faster decoding.
func decodeCoordsD2Array(buf []byte, scale float64) ([]Coord2D, []byte, error) {
	if len(buf) == 0 {
		return nil, buf, nil
	}

	// Estimate capacity: typical polyline is ~5 bytes per value, 2 values per coord
	estimatedCoords := max(len(buf)/10+1, 8)
	coords := make([]Coord2D, 0, estimatedCoords)

	// Decode first coordinate
	lat, buf, err := DecodeInt(buf)
	if err != nil {
		return nil, nil, err
	}
	lon, buf, err := DecodeInt(buf)
	if err != nil {
		return nil, nil, err
	}
	coords = append(coords, Coord2D{float64(lat) / scale, float64(lon) / scale})

	// Decode remaining coordinates (delta-encoded)
	for len(buf) > 0 {
		dLat, remaining, err := DecodeInt(buf)
		if err != nil {
			return nil, nil, err
		}
		dLon, remaining, err := DecodeInt(remaining)
		if err != nil {
			return nil, nil, err
		}
		buf = remaining

		lat += dLat
		lon += dLon
		coords = append(coords, Coord2D{float64(lat) / scale, float64(lon) / scale})
	}

	return coords, nil, nil
}

// encodeCoordsD2Array is an optimized encoder for 2D coordinates using fixed-size arrays.
func encodeCoordsD2Array(buf []byte, coords []Coord2D, scale float64) []byte {
	if len(coords) == 0 {
		return buf
	}

	// Pre-allocate buffer: estimate ~5 bytes per coordinate value
	if cap(buf)-len(buf) < len(coords)*10 {
		newBuf := make([]byte, len(buf), len(buf)+len(coords)*10)
		copy(newBuf, buf)
		buf = newBuf
	}

	var lastLat, lastLon int

	for _, coord := range coords {
		lat := round(scale * coord[0])
		lon := round(scale * coord[1])

		buf = EncodeInt(buf, lat-lastLat)
		buf = EncodeInt(buf, lon-lastLon)

		lastLat = lat
		lastLon = lon
	}

	return buf
}
