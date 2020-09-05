package storage

import (
	"encoding/binary"
	"errors"
	"github.com/dgraph-io/badger/v2"
	"github.com/pingcap/failpoint"
	"github.com/speps/go-hashids"
	"go.uber.org/zap"
	"math"
)

var (
	ErrShortNotExist = errors.New("short form does not exist")
	ErrInvalidShort  = errors.New("invalid short form")
)

// Storage defines fields used in db interaction process
type Storage struct {
	logger *zap.Logger
	db     *badger.DB
	seq    *badger.Sequence
	hashID *hashids.HashID
}

// New constructs Storage instance with provided path and default badger options
func New(logger *zap.Logger, path string) (*Storage, error) {
	if logger == nil {
		return nil, errors.New("no logger provided")
	}

	db, err := badger.Open(badger.DefaultOptions(path))
	failpoint.Inject("openDatabaseErr", func() {
		err = errors.New("mock open database error")
	})
	if err != nil {
		logger.Error("opening database", zap.String("path", path), zap.Error(err))
		if db != nil {
			logger.Info("trying to close database")
			_ = db.Close()
		}
		return nil, err
	}

	seq, err := db.GetSequence([]byte("seq"), 100)
	failpoint.Inject("getSequenceErr", func() {
		err = errors.New("mock get sequence error")
	})
	if err != nil {
		logger.Error("retrieving sequence", zap.String("key", "seq"), zap.Error(err))
		logger.Info("closing database")
		_ = db.Close()
		return nil, err
	}

	data := hashids.NewData()
	data.MinLength = 7

	hashID, err := hashids.NewWithData(data)
	failpoint.Inject("newWithDataErr", func() {
		err = errors.New("mock NewWithData error")
	})
	if err != nil {
		logger.Error("generating new hashID", zap.Error(err))
		logger.Info("releasing sequence")
		_ = seq.Release()
		logger.Info("closing database")
		_ = db.Close()
		return nil, err
	}

	return &Storage{
		logger: logger,
		db:     db,
		seq:    seq,
		hashID: hashID,
	}, err
}

// Close releases sequence and closes database
func (s *Storage) Close() error {
	s.logger.Info("closing storage")
	err := s.seq.Release()
	failpoint.Inject("releaseSequenceOnCloseErr", func() {
		err = errors.New("mock release sequence error")
	})
	if err != nil {
		s.logger.Error("releasing sequence", zap.Error(err))
		s.logger.Warn("trying to close database anyway")
		_ = s.db.Close()
		return err
	}
	err = s.db.Close()
	failpoint.Inject("closeDatabaseErr", func() {
		err = errors.New("mock close database error")
	})
	if err != nil {
		s.logger.Error("closing database", zap.Error(err))
		return err
	}

	s.logger.Info("storage closed")

	return nil
}

// SaveURL returns short unique string ID for provided URL
func (s *Storage) SaveURL(reqID uint64, url string) (string, error) {
	logger := s.logger.With(zap.Uint64("request id", reqID))

	id, err := s.seq.Next()
	failpoint.Inject("nextIDErr", func() {
		err = errors.New("mock next ID error")
	})
	if err != nil {
		logger.Error("retrieving next id for url", zap.Error(err))
		return "", err
	}

	// skip error handing due to impossible condition
	// EncodeInt64 inside checks if provided int64 slice is not empty and holds values grater or equal zero
	// uint64ToInt64Slice by design can not return empty slice (compilation check)
	// uint64ToInt64Slice also can not hold negative values
	short, _ := s.hashID.EncodeInt64(uint64ToInt64Slice(id))
	//if err != nil {
	//	logger.Error("generating short form", zap.Error(err))
	//	return "", err
	//}

	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(utob(id), []byte(url))
	})
	failpoint.Inject("updateErr", func() {
		err = errors.New("mock update error")
	})
	if err != nil {
		logger.Error("updating database", zap.Error(err))
		return "", err
	}

	return short, nil
}

// GetURL returns URL that has been saved referenced by short string ID
func (s *Storage) GetURL(reqID uint64, short string) (string, error) {
	logger := s.logger.With(zap.Uint64("request id", reqID))

	ids, err := s.hashID.DecodeInt64WithError(short)
	if err != nil {
		return "", ErrInvalidShort
	}

	id := int64SliceToUint64(ids)

	var urlBytes []byte
	err = s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(utob(id))
		if err != nil {
			return err
		}

		urlBytes, err = item.ValueCopy(nil)
		failpoint.Inject("valueCopyErr", func() {
			err = errors.New("mock value copy error")
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return "", ErrShortNotExist
		}
		logger.Error("retrieving source URL", zap.Error(err))
		return "", err
	}

	return string(urlBytes), nil
}

// utob converts uint64 to byte slice
func utob(u uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, u)

	return buf
}

// uint64ToInt64Slice converts uint64 integer to slice of int64
// that slice after summing up via int64SliceToUint64 gives original uint64
func uint64ToInt64Slice(u uint64) []int64 {
	switch {
	case u == math.MaxUint64:
		return []int64{math.MaxInt64, math.MaxInt64, 1}
	case u > math.MaxInt64:
		return []int64{math.MaxInt64, int64(u - math.MaxInt64)}
	default:
		return []int64{int64(u)}
	}
}

// int64SliceToUint64 sums up each int64 in slice and returns corresponding uint64
func int64SliceToUint64(s []int64) uint64 {
	var u uint64 = 0
	for _, v := range s {
		u += uint64(v)
	}
	return u
}
