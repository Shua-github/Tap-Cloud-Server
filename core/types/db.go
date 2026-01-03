package types

import "gorm.io/gorm"

type NewDb func(name string) *gorm.DB
