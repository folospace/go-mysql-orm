package orm

////dest: *int | *int64 | ...
//func (m Query[T]) SelectCount(dest interface{}) QueryResult {
//    return m.Select(dest, "count(*)")
//}
//
////dest: *int | *string | ...
//func (m Query[T]) SelectValueOfFirstCell(dest interface{}, columns ...interface{}) QueryResult {
//    return m.Select(dest, columns...)
//}
//
////dest: *[]int | *[]string | ...
//func (m Query[T]) SelectSliceOfColumn1(dest interface{}, columns ...interface{}) QueryResult {
//    return m.Select(dest, columns...)
//}
//
////dest: *struct
//func (m Query[T]) SelectStructOfRow1(dest interface{}, columns ...interface{}) QueryResult {
//    return m.Select(dest, columns...)
//}
//
////dest: *[]struct
//func (m Query[T]) SelectSliceOfStruct(dest interface{}, columns ...interface{}) QueryResult {
//    return m.Select(dest, columns...)
//}
//
////dest: *map [int | string | ...] struct
//func (m Query[T]) SelectMapOfStructKeyByColumn1(dest interface{}, columns ...interface{}) QueryResult {
//    return m.Select(dest, columns...)
//}
//
////dest: *map [int | string | ...] []struct
//func (m Query[T]) SelectMapOfStructSliceKeyByColumn1(dest interface{}, columns ...interface{}) QueryResult {
//    return m.Select(dest, columns...)
//}
//
////dest: *map [int | string | ...] int | string ...
//func (m Query[T]) SelectMapOfColumn2KeyByColumn1(dest interface{}, columns ...interface{}) QueryResult {
//    return m.Select(dest, columns...)
//}
//
////dest, first of sql rows: *map[column_name]column_value
//func (m Query[T]) SelectMapStr2Interface(dest *map[string]interface{}, columns ...interface{}) QueryResult {
//    return m.Select(dest, columns...)
//}
//
////dest, sql rows: *[]map[column_name]column_value
//func (m Query[T]) SelectSliceOfMapStr2Interface(dest *[]map[string]interface{}, columns ...interface{}) QueryResult {
//    return m.Select(dest, columns...)
//}

func (m Query[T]) SelectForUpdate(columns ...interface{}) Query[T] {
    m.columns = columns
    return m
}

func (m Query[T]) Select(columns ...interface{}) Query[T] {
    m.columns = columns
    return m
}
