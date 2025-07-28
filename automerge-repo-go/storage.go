package repo

// StorageAdapter is the interface for custom storage implementations.
type StorageAdapter interface {
	Load(id DocumentID) (*Document, error)
	Save(doc *Document) error
	Compact(doc *Document) error
	List() ([]DocumentID, error)
}
