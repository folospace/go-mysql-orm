package orm

import "errors"

var (
    ErrDbNotSelected                    = errors.New("db not selecteed")
    ErrTableNotExisted                  = errors.New("table not existed")
    ErrTableNotSelected                 = errors.New("table not selected")
    ErrColumnNotSelected                = errors.New("column not selected")
    ErrColumnNotExisted                 = errors.New("column not existed")
    ErrRawSqlRequired                   = errors.New("raw sql required")
    ErrParamMustBePtr                   = errors.New("param must be ptr")
    ErrParamElemKindMustBeStruct        = errors.New("param elem kind must be struct")
    ErrColumnShouldBeStringOrPtr        = errors.New("select|where column should be string or ptr of Table.T.field")
    ErrDestOfGetToMustBePtr             = errors.New("dest of Get-to must be ptr")
    ErrDestOfGetToSliceElemMustNotBePtr = errors.New("dest of Get-to slice elem kind must not be ptr")
    ErrDestOfGetToMapElemMustNotBePtr   = errors.New("dest of Get-to map elem kind must not be ptr")
    ErrInsertPtrNotAllowed              = errors.New("insert ptr data not allowed")
    ErrUpdateWithoutCondition           = errors.New("update without condition not allowed")
    ErrDeleteWithoutCondition           = errors.New("delete without condition not allowed")
)
