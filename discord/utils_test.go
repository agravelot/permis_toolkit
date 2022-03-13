package discord

import (
	"reflect"
	"testing"

	"github.com/agravelot/permis_toolkit/ornikar"
)

func TestChunk(t *testing.T) {
	type args struct {
		lessons []ornikar.InstructorNextLessonsInterval
		size    int
	}
	tests := []struct {
		name string
		args args
		want [][]ornikar.InstructorNextLessonsInterval
	}{
		{
			name: "Chunk 7 elements by 3",
			args: args{
				lessons: []ornikar.InstructorNextLessonsInterval{
					{ID: "1"},
					{ID: "2"},
					{ID: "3"},
					{ID: "4"},
					{ID: "5"},
					{ID: "6"},
					{ID: "7"},
				},
				size: 3,
			},
			want: [][]ornikar.InstructorNextLessonsInterval{
				{
					{ID: "1"},
					{ID: "2"},
					{ID: "3"},
				},
				{
					{ID: "4"},
					{ID: "5"},
					{ID: "6"},
				},
				{
					{ID: "7"},
				},
			},
		},
		{
			name: "Chunk 7 elements by 1",
			args: args{
				lessons: []ornikar.InstructorNextLessonsInterval{
					{ID: "1"},
					{ID: "2"},
					{ID: "3"},
					{ID: "4"},
					{ID: "5"},
					{ID: "6"},
					{ID: "7"},
				},
				size: 1,
			},
			want: [][]ornikar.InstructorNextLessonsInterval{
				{
					{ID: "1"},
				},
				{
					{ID: "2"},
				},
				{
					{ID: "3"},
				},
				{
					{ID: "4"},
				},
				{
					{ID: "5"},
				},
				{
					{ID: "6"},
				},
				{
					{ID: "7"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Chunk(tt.args.lessons, tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chunk() = %v, want %v", got, tt.want)
			}
		})
	}
}
