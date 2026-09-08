package models

import "time"

type Review struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	UserID    int       `json:"user_id"`
	UserName  string    `json:"user_name"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type ReviewListResponse struct {
	Reviews    []Review   `json:"reviews"`
	AvgRating  float64    `json:"avg_rating"`
	Count      int        `json:"count"`
	Pagination Pagination `json:"pagination"`
}

type CreateReviewRequest struct {
	ProductID int    `json:"product_id"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
}
