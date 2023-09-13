Base on [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql).

## Get Started
```go
import (
    "database/sql"
    "github.com/folospace/go-mysql-orm/orm"
)

//connect mysql db
var db, _ = orm.OpenMysql("user:password@tcp(127.0.0.1:3306)/mydb?parseTime=true&charset=utf8mb4&loc=Asia%2FShanghai")

//user table
var UserTable = new(User)

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
```

## create table | create struct 

```go
func main() {
    orm.NewQuery(UserTable).CreateTable() //create db table, add new columns if table already exist.
    orm.NewQuery(UserTable).CreateStruct() //create struct fields in code
}        
```

## select query usage

```go
    //get first user (name='join') as struct
    user, query := orm.NewQuery(UserTable).Where(&UserTable.Name, "john").Get()
    
    //get users as slice of struct by primary ids
    users, query = orm.NewQuery(UserTable).Gets(1, 2, 3)
    
    //get users count(*)
    count, query := orm.NewQuery(UserTable).GetCount()
    
    //get user names map key by id
    var userNameKeyById map[int]string
    orm.NewQuery(UserTable).Select(&UserTable.Id, &UserTable.Name).GetTo(&userNameKeyById)
    
    //get users map key by name
    //useful when find has-many relations
    var usersMapkeyByName map[string][]User
    orm.NewQuery(UserTable).Select(&UserTable.Name, UserTable).GetTo(&usersMapkeyByName)
    
```

## update | delete | insert

```go
    //update user set name="hello" where id = 1
    orm.NewQuery(UserTable).WherePrimary(1).Update(&UserTable.Name, "hello")
    
    //query delete
    orm.NewQuery(UserTable).Delete(1, 2, 3)
    
    //query insert
    _ = orm.NewQuery(UserTable).Insert(User{Name: "han"}).LastInsertId //insert one row and get id

```

### join

```go
    //query join 
    orm.NewQuery(UserTable).Join(OrderTable, func (query *orm.Query[User]) *orm.Query[User] {
            return query.Where(&UserTable.Id, &OrderTable.UserId)
    }).Select(UserTable).Gets()
```

## transaction

```go
    //transaction
    _ = orm.NewQuery(UserTable).Transaction(func (query *orm.Query[User]) error {
        newId := query.Insert(User{Name: "john"}).LastInsertId //insert
        //newId := orm.NewQuery(UserTable).UseTx(query.Tx()).Insert(User{Name: "john"}).LastInsertId
        fmt.Println(newId)
        return errors.New("I want rollback") //rollback
    })
```

## subquery

```go
    //subquery
    subquery := orm.NewQuery(UserTable).Where(&UserQuery.T.Id, 1).SubQuery()
    
    //where in suquery
    orm.NewQuery(UserTable).Where(&UserQuery.T.Id, orm.WhereIn, subquery).Gets()
    
    //insert subquery
    orm.NewQuery(UserTable).InsertSubquery(subquery, nil)
    
    //join subquery
    orm.NewQuery(UserTable).Join(subquery, func (query *orm.Query[User]) *orm.Query[User] {
        return query.Where(&UserTable.Id, orm.Raw("sub.id"))
    }).Gets()
    
```