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
        //get first user (name='join') as struct
        user, query := UserTable.Where(&UserTable.T.Name, "john").Get()
        fmt.Println(user, query.Sql(), query.Error())

        //get users by primary ids
        users, query := UserTable.Gets(1, 2, 3)
        fmt.Println(users, query.Sql(), query.Error())

        //get user rows as []map[string]interface
        rows, query := UserTable.Limit(5).GetRows()
        fmt.Println(rows, query.Sql(), query.Error())

        //get users count(*)
        count, query := UserTable.GetCount()
        fmt.Println(count, query.Sql(), query.Error())

        //get user names map key by id
        var userNameKeyById map[int]string
        UserTable.Select(&UserTable.T.Id, &UserTable.T.Name).GetTo(&userNameKeyById)

        //get users map key by name
        var usersMapkeyByName map[string][]User
        UserTable.Select(&UserTable.T.Name, UserTable.AllCols()).GetTo(&usersMapkeyByName)

        //select rank by column
        OrderTable.Where(&OrderTable.T.UserId, 1).
            Select(OrderTable.AllCols()).
            SelectRank(&OrderTable.T.OrderAmount, "order_amount_rank").GetRows()
    }

    //query update and delete and insert
    {
        //update user set name="hello" where id=1
        UserTable.WherePrimary(1, 2, 3).Update(&UserTable.T.Name, "hello")

        //query delete
        UserTable.Delete(1, 2, 3)

        //query insert
        _ = UserTable.Insert(User{Name: "han"}).LastInsertId //insert one row and get id

        //insert batch on duplicate key update name=values(name)
        _ = UserTable.InsertsIgnore([]User{{Id: 1, Name: "han"}, {Id: 2, Name: "jen"}},
            []orm.UpdateColumn{{Column: &UserTable.T.Name, Val: &UserTable.T.Name}})
    }

    //query join
    {
        UserTable.Join(OrderTable.T, func(query orm.Query[User]) orm.Query[User] {
            return query.Where(&UserTable.T.Id, &OrderTable.T.UserId)
        }).Where(&OrderTable.T.OrderAmount, 100).
            Select(UserTable.AllCols()).Gets()
    }
    {
        //transaction
        _ = UserTable.Transaction(func(tx *sql.Tx) error {
            newId := UserTable.UseTx(tx).Insert(User{Name: "john"}).LastInsertId //insert
            fmt.Println(newId)
            return errors.New("I want rollback") //rollback
        })
    }

    {
        //subquery
        subquery := UserTable.Where(&UserTable.T.Id, 1).SubQuery()

        //where in suquery
        UserTable.Where(&UserTable.T.Id, orm.WhereIn, subquery).Gets()

        //insert subquery
        UserTable.InsertSubquery(subquery, nil)

        //join subquery
        UserTable.Join(subquery, func(query orm.Query[User]) orm.Query[User] {
            return query.Where(&UserTable.T.Id, orm.Raw("sub.id"))
        }).Gets()
    }
}
