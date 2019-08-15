package service

import (
	"context"

	memdb "github.com/hashicorp/go-memdb"
)

type StorageService interface {
	SaveTransaction(ctx context.Context, trx *Transaction) error
	SaveWorkflow(ctx context.Context, wf *Workflow) error
	RemoveWorkflow(ctx context.Context, wf *Workflow) error
	LoadAllWorkflows(ctx context.Context) ([]*Workflow, error)
	LoadActiveTransactions(ctx context.Context) ([]*Transaction, error)
}

type DebugStorage struct {
	db *memdb.MemDB
}

var _ StorageService = &DebugStorage{}

func NewDebugStorage(workflows []*Workflow, transactions []*Transaction) (*DebugStorage, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"workflow": &memdb.TableSchema{
				Name: "workflow",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
				},
			},
			"transaction": &memdb.TableSchema{
				Name: "transaction",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}

	storage := &DebugStorage{
		db: db,
	}

	txn := db.Txn(true)
	for _, wf := range workflows {
		txn.Insert("workflows", wf)
	}

	for _, trx := range transactions {
		txn.Insert("transaction", trx)
	}

	txn.Commit()
	return storage, nil
}

func (storage *DebugStorage) SaveTransaction(ctx context.Context, trx *Transaction) error {
	transact := storage.db.Txn(true)
	if err := transact.Insert("transaction", trx); err != nil {
		transact.Abort()
		return err
	}

	transact.Commit()
	return nil
}

func (storage *DebugStorage) SaveWorkflow(ctx context.Context, wf *Workflow) error {
	transact := storage.db.Txn(true)
	if err := transact.Insert("workflow", wf); err != nil {
		transact.Abort()
		return err
	}

	transact.Commit()
	return nil
}

func (storage *DebugStorage) LoadAllWorkflows(ctx context.Context) ([]*Workflow, error) {
	result := make([]*Workflow, 0)
	transact := storage.db.Txn(true)
	it, err := transact.Get("workflows", "id")
	if err != nil {
		transact.Abort()
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		wf := obj.(*Workflow)
		result = append(result, wf)
	}

	return result, nil
}

func (storage *DebugStorage) LoadActiveTransactions(ctx context.Context) ([]*Transaction, error) {
	result := make([]*Transaction, 0)
	transact := storage.db.Txn(true)
	it, err := transact.Get("transaction", "id")
	if err != nil {
		transact.Abort()
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		trx := obj.(*Transaction)
		if !trx.IsCompleted() {
			result = append(result, trx)
		}
	}

	return result, nil
}

func (storage *DebugStorage) RemoveWorkflow(ctx context.Context, wf *Workflow) error {
	transact := storage.db.Txn(true)
	iwf, err := transact.First("workflow", "id", wf.Name)
	if err == nil && iwf != nil {
		err = transact.Delete("workflow", iwf)
	}

	if err != nil {
		transact.Abort()
		return err
	}

	transact.Commit()
	return nil
}
