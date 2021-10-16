package model

type StudentMoney struct {
	ID            int32  `gorm:"primarykey"`
	IdempotenceId string `gorm:"index:idx_idempotence_id;unique;type:varchar(36);not null"`
	Name          string `gorm:"type:varchar(12);not null"`
	Age           int    `gorm:"column:age;default:0;type:int"`
	Money         int    `gorm:"column:money;default:1;type:int"`
}
