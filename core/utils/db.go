package utils

import "gorm.io/gorm"

type Db = gorm.DB
type NewDb func(name string) *Db
