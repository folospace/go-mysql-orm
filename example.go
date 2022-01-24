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
var UserTable = new(User)

type User struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

func (m *User) Query() *orm.Query {
    return new(orm.Query).UseDB(db).FromTable(m)
}
func (*User) TableName() string {
    return "user"
}
func (*User) DatabaseName() string {
    return "mydb"
}

//user table model
var OrderTable = new(Order)

type Order struct {
    Id          int `json:"id"`
    UserId      int `json:"user_id"`
    OrderAmount int `json:"order_amount"`
}

func (m *Order) Query() *orm.Query {
    return new(orm.Query).UseDB(db).FromTable(m)
}
func (*Order) TableName() string {
    return "order"
}
func (*Order) DatabaseName() string {
    return "mydb"
}

func main() {
    //query select
    {
        var data User //select one user
        UserTable.Query().Limit(1).Select(&data)
        fmt.Println(data)
    }
    {
        var data []User //select users
        UserTable.Query().Limit(5).Select(&data)
        fmt.Println(data)
    }
    {
        var data int //select count
        UserTable.Query().SelectCount(&data)
        fmt.Println(data)
    }
    {
        var data []int //select user.ids
        UserTable.Query().Limit(5).Select(&data, &UserTable.Id)
        fmt.Println(data)
    }
    {
        var data map[int]User //select map[id]user
        UserTable.Query().Limit(5).Select(&data)
        fmt.Println(data)
    }
    {
        var data map[int]string //select map[id]name
        UserTable.Query().Limit(5).Select(&data, &UserTable.Id, &UserTable.Name)
        fmt.Println(data)
    }

    //query where
    {
        //update user set name="hello" where id=1
        UserTable.Query().Where(&UserTable.Id, 1).Update(&UserTable.Name, "hello")
    }

    //query join and where
    {
        UserTable.Query().Join(OrderTable, func(query *orm.Query) {
            query.Where(&UserTable.Id, &OrderTable.UserId)
        }).
            Where(&UserTable.Id, orm.WhereIn, []int{1, 2}).
            Update(&OrderTable.OrderAmount, 100)
    }

    {
        //query delete
        UserTable.Query().Where(&UserTable.Id, 1).Delete()
    }

    {
        //query insert
        _ = UserTable.Query().Insert(User{Name: "han"}).LastInsertId //insert one row and get id
        //insert rows and update column
        OrderTable.Query().InsertIgnore([]Order{{Id: 1, OrderAmount: 100}, {Id: 2, OrderAmount: 120}},
            []interface{}{&OrderTable.Id, &OrderTable.OrderAmount},
            orm.UpdateColumn{ //update order amount if order id exist and amount is zero
                Column: &OrderTable.OrderAmount,
                Val:    orm.Raw("if(order_amount, order_amount, values(order_amount))"),
            })
    }

    {
        //transaction
        var data User
        _ = UserTable.Query().Transaction(func(db *orm.Query) error {
            db.FromTable(UserTable).Insert(User{Name: "john"})                        //insert
            db.FromTable(UserTable).OrderByDesc(&UserTable.Id).Limit(1).Select(&data) //select
            return errors.New("I want rollback")                                      //rollback
        })
    }

    //subquery
    {
        //join subquery
        var data []Order
        //select * from order join (select id from user limit 5) temp on order.user_id=temp.id
        OrderTable.Query().Join(UserTable.Query().Limit(5).PreSelectAsTemp(&UserTable.Id), func(join *orm.Query) {
            join.Where(&OrderTable.UserId, orm.Raw("temp.id"))
        }).Select(&data)
    }
    {
        subquery := UserTable.Query().Limit(5).PreSelectAsTemp(&UserTable.Id)

        var data []User
        //select * from (subquery)
        subquery.Query().Select(&data)
        UserTable.Query().FromTable(subquery).Select(&data)

        //select * from user where id in (subquery)
        UserTable.Query().Where(&UserTable.Id, orm.WhereIn, subquery).Select(&data)
    }
}
