package main

import (
    "database/sql"
    "errors"
    "fmt"
    "github.com/folospace/go-mysql-orm/orm"
    _ "github.com/go-sql-driver/mysql"
)

//connect mysql db
var db, _ = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/mydb?parseTime=true&charset=utf8mb4&loc=Asia%2FShanghai")

//user table model
var UserTable = orm.NewQuery(User{}, db)

type User struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

func (User) TableName() string {
    return "user"
}
func (User) DatabaseName() string {
    return "mydb"
}

//user table model
var OrderTable = orm.NewQuery(Order{}, db)

type Order struct {
    Id          int `json:"id"`
    UserId      int `json:"user_id"`
    OrderAmount int `json:"order_amount"`
}

func (Order) TableName() string {
    return "order"
}
func (Order) DatabaseName() string {
    return "mydb"
}

func main() {
    {
        //create db table from go struct
        _, _ = orm.CreateTableFromStruct(UserTable)
        //create go struct from db table
        _ = orm.CreateStructFromTable(UserTable)
    }

    //query select
    {
        //get first user as struct
        user, query := UserTable.Get()
        fmt.Println(user, query.Sql(), query.Error())

        //get users as struct slice
        users, query := UserTable.Limit(5).Gets()
        fmt.Println(users, query.Sql(), query.Error())

        //get user first row as map[string]interface
        row, query := UserTable.GetRow()
        fmt.Println(row, query.Sql(), query.Error())

        //get user rows as []map[string]interface
        rows, query := UserTable.Limit(5).GetRows()
        fmt.Println(rows, query.Sql(), query.Error())

        //get users count(*)
        count, query := UserTable.GetCount()
        fmt.Println(count, query.Sql(), query.Error())

        //get users map key by id
        var usersKeyById map[int]User
        UserTable.GetTo(&usersKeyById)

        //get user names map key by id
        var userNameKeyById map[int]string
        UserTable.Select(&UserTable.T.Id, &UserTable.T.Name).GetTo(&userNameKeyById)

        //get users map key by name
        var usersMapkeyByName map[string][]User
        UserTable.Select(&UserTable.T.Name, UserTable.AllCols()).GetTo(&usersMapkeyByName)
    }

    //query update and delete and insert
    {
        //update user set name="hello" where id=1
        UserTable.Where(&UserTable.T.Id, 1).Update(&UserTable.T.Name, "hello")

        //query delete
        UserTable.Where(&UserTable.T.Id, 1).Delete()

        //query insert
        _ = UserTable.Insert(User{Name: "han"}).LastInsertId //insert one row and get id
    }

    //query join
    {
        UserTable.Join(OrderTable.T, func(query orm.Query[User]) orm.Query[User] {
            return query.Where(&UserTable.T.Id, &OrderTable.T.UserId)
        }).Where(&OrderTable.T.OrderAmount, 100).
            Select(UserTable.AllCols()).Gets()

        {
            //transaction
            _ = UserTable.Transaction(func(tx *sql.Tx) error {
                newId := UserTable.UseTx(tx).Insert(User{Name: "john"}).LastInsertId //insert
                fmt.Println(newId)
                return errors.New("I want rollback") //rollback
            })
        }

        //subquery
        subquery := UserTable.Query().Limit(5).SelectSub(&UserTable.Id)
        {
            //join subquery
            var data []Order

            //select * from order join (select id from user limit 5) sub on order.user_id=sub.id
            OrderTable.Query().Join(subquery, func(join *orm.Query) {
                join.Where(&OrderTable.UserId, orm.Raw("sub.id"))
            }).Select(&data)
        }
        {
            var data []User
            //select * from (subquery)
            subquery.Query().Select(&data)
            UserTable.Query().FromTable(subquery).Select(&data)

            //select * from user where id in (subquery)
            UserTable.Query().Where(&UserTable.Id, orm.WhereIn, subquery).Select(&data)

            //insert ingore into user (id) select id from user limit 5 on duplicate key update name="change selected users' name"
            UserTable.Query().InsertIgnore(subquery, []interface{}{&UserTable.Id}, orm.UpdateColumn{Column: &UserTable.Name, Val: "change selected users' name"})
        }
    }
