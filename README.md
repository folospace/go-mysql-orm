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

func (*User) Connections() []*sql.DB {
    return []*sql.DB{db}
}
func (*User) DatabaseName() string {
    return "mydb"
}
func (*User) TableName() string {
    return "user"
}
func (u *User) Query() *orm.Query[*User] {
    return orm.NewQuery(UserTable).WherePrimaryIfNotZero(u.Id)
}

func main() {
    //create db table, add new columns if table already exist.
    UserTable.Query().CreateTable()  
    
    //create struct from db table
    UserTable.Query().CreateStruct() 

    //insert one user
    UserTable.Query().Insert(&User{Name:"john"})
    
    //get user id = 1
    user, _ := UserTable.Query().Get(1)
    
    //update user name
    user.Query().Update(&UserTable.Name, "john 2")
}
```

## select

```go
    //get users where name='join'
    users, _ := UserTable.Query().Where(&UserTable.Name, "john").Gets()
    
    //get users count(*)
    count, _ := UserTable.Query().GetCount()
    
    //get user names map key by id
    var userNameKeyById map[int]string
    UserTable.Query().Select(&UserTable.Id, &UserTable.Name).GetTo(&userNameKeyById)
    
    //get users map key by name
    //useful when find has-many relations
    var usersMapkeyByName map[string][]*User
    UserTable.Query().Select(&UserTable.Name, UserTable).GetTo(&usersMapkeyByName)
    
```

## update | delete | insert

```go
    //update user set name="john 2" where id = 1
    UserTable.Query().WherePrimary(1).Update(&UserTable.Name, "john 2")
    
    //query delete
    UserTable.Query().Delete(1, 2, 3)
    
    //query insert
    _ = UserTable.Query().Insert(&User{Name: "han"}).LastInsertId //insert one row and get id

```

### join

```go
    //query join 
    UserTable.Query().Join(OrderTable, func (query *orm.Query[*User]) *orm.Query[*User] {
            return query.Where(&UserTable.Id, &OrderTable.UserId)
    }).Select(UserTable).Gets()
```

## transaction

```go
    //transaction
    _ = UserTable.Query().Transaction(func (query *orm.Query[*User]) error {
        newId := query.Insert(&User{Name: "john"}).LastInsertId //insert
        //newId := orm.NewQuery(UserTable).UseTx(query.Tx()).Insert(User{Name: "john"}).LastInsertId
        fmt.Println(newId)
        return errors.New("I want rollback") //rollback
    })
```

## subquery

```go
    //subquery
    subquery := UserTable.Query().Where(&UserTable.Id, 1).SubQuery()
    
    //where in suquery
    UserTable.Query().Where(&UserTable.Id, orm.WhereIn, subquery).Gets()
    
    //insert subquery
    UserTable.Query().InsertSubquery(subquery)
    
    //join subquery
    UserTable.Query().Join(subquery, func (query *orm.Query[*User]) *orm.Query[*User] {
        return query.Where(&UserTable.Id, orm.Raw("sub.id"))
    }).Gets()
    
```