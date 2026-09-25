package projects

import "github.com/jackc/pgx/v5/pgxpool"

type Genre struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type GenreModel struct {
	db *pgxpool.Pool
}

func NewGenreModel(db *pgxpool.Pool) *GenreModel {
	return &GenreModel{db: db}
}
