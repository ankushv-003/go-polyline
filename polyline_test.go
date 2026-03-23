package polyline_test

import (
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"testing/quick"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-polyline"
)

func TestUint(t *testing.T) {
	for i, tc := range []struct {
		u            uint
		s            string
		nonCanonical bool
	}{
		{u: 0, s: "?"},
		{u: 2, s: "a?", nonCanonical: true},
		{u: 2, s: "A"},
		{u: 31, s: "^"},
		{u: 32, s: "_@"},
		{u: 174, s: "mD"},
		{u: 18446744073709551614, s: "}~~~~~~~~~~~N"},
		{u: 18446744073709551615, s: "~~~~~~~~~~~~N"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, b, err := polyline.DecodeUint([]byte(tc.s))
			assert.NoError(t, err)
			assert.Equal(t, tc.u, got)
			assert.Equal(t, 0, len(b))
			if !tc.nonCanonical {
				assert.Equal(t, []byte(tc.s), polyline.EncodeUint(nil, tc.u))
			}
		})
	}
}

func TestDecodeErrors(t *testing.T) {
	for i, tc := range []struct {
		s         string
		err       error
		coordsErr error
	}{
		{s: "", err: polyline.ErrEmpty},
		{s: ">", err: polyline.ErrInvalidByte, coordsErr: polyline.ErrInvalidByte},
		{s: "\x80", err: polyline.ErrInvalidByte, coordsErr: polyline.ErrInvalidByte},
		{s: "_", err: polyline.ErrUnterminatedSequence, coordsErr: polyline.ErrUnterminatedSequence},
		{s: "~~~~~~~~~~~~", err: polyline.ErrUnterminatedSequence, coordsErr: polyline.ErrUnterminatedSequence},
		{s: "~~~~~~~~~~~~O", err: polyline.ErrOverflow, coordsErr: polyline.ErrOverflow},
		{s: "~~~~~~~~~~~~_", err: polyline.ErrOverflow, coordsErr: polyline.ErrOverflow},
		{s: "~~~~~~~~~~~~\x80", err: polyline.ErrInvalidByte, coordsErr: polyline.ErrInvalidByte},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, _, err := polyline.DecodeUint([]byte(tc.s))
			assert.Equal(t, tc.err, err)
			_, _, err = polyline.DecodeInt([]byte(tc.s))
			assert.Equal(t, tc.err, err)
			_, _, err = polyline.DecodeCoord([]byte(tc.s))
			assert.Equal(t, tc.err, err)
			_, _, err = polyline.DecodeCoords([]byte(tc.s))
			assert.Equal(t, tc.coordsErr, err)
			c := polyline.Codec{Dim: 1, Scale: 1e5}
			_, _, err = c.DecodeFlatCoords([]float64{0}, []byte(tc.s))
			assert.Equal(t, tc.coordsErr, err)
		})
	}
}

func TestMultidimensionalDecodeErrors(t *testing.T) {
	for i, tc := range []struct {
		s   string
		err error
	}{
		{s: "_p~iF~ps|U_p~iF>", err: polyline.ErrInvalidByte},
		{s: "_p~iF~ps|U_p~iF\x80", err: polyline.ErrInvalidByte},
		{s: "_p~iF~ps|U_p~iF~ps|", err: polyline.ErrUnterminatedSequence},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, _, err := polyline.DecodeCoords([]byte(tc.s))
			assert.Equal(t, tc.err, err)
			c := polyline.Codec{Dim: 2, Scale: 1e5}
			_, _, err = c.DecodeFlatCoords([]float64{0, 0}, []byte(tc.s))
			assert.Equal(t, tc.err, err)
		})
	}
}

func TestInt(t *testing.T) {
	for i, tc := range []struct {
		i int
		s string
	}{
		{i: 3850000, s: "_p~iF"},
		{i: -12020000, s: "~ps|U"},
		{i: -17998321, s: "`~oia@"},
		{i: 220000, s: "_ulL"},
		{i: -75000, s: "nnqC"},
		{i: 255200, s: "_mqN"},
		{i: -550300, s: "vxq`@"},
		{i: math.MaxInt64, s: "}~~~~~~~~~~~N"},
		{i: math.MaxInt64 - 1, s: "{~~~~~~~~~~~N"},
		{i: math.MinInt64 + 1, s: "|~~~~~~~~~~~N"},
		{i: math.MinInt64, s: "~~~~~~~~~~~~N"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, b, err := polyline.DecodeInt([]byte(tc.s))
			assert.NoError(t, err)
			assert.Equal(t, 0, len(b))
			assert.Equal(t, tc.i, got)
			assert.Equal(t, []byte(tc.s), polyline.EncodeInt(nil, tc.i))
		})
	}
}

func TestCoord(t *testing.T) {
	for i, tc := range []struct {
		s            string
		c            []float64
		nonCanonical bool
	}{
		{
			s: "_p~iF~ps|U",
			c: []float64{38.5, -120.2},
		},
		{
			s:            "a?Z",
			c:            []float64{1e-05, -0.00014},
			nonCanonical: true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, b, err := polyline.DecodeCoord([]byte(tc.s))
			assert.NoError(t, err)
			assert.Equal(t, 0, len(b))
			assert.Equal(t, tc.c, got)
			if !tc.nonCanonical {
				assert.Equal(t, []byte(tc.s), polyline.EncodeCoord(tc.c))
			}
		})
	}
}

