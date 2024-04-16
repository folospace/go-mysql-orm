package orm

func sliceContain[T comparable](a []T, b T) bool {
    return sliceContainIndex(a, b) > -1
}

func sliceContainIndex[T comparable](a []T, b T) int {
    for k, v := range a {
        if v == b {
            return k
        }
    }
    return -1
}
