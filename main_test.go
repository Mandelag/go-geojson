package main

import "testing"

func Test_TestTestGeometry(t *testing.T) {
	type args struct {
		geo GeoJSON
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test simple geometry",
			args: args{
				geo: GeoJSON{
					Features: []MultiPolygonFeature{
						{
							Properties: map[string]string{"spontan": "uhuy"},
							Geometry: MultiPolygon{
								Type: "MultiPolygon",
								Coordinates: [][][][2]float32{
									{
										{
											{-6.0, 106.7},
											{-6.0, 106.9},
											{-6.1, 106.9},
											{-6.1, 106.7},
										},
									},
									{
										{
											{-6.3, 106.7},
											{-6.4, 106.7},
											{-6.4, 106.9},
											{-6.3, 106.9},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Test 2 feature",
			args: args{
				geo: GeoJSON{
					Features: []MultiPolygonFeature{
						{
							Properties: map[string]string{"spontan": "uhuy"},
							Geometry: MultiPolygon{
								Type: "MultiPolygon",
								Coordinates: [][][][2]float32{
									{ // polygon
										{ // the linear ring
											{-6.0, 106.7},
											{-6.0, 106.9},
											{-6.1, 106.9},
											{-6.1, 106.7},
										},
										// other linear ring (holes)
									},
									{ // polygon 2
										{ // the linear ring
											{-6.3, 106.7},
											{-6.4, 106.7},
											{-6.4, 106.9},
											{-6.3, 106.9},
										},
									},
								},
							},
						},
						{
							Properties: map[string]string{"spontan": "oh may good"},
							Geometry: MultiPolygon{
								Type: "MultiPolygon",
								Coordinates: [][][][2]float32{
									{
										{
											{-7.0, 106.7},
											{-7.0, 106.9},
											{-7.1, 106.9},
											{-7.1, 106.7},
										},
									},
									{
										{
											{-7.3, 106.7},
											{-7.4, 106.7},
											{-7.4, 106.9},
											{-7.3, 106.9},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TestTestGeometry(tt.args.geo)
		})
	}
}
