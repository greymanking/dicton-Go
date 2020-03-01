// models
package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Task struct {
	ID       uint   `gorm:"primary_key" json:"taskid"`
	Keys     string `json:"keys"`
	Info     string `json:"info"`
	Audio    string `json:"audio"`
	Grade    byte
	Phonetic string `json:"phonetic"`
}

type User struct {
	gorm.Model
	Name     string `json:"name" gorm:"unique_index"`
	Password string `json:"password"`
	GroupId  byte
	Records  []Record
	Outlines []Outline
}

type Record struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	Task      Task `gorm:"foreignkey:TaskID"`
	TaskID    uint `json:"taskid"`
	User      User `gorm:"foreignkey:UserID"`
	UserID    uint
	Status    uint32
	TotalF    int
	TotalS    int
	TotalA    int
	Outdated  bool
}

type Outline struct {
	ID            uint `gorm:"primary_key"`
	CreatedAt     time.Time
	User          User `gorm:"foreignkey:UserID"`
	UserID        uint
	Coins         uint
	Diamonds      uint
	TotalDiamonds uint
	TotalCoins    uint
	Outdated      bool
}

const (
	DictFalse = 1 << iota
	DictSuccess
	PuzzleFalse
	PuzzleSuccess
)
