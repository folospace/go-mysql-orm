# golang mysql orm
- A painless orm 
- In progress
## struct of table 
```go
import (
    "database/sql"
    "github.com/folospace/go-mysql-orm/orm"
    _ "github.com/go-sql-driver/mysql"
)

//mysql db
var db, _ = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/mydb?parseTime=true&charset=utf8mb4&loc=Asia%2FShanghai")

//user table 
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
```
## select query 
```go
func main() {
    {
        var data User //select one user
        UserTable.Query().Limit(1).Select(&data)
    }
    {
        var data []User //select users
        UserTable.Query().Limit(5).Select(&data)
    }
    {
        var data int //select count
        UserTable.Query().SelectCount(&data)
    }
    {
        var data []int //select user.ids
        UserTable.Query().Limit(5).Select(&data, &UserTable.Id)
    }
    {
        var data map[int]User //select map[id]User
        UserTable.Query().Limit(5).Select(&data)
    }
    {
        var data map[int][]User //select map[id][]user
        UserTable.Query().Limit(5).Select(&data)
    }
    {
        var data map[int]string //select map[id]name
        UserTable.Query().Limit(5).Select(&data, &UserTable.Id, &UserTable.Name)
    }
}
```
## update query
```go
    {
        //update user set name="hello" where id=1
        UserTable.Query().Where(&UserTable.Id, 1).Update(&UserTable.Name, "hello")
    }

```

### join and where 
```go
       {
           //update user join order on user.id=order.user_id 
           //set order.order_amount=100
           //where user.id in (1,2)
           UserTable.Query().Join(OrderTable, func(query *orm.Query) {
               query.Where(&UserTable.Id, &OrderTable.UserId)
           }).
               Where(&UserTable.Id, orm.WhereIn, []int{1,2}). 
               Update(&OrderTable.OrderAmount, 100)
       }
```

## delete query
```go
	//query delete
	UserTable.Query().Where(&UserTable.Id, 1).Delete()
```

## insert query
```go
	//query insert
	_ = UserTable.Query().Insert(User{Name: "han"}).LastInsertId //insert one row and get id
	
	//insert rows and update column
	OrderTable.Query().InsertIgnore([]Order{{Id: 1, OrderAmount: 100}, {Id: 2, OrderAmount: 120}}, 
	[]interface{}{&OrderTable.Id, &OrderTable.OrderAmount},
        orm.UpdateColumn{ //update order amount if order id exist and amount is zero
            Column: &OrderTable.OrderAmount,
            Val:    orm.Raw("if(order_amount, order_amount, values(order_amount))"),
	})
```

## transaction
```go
    //transaction
    var data User
    _ = UserTable.Query().Transaction(func(db *orm.Query) error {
        db.FromTable(UserTable).Insert(User{Name: "john"}) //insert
        db.FromTable(UserTable).OrderByDesc(&UserTable.Id).Limit(1).Select(&data) //select
        return errors.New("I want rollback") //rollback
    }) 
```

## subquery
```go
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
```

## Relation (has many | belongs to)
```go
    //each user has many orders
    func (*User) LoadOrders(users []User) {
        var userIds []int
        for _, v := range users {
            userIds = append(userIds, v.Id)
        }
        
        var userOrders map[int][]Order
        OrderTable.Query().Where(&OrderTable.UserId, orm.WhereIn, userIds).
            Select(userOrders, &OrderTable.UserId, orm.AllCols)
        
        for k := range users {
            users[k].Orders = userOrders[users[k].Id]
        }
    }   
```

## migrate (create table)
- orm tag will override json tag
```go
    type User struct {
            Id int `json:"id"`
            Email string `json:"email" orm:"email,varchar(64),null,unique,index_email_and_score" comment:"user email"`
            Score int `json:"score" orm:"score,index,index_email_and_score" comment:"user score"`
            Name string `json:"name" default:"john" comment:"user name"`
            CreatedAt orm.JsonTime `json:"created_at"`
            UpdatedAt orm.JsonTime `json:"updated_at"`
    }
    createTableSql, err := UserTable.Query().Migrate()
    //create table IF NOT EXISTS `user` (`id` int not null auto_increment,
    //`email` varchar(64) null comment 'user email',
    //`score` int not null default '0' comment 'user score',
    //`name` varchar(255) not null default 'john' comment 'user name',
    //`created_at` timestamp not null default CURRENT_TIMESTAMP,
    //`updated_at` timestamp not null default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    //primary key (`id`),
    //unique key `email` (`email`),
    //key `score` (`score`),
    //key `index_email_and_score` (`email`,`score`)) 
```