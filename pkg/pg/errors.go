package pg

import (
	stderr "errors"

	"github.com/toxanetoxa/dating-backend/pkg/errors"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

var (
	ErrUnclassified        = errors.New("unclassified database error")
	ErrEntityAlreadyExists = errors.New("entity already exists")
	ErrEntityDoesntExist   = errors.New("entity doesn't exist")
	ErrIncompatibleEntity  = errors.New("incompatible entity")
)

func ProcessDbError(errPtr *error) {
	if errPtr == nil || *errPtr == nil {
		return
	}

	err := *errPtr
	if pqErr, ok := err.(*pgconn.PgError); ok {
		switch pqErr.Code {
		//pq errors code corresponding duplicate primary key
		case "23505":
			*errPtr = errors.Wrap(ErrEntityAlreadyExists, err.Error())
			return
		case "23503":
			*errPtr = errors.Wrap(ErrIncompatibleEntity, err.Error())
			return
		}
	}

	if stderr.Is(err, gorm.ErrRecordNotFound) {
		*errPtr = errors.Wrap(ErrEntityDoesntExist, err.Error())
		return
	}

	if stderr.Is(err, gorm.ErrDuplicatedKey) {
		*errPtr = errors.Wrap(ErrEntityAlreadyExists, err.Error())
		return
	}

	*errPtr = errors.Wrap(ErrUnclassified, err.Error())
}
