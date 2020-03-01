package main

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var dbg *gorm.DB

const (
	AllSorts = iota
	PuzzleDictation
	OnlyDictation
)

func init() {
	db, err := gorm.Open("sqlite3", "data.sqlite3")
	if err != nil {
		panic("failed to connect database, reason:" + err.Error()) //todo:defer仍会执行
	}
	dbg = db

	Initialize()
}

func CloseDB() { dbg.Close() }

func Initialize() {
	// Migrate the schema
	dbg.AutoMigrate(&Task{})
	dbg.AutoMigrate(&User{})
	dbg.AutoMigrate(&Record{})
	dbg.AutoMigrate(&Outline{})
}
func AddUser(usr User) error {
	return dbg.Create(&usr).Error
}

//返回user的id，须保证id不为0
func CheckUser(name, password string) uint {
	var user User
	if err := dbg.Where("name = ? AND password= ?", name, password).
		First(&user).Error; err != nil {
		return 0
	}
	return user.ID
}

func FindUser(name string) uint {
	var user User
	if err := dbg.Where("name = ?", name).First(&user).Error; err != nil {
		return 0
	}
	return user.ID
}

func QueryUserID(name string) (uint, error) {
	var user User
	res := dbg.Where("name=?", name).First(&user)
	return user.ID, res.Error
}

//创建Records，需要recs中有TaskID
func AddRecords(userId uint, recs []Record) error {
	tx := dbg.Begin()

	for _, rec := range recs {
		rec.UserID = userId
		rec.CreatedAt = time.Now()
		rec.Outdated = false

		if err := tx.Create(&rec).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

type RecordIn struct {
	LastRecordID uint   `json:"lastrec"`
	TaskID       uint   `json:"taskid"`
	Status       uint32 `json:"status"`
}

func CommitRecordIns(userId uint, recIns []RecordIn) error {
	tx := dbg.Begin()

	for _, recIn := range recIns {

		var rec Record

		if recIn.LastRecordID != 0 {
			if err := tx.Where("id=?", recIn.LastRecordID).First(&rec).Error; err != nil {
				fmt.Println(err)
				tx.Rollback()
				return err
			}
			fmt.Println(rec)
			rec.Outdated = true
			if err := tx.Save(&rec).Error; err != nil {
				fmt.Println("last rec update fail")

				tx.Rollback()
				return err
			}
		}

		rec.ID = 0 //auto increment
		rec.UserID = userId
		rec.CreatedAt = time.Now()
		rec.Outdated = false
		rec.Status = recIn.Status
		rec.TaskID = recIn.TaskID

		if recIn.Status&DictFalse == DictFalse || recIn.Status&PuzzleFalse == PuzzleFalse {
			rec.TotalF++
		} else {
			rec.TotalS++
		}

		rec.TotalA = rec.TotalF + rec.TotalS

		if err := tx.Create(&rec).Error; err != nil {
			fmt.Println("new rec create fail")
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func TotalLearned(userId uint) uint {
	max := struct{ M uint }{}

	if e := dbg.Raw("SELECT MAX(task_id) AS m FROM records WHERE user_id=?", userId).
		Scan(&max).Error; e != nil {
		fmt.Println(e)
	}
	return max.M
}

type TaskOut struct {
	Task
	LastRec uint `json:"lastrec"`
	Kind    byte `json:"kind"`
}

func GetTasksConditional(userid uint, limit uint, totalStmt string, kind byte,
	before time.Time) []TaskOut {
	var recs []Record
	var tasks []TaskOut

	dbg.Limit(limit).Where("user_id=? AND "+totalStmt+
		" AND NOT outdated AND created_at<=?", userid, before).Preload("Task").Find(&recs)

	for _, rec := range recs {
		tasks = append(tasks, TaskOut{rec.Task, rec.ID, kind})
	}

	return tasks
}

func LoadTasks(userId uint, workload, learned uint) []TaskOut {
	var alltasks, curtasks []TaskOut

	rest := workload
	fmt.Println("workload", workload)

	//学习过一次，学第二次时每个任务工作量为4
	curtasks = GetTasksConditional(userId, rest/4, "total_a=1",
		PuzzleDictation, time.Now())
	alltasks = append(alltasks, curtasks...)
	rest = rest - uint(len(curtasks)*4)

	//学第三次时每个任务工作量为2
	if rest > 0 {
		curtasks = GetTasksConditional(userId, rest/2, "total_f>0 AND total_a=2",
			OnlyDictation, time.Now())
		alltasks = append(alltasks, curtasks...)
		rest = rest - uint(len(curtasks)*2)
	} else {
		return alltasks
	}

	//学第四次时每个任务工作量为2，且上次学习在六天前
	if rest > 0 {
		dur, _ := time.ParseDuration("-144h")
		curtasks = GetTasksConditional(userId, rest/2, "total_f>0 AND total_a=3",
			OnlyDictation, time.Now().Add(dur))
		alltasks = append(alltasks, curtasks...)
		rest = rest - uint(len(curtasks)*2)
	} else {
		return alltasks
	}

	virneeds := rest / 5 //剩下的任务量需要多少空白任务

	if virneeds > 0 {
		var tasks []Task

		if learned <= 0 {
			dbg.Limit(virneeds).Find(&tasks)
		} else {
			dbg.Offset(learned).Limit(virneeds).Find(&tasks)
		}

		//装载新任务
		for _, task := range tasks {
			alltasks = append(alltasks, TaskOut{Task: task, Kind: AllSorts})
		}
	}
	return alltasks
}

func QueryAssets(userid uint) (uint, uint) {
	var ol Outline
	if err := dbg.Where("user_id=? AND NOT outdated", userid).Last(&ol).Error; err != nil {
		fmt.Println(err)
	}
	fmt.Println("ol", ol)
	return ol.TotalCoins, ol.TotalDiamonds
}

func UpdateAssets(userid, coins, diamonds uint) error {
	var ol Outline
	if err := dbg.Where("user_id=? AND NOT outdated", userid).Last(&ol).Error; err == nil {
		ol.Outdated = true
		dbg.Save(&ol)
	}

	ol.ID = 0 //auto increment
	ol.UserID = userid
	ol.Coins = coins
	ol.Diamonds = diamonds
	ol.CreatedAt = time.Now()
	ol.Outdated = false
	ol.TotalCoins = ol.TotalCoins + coins
	ol.TotalDiamonds = ol.TotalDiamonds + diamonds

	return dbg.Create(&ol).Error
}

func Fallible(userid uint, workload uint) []TaskOut {
	var tasks []TaskOut
	var recs []Record

	dbg.Limit(workload/4).Order("total_a, total_f desc").Where("user_id=? AND NOT outdated AND total_f>0",
		userid).Preload("Task").Find(&recs)

	for _, rec := range recs {
		tasks = append(tasks, TaskOut{rec.Task, rec.ID, PuzzleDictation})
	}

	return tasks
}
