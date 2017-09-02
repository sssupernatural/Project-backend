package dataClient

import (
	"github.com/go-pg/pg"
	"external/comm"
	"github.com/go-pg/pg/orm"
)

const (
	RetRecordNotFound string = "pg: no rows in result set"
)

type DataCenterClient struct {
	db *pg.DB
}

type DataCenterDesc struct {
	Addr string
	User string
	Password string
	Database string
}

func New(dataCenter *DataCenterDesc) *DataCenterClient {

	return &DataCenterClient{
		db: pg.Connect(&pg.Options{
			Addr: dataCenter.Addr,
			User: dataCenter.User,
			Password: dataCenter.Password,
			Database: dataCenter.Database,
		}),
	}
}

func (c *DataCenterClient)InitUsersData() error {

	err := c.db.CreateTable(&comm.UserCheckInfo{}, &orm.CreateTableOptions{IfNotExists: true})
	if err != nil {
		return err
	}

	err = c.db.CreateTable(&comm.UserInfo{}, &orm.CreateTableOptions{IfNotExists: true})
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)InitTasksData() error {
	err := c.db.CreateTable(&comm.TaskInfo{}, &orm.CreateTableOptions{IfNotExists: true})
	if err != nil {
		return err
	}

	return nil
}

/* User's API */
func (c *DataCenterClient)PutUserCheckInfo(ci *comm.UserCheckInfo) error {
	err := c.db.Insert(ci)
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient) GetAllUsers() ([]comm.UserInfo, error) {
	var allUsers []comm.UserInfo

	err := c.db.Model(&allUsers).Select()
	if err != nil {
		return nil, err
	}

	return allUsers, nil
}

func (c *DataCenterClient)GetUserCheckInfo(ci *comm.UserCheckInfo) error {
	err := c.db.Select(ci)
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)DeleteUserCheckInfo(ci *comm.UserCheckInfo) error {
	err := c.db.Delete(ci)
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)PutUserInfo(ui *comm.UserInfo) error {
	err := c.db.Insert(ui)
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)GetUserInfo(ui *comm.UserInfo) error {
	err := c.db.Select(ui)
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)GetUserInfoByID(id uint32) (*comm.UserInfo, error) {
	userInfo := comm.UserInfo{
		ID: id,
	}

	err := c.db.Select(&userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (c *DataCenterClient)GetUserInfoByPhoneNumber(pn string) (*comm.UserInfo, error) {
	userInfo := comm.UserInfo{}
	err := c.db.Model(&userInfo).Where("phone_number = ?", pn).Limit(1).Select()
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (c *DataCenterClient)UpdateUserInfo(ui *comm.UserInfo) error {
	err := c.db.Update(ui)
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)UpdateUserStatusByPhoneNumber(phoneNumber string, curStatus int32) error {
	_, err := c.db.Model(&comm.UserInfo{}).Set("status = ?", curStatus).Where("phone_number = ?", phoneNumber).Update()
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)DeleteUserInfo(ui *comm.UserInfo) error {
	err := c.db.Delete(ui)
	if err != nil {
		return err
	}

	return nil
}
/* User's API */

/* Task's API */
func (c *DataCenterClient)PutTaskInfo(ti *comm.TaskInfo) error {
	err := c.db.Insert(ti)
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)UpdateTaskStatusByTaskID(id uint64, curStatus int32) error {
	_, err := c.db.Model(&comm.TaskInfo{}).Set("status = ?", curStatus).Where("id = ?", id).Update()
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)UpdateTaskResponsersByTaskID(id uint64, responsers []uint32) error {
	_, err := c.db.Model(&comm.TaskInfo{}).Set("responsers = ?", responsers).Where("id = ?", id).Update()
	if err != nil {
		return err
	}

	return nil
}

func (c *DataCenterClient)UpdateTaskInfo(ti *comm.TaskInfo) error {
	err := c.db.Update(ti)
	if err != nil {
		return err
	}

	return nil
}



/* Task's API */

