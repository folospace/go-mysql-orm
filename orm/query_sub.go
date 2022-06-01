package orm

func (m Query[T]) SubQuery() *tempTable {
    tempTable := m.generateSelectQuery(m.columns...)

    tempTable.db = m.db
    tempTable.tx = m.tx
    tempTable.dbName = m.tables[0].table.DatabaseName()
    tempTable.err = m.result.Err

    return &tempTable
}
