# golang mysql orm [![Go Report Card](https://goreportcard.com/badge/github.com/beatlabs/harvester)](https://goreportcard.com/report/github.com/folospace/go-mysql-orm) [![GoDoc](https://godoc.org/github.com/folospace/go-mysql-orm?status.svg)](https://godoc.org/github.com/folospace/go-mysql-orm) ![GitHub release](https://img.shields.io/github/v/release/folospace/go-mysql-orm.svg)

A golang orm package dedicated to simplify developing with mysql database.

Completed Features:
- support golang 1.18 generics
- try to avoid using raw strings.
- create mysql table from golang struct
- create golang struct from mysql table
- update query, insert query, delete query
- transaction
- select (update | insert) with subquery
- join (left join | right join) table
- select with window function (rank())
- select with common table expression (cte)

## Get Started

```go
import (
    "database/sql"
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
```

## migration (create table from struct | create struct from table)

```go
func main() {
    orm.CreateTableFromStruct(UserTable) //create db table, add new columns if table already exist.
    orm.CreateStructFromTable(UserTable) //create struct fields in code
}        
```

## select query usage

```go
    //get first user (name='join') as struct
    user, query := UserTable.Where("name", "john").Get()
    //replace raw string "name" to &T.Name (avoid misspelling and rename consequence)
    user, query := UserTable.Where(&UserTable.T.Name, "john").Get()
    
    //get user where primary id = 1
    user, query = UserTable.Get(1)
    
    //get users as slice of struct by primary ids
    users, query = UserTable.Gets(1, 2, 3)
    
    //get user first row as map[string]interface
    row, query := UserTable.GetRow()
    //it's useful when table struct not defined
    row, query := orm.NewQueryRaw(db, "user").GetRow()

    //get user rows as []map[string]interface
    rows, query := UserTable.Limit(5).GetRows()
    
    //get users count(*)
    count, query := UserTable.GetCount()
    
    //get users map key by id
    var usersKeyById map[int]User
    UserTable.GetTo(&usersKeyById)
    
    //get user names map key by id
    var userNameKeyById map[int]string
    UserTable.Select(&UserTable.T.Id, &UserTable.T.Name).GetTo(&userNameKeyById)
    
    //get users map key by name
    var usersMapkeyByName map[string][]User
    UserTable.Select(&UserTable.T.Name, UserTable.AllCols()).GetTo(&usersMapkeyByName)
    
    //simplify window function
    //select orders (where user_id=1) with rank by order_amount
    OrderTable.Where(&OrderTable.T.UserId, 1).
    Select(OrderTable.AllCols()).
    SelectRank(&OrderTable.T.OrderAmount, "order_amount_rank").GetRows()
    
    //simplify with recursive cte
    //select recursive to find children ...
    FileFolderTable.Where("id", 1).WithChildrenOnColumn("pid").GetRows()
    //select recursive to find parents ...
    FileFolderTable.Where("id", 9).WithParentsOnColumn("pid").GetRows()
```

## update | delete | insert

```go
    //update user set name="hello" where id=1
    UserTable.Where(&UserTable.T.Id, 1).Update(&UserTable.T.Name, "hello")
    
    //query delete
    UserTable.Delete(1, 2, 3)
    
    //query insert
    _ = UserTable.Insert(User{Name: "han"}).LastInsertId //insert one row and get id

```

### join and where

```go
    //query join 
    UserTable.Join(OrderTable.T, func (query orm.Query[User]) orm.Query[User] {
            return query.Where(&UserTable.T.Id, &OrderTable.T.UserId)
    }).Where(&OrderTable.T.OrderAmount, 100).Select(UserTable.AllCols()).Gets()
```

## transaction

```go
    //transaction
    _ = UserTable.Transaction(func (tx *sql.Tx) error {
        newId := UserTable.UseTx(tx).Insert(User{Name: "john"}).LastInsertId //insert
        fmt.Println(newId)
        return errors.New("I want rollback") //rollback
    })
```

## subquery

```go
    //subquery
    subquery := UserTable.Where(&UserTable.T.Id, 1).SubQuery()
    
    //where in suquery
    UserTable.Where(&UserTable.T.Id, orm.WhereIn, subquery).Gets()
    
    //insert subquery
    UserTable.InsertSubquery(subquery, nil)
    
    //join subquery
    UserTable.Join(subquery, func (query orm.Query[User]) orm.Query[User] {
        return query.Where(&UserTable.T.Id, orm.Raw("sub.id"))
    }).Gets()
```

## Relation (has many | belongs to)

```go
    users, _ := UserTable.Limit(5).Gets()
    var userIds []int
    for _, v := range users {
        userIds = append(userIds, v.Id)
    }
    
    //each user has many orders
    var userOrders map[int][]Order
    OrderTable.Where(&OrderTable.UserId, orm.WhereIn, userIds).
        Select(&OrderTable.UserId, OrderTable.AllCols()).
        GetTo(&userOrders)
    
    //set user has orders
    for k := range users {
        users[k].Orders = userOrders[users[k].Id]
    }
```

## about migration

- use json tag by default
- orm tag will override json tag
- default: column default value
- comment: column comment
- first column auto mark as primary key
- created_at, updated_at: predefined columns

```go
    type User struct {
        Id int `json:"id"`
        Email string `json:"email" orm:"email,varchar(64),null,unique,index_email_and_score" comment:"user email"`
        Score int `json:"score" orm:"score,index,index_email_and_score" comment:"user score"`
        Name string `json:"name" default:"john" comment:"user name"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
    }
//create table IF NOT EXISTS `user` (
//`id` int not null auto_increment,
//`email` varchar(64) null comment 'user email',
//`score` int not null default '0' comment 'user score',
//`name` varchar(255) not null default 'john' comment 'user name',
//`created_at` timestamp not null default CURRENT_TIMESTAMP,
//`updated_at` timestamp not null default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
//primary key (`id`),
//unique key `email` (`email`),
//key `score` (`score`),
//key `index_email_and_score` (`email`,`score`)
//) 
```

## Test in [go playground](https://go.dev/play/p/IjbPU1hCHMe)

## contribute

Contribute to this project by submitting a(an) PR | issue. Any contribution will be Appreciated.

## Thanks

Thanks to [goland support](https://jb.gg/OpenSourceSupport)