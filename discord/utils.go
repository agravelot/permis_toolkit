package discord

import (
	"github.com/agravelot/permis_toolkit/ornikar"
)

// TODO Make it generic for any type
func Chunk(lessons []ornikar.InstructorNextLessonsInterval, size int) [][]ornikar.InstructorNextLessonsInterval {
	chunks := make([][]ornikar.InstructorNextLessonsInterval, 0)
	for i := 0; i < len(lessons); i += size {
		end := i + size
		// If case of last chunk
		if i+size > len(lessons) {
			chunks = append(chunks, lessons[i:])
			continue
		}
		chunks = append(chunks, lessons[i:end])
	}
	return chunks
}
