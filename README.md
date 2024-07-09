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
    Id        int       `json:"id"`
    Email     string    `json:"email" orm:"email,unique"`
    Name      string    `json:"name" default:"jack"`
    Avatar    string    `json:"avatar" comment:"head image"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
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
}
```

## select

```go
    //select * from user where id = 1 //to struct
    user, _ := UserTable.Query().Get(1)
    fmt.Println(user) //User{Id:1}

    //select * from user where name='john' //to struct slice
    users, _ := UserTable.Query().Where(&UserTable.Name, "john").Gets()
    fmt.Println(users) //User{Id:1}, User{Id:2}, ...
    
    //select email from user //to slice
    emails, _ := UserTable.Query().Select(&UserTable.Email).Limit(10).GetSliceString()
    fmt.Println(emails) //a**@gmail.com, b**@gmail.com, ...
    
    //select user info to slice, group by id
    var userInfoMap map[int][]string
    UserTable.Query().Select(&UserTable.Id, &UserTable.Email, &UserTable.Name).Limit(10).GetTo(&userInfoMap)
    fmt.Println(userInfoMap) //{1:[a**@gmail.com, a**], 2:[b**@gmail.com, b**], ...}
    
    //select user id to slice, group by name
    var sameNameUsers map[string][]int
    UserTable.Query().Select(&UserTable.Name, &UserTable.Id).Limit(10).GetTo(&sameNameUsers)
    fmt.Println(sameNameUsers) //{a**:[1,3], b**:[2,4], ...}
    
```

## update | delete | insert

```go
    //update user set name="john 2" where id = 1
    UserTable.Query().WherePrimary(1).Update(&UserTable.Name, "john 2")
    
    //delete
    UserTable.Query().Delete(1, 2, 3)
    
    //insert
    _ = UserTable.Query().Insert(&User{Name: "han"})   
    
    //update users with different names
    _ = UserTable.Query().OnConflictUpdate(&UserTable.Name, &UserTable.Name).
    Insert(&User{Id: 1, Name: "han"}, &User{Id: 2, Name: "join"})
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
    subquery := UserTable.Query().WherePrimary(1).Select(&UserTable.Id).SubQuery()
    
    //where in suquery
    UserTable.Query().Where(&UserTable.Id, orm.WhereIn, subquery).Gets()
    
    //insert subquery
    UserTable.Query().InsertSubquery(subquery)
    
    //join subquery
    UserTable.Query().Join(subquery, func (query *orm.Query[*User]) *orm.Query[*User] {
        return query.Where(&UserTable.Id, orm.Raw("sub.id"))
    }).Gets()
    
```