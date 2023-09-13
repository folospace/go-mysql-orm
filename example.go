package main

import (
    "database/sql"
    "errors"
    "fmt"
    "github.com/folospace/go-mysql-orm/orm"
)

//connect mysql db
var db, _ = orm.OpenMysql("user:password@tcp(127.0.0.1:3306)/mydb?parseTime=true&charset=utf8mb4&loc=Asia%2FShanghai")

//query user
var UserQuery = orm.NewQuery(User{})

type User struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

func (User) Connection() []*sql.DB {
    return []*sql.DB{db}
}
func (User) DatabaseName() string {
    return "mydb"
}
func (User) TableName() string {
    return "user"
}

//user table model
var OrderQuery = orm.NewQuery(Order{})

type Order struct {
    Id          int `json:"id"`
    UserId      int `json:"user_id"`
    OrderAmount int `json:"order_amount"`
}

func (o Order) Connection() []*sql.DB {
    return []*sql.DB{db}
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
        _, _ = orm.CreateTableFromStruct(UserQuery)
        //create go struct from db table
        _ = orm.CreateStructFromTable(UserQuery)
    }

    //query select
    {
        //get first user (name='join') as struct
        user, query := UserQuery.Where(&UserQuery.T.Name, "john").Get()
        fmt.Println(user, query.Sql(), query.Error())

        //get users by primary ids
        users, query := UserQuery.Gets(1, 2, 3)
        fmt.Println(users, query.Sql(), query.Error())

        //get user rows as []map[string]interface
        rows, query := UserQuery.Limit(5).GetRows()
        fmt.Println(rows, query.Sql(), query.Error())

        //get users count(*)
        count, query := UserQuery.GetCount()
        fmt.Println(count, query.Sql(), query.Error())

        //get user names map key by id
        var userNameKeyById map[int]string
        UserQuery.Select(&UserQuery.T.Id, &UserQuery.T.Name).GetTo(&userNameKeyById)

        //get users map key by name
        var usersMapkeyByName map[string][]User
        UserQuery.Select(&UserQuery.T.Name, UserQuery.AllCols()).GetTo(&usersMapkeyByName)

        //select rank by column
        OrderQuery.Where(&OrderQuery.T.UserId, 1).
            Select(OrderQuery.AllCols()).
            SelectRank(&OrderQuery.T.OrderAmount, "order_amount_rank").GetRows()
    }

    //query update and delete and insert
    {
        //update user set name="hello" where id=1
        UserQuery.WherePrimary(1).Update(&UserQuery.T.Name, "hello")

        //query delete
        UserQuery.Delete(1, 2, 3)

        //query insert
        _ = UserQuery.Insert(User{Name: "han"}).LastInsertId //insert one row and get id

        //insert batch on duplicate key update name=values(name)
        _ = UserQuery.InsertsIgnore([]User{{Id: 1, Name: "han"}, {Id: 2, Name: "jen"}},
            []orm.UpdateColumn{{Column: &UserQuery.T.Name, Val: &UserQuery.T.Name}})
    }

    //query join
    {
        UserQuery.Join(OrderQuery.T, func(query orm.Query[User]) orm.Query[User] {
            return query.Where(&UserQuery.T.Id, &OrderQuery.T.UserId)
        }).Where(&OrderQuery.T.OrderAmount, 100).
            Select(UserQuery.AllCols()).Gets()
    }
    {
        //transaction
        _ = UserQuery.Transaction(func(tx *sql.Tx) error {
            newId := UserQuery.UseTx(tx).Insert(User{Name: "john"}).LastInsertId //insert
            fmt.Println(newId)
            return errors.New("I want rollback") //rollback
        })
    }

    {
        //subquery
        subquery := UserQuery.Where(&UserQuery.T.Id, 1).SubQuery()

        //where in suquery
        UserQuery.Where(&UserQuery.T.Id, orm.WhereIn, subquery).Gets()

        //insert subquery
        UserQuery.InsertSubquery(subquery, nil)

        //join subquery
        UserQuery.Join(subquery, func(query orm.Query[User]) orm.Query[User] {
            return query.Where(&UserQuery.T.Id, orm.Raw("sub.id"))
        }).Gets()
    }
}
