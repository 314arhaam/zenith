package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"zenith/core"
)

// Includes response construction and writing the full JSON payload, not TCP.
func BenchmarkStatusAll(b *testing.B) {
	for _, size := range []int{10, 1000, 10000} {
		b.Run(fmt.Sprintf("services_%d", size), func(b *testing.B) {
			h := NewHandler()
			for i := 0; i < size; i++ {
				h.Core.Set(fmt.Sprintf("service-%05d", i), core.NewCustomService(uint64(i+1), "2026-01-01T00:00:00Z"))
			}
			req := httptest.NewRequest(http.MethodGet, "/status", nil)
			h.Status(httptest.NewRecorder(), req)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rec := httptest.NewRecorder()
				h.Status(rec, req)
				if rec.Code != 200 {
					b.Fatal(rec.Code)
				}
			}
		})
	}
}
