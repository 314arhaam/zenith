package core

import (
	"fmt"
	"testing"
)

// The identical harness is used for baseline and optimized measurements.
func BenchmarkSystemMarshal(b *testing.B) {
	for _, size := range []int{10, 1000, 10000} {
		b.Run(fmt.Sprintf("read_only_%d", size), func(b *testing.B) {
			system := NewSystem()
			for i := 0; i < size; i++ {
				system.Set(fmt.Sprintf("service-%05d", i), NewCustomService(uint64(i+1), "2026-01-01T00:00:00Z"))
			}
			if _, err := system.Marshal(); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := system.Marshal(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	b.Run("write_then_read_1000", func(b *testing.B) {
		system := NewSystem()
		for i := 0; i < 1000; i++ {
			system.Set(fmt.Sprintf("service-%05d", i), NewCustomService(uint64(i+1), "2026-01-01T00:00:00Z"))
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			system.Set("service-00000", NewCustomService(uint64(i+1001), "2026-01-01T00:00:00Z"))
			if _, err := system.Marshal(); err != nil {
				b.Fatal(err)
			}
		}
	})
}