func TestCoords(t *testing.T) {
	for i, tc := range []struct {
		cs [][]float64
		s  string
	}{
		{
			cs: [][]float64{{38.5, -120.2}, {40.7, -120.95}, {43.252, -126.453}},
			s:  "_p~iF~ps|U_ulLnnqC_mqNvxq`@",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, b, err := polyline.DecodeCoords([]byte(tc.s))
			assert.NoError(t, err)
			assert.Equal(t, 0, len(b))
			assert.Equal(t, tc.cs, got)
			assert.Equal(t, []byte(tc.s), polyline.EncodeCoords(tc.cs))
		})
	}
}

func TestFlatCoords(t *testing.T) {
	for i, tc := range []struct {
		fcs []float64
		s   string
	}{
		{
			fcs: []float64{38.5, -120.2, 40.7, -120.95, 43.252, -126.453},
			s:   "_p~iF~ps|U_ulLnnqC_mqNvxq`@",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			codec := polyline.Codec{Dim: 2, Scale: 1e5}
			gotFCS, b, err := codec.DecodeFlatCoords(nil, []byte(tc.s))
			assert.NoError(t, err)
			assert.Equal(t, 0, len(b))
			assert.Equal(t, tc.fcs, gotFCS)
			gotBytes, err := codec.EncodeFlatCoords(nil, tc.fcs)
			assert.NoError(t, err)
			assert.Equal(t, []byte(tc.s), gotBytes)
		})
	}
}

func TestFlatCoordsEmpty(t *testing.T) {
	codec := polyline.Codec{Dim: 2, Scale: 1e5}
	gotFCS, b, err := codec.DecodeFlatCoords(nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(b))
	assert.Equal(t, 0, len(gotFCS))
	gotBytes, err := codec.EncodeFlatCoords(nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(gotBytes))
}

func TestDecodeFlatCoordsErrors(t *testing.T) {
	for i, tc := range []struct {
		fcs []float64
		s   string
		err error
	}{
		{
			fcs: []float64{0},
			s:   "",
			err: polyline.ErrDimensionalMismatch,
		},
		{
			s:   "_p~iF",
			err: polyline.ErrEmpty,
		},
		{
			s:   "_p~iF~ps|U_p~iF",
			err: polyline.ErrEmpty,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			codec := polyline.Codec{Dim: 2, Scale: 1e5}
			_, _, err := codec.DecodeFlatCoords(tc.fcs, []byte(tc.s))
			assert.Equal(t, tc.err, err)
		})
	}
}

func TestEncodeFlatCoordErrors(t *testing.T) {
	for i, tc := range []struct {
		fcs []float64
		err error
	}{
		{
			fcs: []float64{0},
			err: polyline.ErrDimensionalMismatch,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			codec := polyline.Codec{Dim: 2, Scale: 1e5}
			_, err := codec.EncodeFlatCoords(nil, tc.fcs)
			assert.Equal(t, tc.err, err)
		})
	}
}

func TestCodec(t *testing.T) {
	for i, tc := range []struct {
		c  polyline.Codec
		cs [][]float64
		s  string
	}{
		{
			c:  polyline.Codec{Dim: 2, Scale: 1e5},
			cs: [][]float64{{38.5, -120.2}, {40.7, -120.95}, {43.252, -126.453}},
			s:  "_p~iF~ps|U_ulLnnqC_mqNvxq`@",
		},
		{
			c:  polyline.Codec{Dim: 2, Scale: 1e6},
			cs: [][]float64{{38.5, -120.2}, {40.7, -120.95}, {43.252, -126.453}},
			s:  "_izlhA~rlgdF_{geC~ywl@_kwzCn`{nI",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, b, err := tc.c.DecodeCoords([]byte(tc.s))
			assert.NoError(t, err)
			assert.Equal(t, tc.cs, got)
			assert.Equal(t, 0, len(b))
			assert.Equal(t, []byte(tc.s), tc.c.EncodeCoords(nil, tc.cs))
		})
	}
}

func float64ArrayWithin(a, b []float64, prec float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, xa := range a {
		if math.Abs(xa-b[i]) > prec {
			return false
		}
	}
	return true
}

type QuickCoords [][]float64

func (qc QuickCoords) Generate(r *rand.Rand, size int) reflect.Value {
	result := make([][]float64, size)
	for i := range result {
		result[i] = []float64{180*r.Float64() - 90, 360*r.Float64() - 180}
	}
	return reflect.ValueOf(result)
}

func TestCoordsQuick(t *testing.T) {
	f := func(qc QuickCoords) bool {
		buf := polyline.EncodeCoords([][]float64(qc))
		cs, buf, err := polyline.DecodeCoords(buf)
		if len(buf) != 0 || err != nil {
			return false
		}
		if len(cs) != len(qc) {
			return false
		}
		for i, c := range cs {
			if !float64ArrayWithin(c, qc[i], 5e-6) {
				return false
			}
		}
		return true
	}
	assert.NoError(t, quick.Check(f, nil))
}

