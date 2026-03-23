package store

type scannable interface {
	Scan(dest ...any) error
}
