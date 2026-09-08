package models

import (
	"testing"
)

func TestValidateReviewRating(t *testing.T) {
	tests := []struct {
		name    string
		rating  int
		wantErr bool
	}{
		{"rating 1", 1, false},
		{"rating 2", 2, false},
		{"rating 3", 3, false},
		{"rating 4", 4, false},
		{"rating 5", 5, false},
		{"rating 0", 0, true},
		{"rating -1", -1, true},
		{"rating 6", 6, true},
		{"rating 100", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.rating >= 1 && tt.rating <= 5
			if valid == tt.wantErr {
				t.Errorf("rating %d: expected wantErr=%v, got %v", tt.rating, tt.wantErr, !valid)
			}
		})
	}
}

func TestReviewListResponse_Fields(t *testing.T) {
	resp := ReviewListResponse{
		Reviews: []Review{
			{ID: 1, Rating: 5, Comment: "Great"},
			{ID: 2, Rating: 4, Comment: "Good"},
		},
		AvgRating: 4.5,
		Count:     2,
		Pagination: Pagination{
			Page:       1,
			Limit:      10,
			Total:      2,
			TotalPages: 1,
		},
	}

	if len(resp.Reviews) != 2 {
		t.Errorf("expected 2 reviews, got %d", len(resp.Reviews))
	}
	if resp.AvgRating != 4.5 {
		t.Errorf("expected avg rating 4.5, got %f", resp.AvgRating)
	}
	if resp.Count != 2 {
		t.Errorf("expected count 2, got %d", resp.Count)
	}
	if resp.Pagination.TotalPages != 1 {
		t.Errorf("expected 1 total page, got %d", resp.Pagination.TotalPages)
	}
}

func TestCreateReviewRequest_Fields(t *testing.T) {
	req := CreateReviewRequest{
		ProductID: 1,
		Rating:    5,
		Comment:   "Excellent product!",
	}

	if req.ProductID != 1 {
		t.Errorf("expected product ID 1, got %d", req.ProductID)
	}
	if req.Rating != 5 {
		t.Errorf("expected rating 5, got %d", req.Rating)
	}
	if req.Comment != "Excellent product!" {
		t.Errorf("expected comment 'Excellent product!', got %s", req.Comment)
	}
}
