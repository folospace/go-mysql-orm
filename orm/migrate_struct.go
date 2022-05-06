package orm


func (m *Query) MigrateToStruct() error {

    var createSql string

    m.Select(&createSql, "show create table")


    return nil
}