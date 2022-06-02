package orm

import "github.com/mcuadros/go-defaults"

func FillDefaults(table Table) {
    defaults.SetDefaults(table)
}