type QuickFlatCoords []float64

func (qfc QuickFlatCoords) Generate(r *rand.Rand, size int) reflect.Value {
	result := make([]float64, 2*size)
	for i := range result {
		if i%2 == 0 {
			result[i] = 180*r.Float64() - 90
		} else {
			result[i] = 360*r.Float64() - 180
		}
	}
	return reflect.ValueOf(result)
}

func TestFlatCoordsQuick(t *testing.T) {
	f := func(fqc QuickFlatCoords) bool {
		codec := polyline.Codec{Dim: 2, Scale: 1e5}
		buf, err := codec.EncodeFlatCoords(nil, []float64(fqc))
		if err != nil {
			return false
		}
		fcs, _, err := codec.DecodeFlatCoords(nil, buf)
		if err != nil {
			return false
		}
		return float64ArrayWithin([]float64(fqc), fcs, 5e-6)
	}
	assert.NoError(t, quick.Check(f, nil))
}

func FuzzDecodeCoords(f *testing.F) {
	f.Add([]byte("_p~iF~ps|U"))
	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _, _ = polyline.DecodeCoords(data)
	})
}

func TestCoord2D(t *testing.T) {
	for i, tc := range []struct {
		coords []polyline.Coord2D
		str    string
	}{
		{
			coords: []polyline.Coord2D{{38.5, -120.2}},
			str:    "_p~iF~ps|U",
		},
		{
			coords: []polyline.Coord2D{{38.5, -120.2}, {40.7, -120.95}, {43.252, -126.453}},
			str:    "_p~iF~ps|U_ulLnnqC_mqNvxq`@",
		},
		{
			coords: []polyline.Coord2D{{0, 0}},
			str:    "??",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// Test encoding
			got := polyline.EncodeCoords2D(tc.coords)
			assert.Equal(t, []byte(tc.str), got)

			// Test decoding
			coords, remaining, err := polyline.DecodeCoords2D([]byte(tc.str))
			assert.NoError(t, err)
			assert.Equal(t, 0, len(remaining))
			assert.Equal(t, len(tc.coords), len(coords))
			for j := range coords {
				assert.Equal(t, tc.coords[j][0], coords[j][0], "lat mismatch at index %d", j)
				assert.Equal(t, tc.coords[j][1], coords[j][1], "lon mismatch at index %d", j)
			}
		})
	}
}

func TestCoord2DCodec(t *testing.T) {
	c := polyline.Codec{Dim: 2, Scale: 1e5}
	coords := []polyline.Coord2D{{38.5, -120.2}, {40.7, -120.95}}

	// Encode
	encoded := c.EncodeCoords2D(nil, coords)
	assert.Equal(t, len(encoded) > 0, true)

	// Decode
	decoded, remaining, err := c.DecodeCoords2D(encoded)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(remaining))
	assert.Equal(t, len(coords), len(decoded))

	// Verify values
	for i := range coords {
		assert.Equal(t, coords[i][0], decoded[i][0])
		assert.Equal(t, coords[i][1], decoded[i][1])
	}
}

func TestCoord2DEmpty(t *testing.T) {
	coords, remaining, err := polyline.DecodeCoords2D(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(coords))
	assert.Equal(t, 0, len(remaining))

	encoded := polyline.EncodeCoords2D(nil)
	assert.Equal(t, 0, len(encoded))
}

func TestCoord2DWrongDimension(t *testing.T) {
	c := polyline.Codec{Dim: 3, Scale: 1e5}
	_, _, err := c.DecodeCoords2D([]byte("??"))
	assert.Equal(t, polyline.ErrDimensionalMismatch, err)
}

func TestCoord2DVsSlice(t *testing.T) {
	// Test that Coord2D and [][]float64 produce identical results
	testData := [][]float64{
		{38.5, -120.2},
		{40.7, -120.95},
		{43.252, -126.453},
		{41.139, -104.844},
	}

	// Convert to Coord2D
	coordsArray := make([]polyline.Coord2D, len(testData))
	for i, coord := range testData {
		coordsArray[i] = polyline.Coord2D{coord[0], coord[1]}
	}

	// Encode both
	encodedSlice := polyline.EncodeCoords(testData)
	encodedArray := polyline.EncodeCoords2D(coordsArray)

	// Should produce identical output
	assert.Equal(t, encodedSlice, encodedArray)

	// Decode both
	decodedSlice, _, err := polyline.DecodeCoords(encodedSlice)
	assert.NoError(t, err)
	decodedArray, _, err := polyline.DecodeCoords2D(encodedArray)
	assert.NoError(t, err)

	// Should have same values
	assert.Equal(t, len(decodedSlice), len(decodedArray))
	for i := range decodedSlice {
		assert.Equal(t, decodedSlice[i][0], decodedArray[i][0])
		assert.Equal(t, decodedSlice[i][1], decodedArray[i][1])
	}
}
