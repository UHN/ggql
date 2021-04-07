package ggql_test

import (
	"log"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func benchmarkResolveExecutable(root *ggql.Root, b *testing.B) {
	ggql.Sort = false
	exe, err := root.ParseExecutableString(`{__type(name: "Artist"){name} artists{songs{name}}}`)
	if err != nil {
		b.Fatalf("Parse executable failed: %s\n", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := root.ResolveExecutable(exe, "", nil); err != nil {
			b.Errorf("Resolve failed: %s\n", err)
		}
	}
}

func BenchmarkResolveExecutableReflection(b *testing.B) {
	// Benchmark GGQL resolving using the reflection approach.
	schema := setupReflectSongs()
	root := ggql.NewRoot(schema)
	if err := root.AddTypes(NewDateScalar()); err != nil {
		log.Fatalf("no error should be returned when adding a Date type. %s", err)
	}
	if err := root.ParseString(songsSdl); err != nil {
		log.Fatalf("no error should be returned when parsing a valid SDL. %s", err)
	}
	benchmarkResolveExecutable(root, b)
}

func BenchmarkResolveExecutableInterface(b *testing.B) {
	// Benchmark GGQL resolving using the interface approach.
	schema := setupSongs()
	root := ggql.NewRoot(schema)
	if err := root.AddTypes(NewDateScalar()); err != nil {
		log.Fatalf("no error should be returned when adding a Date type. %s", err)
	}
	if err := root.ParseString(songsSdl); err != nil {
		log.Fatalf("no error should be returned when parsing a valid SDL. %s", err)
	}
	benchmarkResolveExecutable(root, b)
}

func BenchmarkResolveExecutableAnyResolver(b *testing.B) {
	// Benchmark GGQL resolving using the AnyResolver approach.
	root, err := setupAnySongs()
	if err != nil {
		log.Fatalf("setupAnySongs failed: %s", err)
	}
	benchmarkResolveExecutable(root, b)
}
