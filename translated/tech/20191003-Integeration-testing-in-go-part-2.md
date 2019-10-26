# Go 语言中的集成测试：第二部分 - 设计和编写测试

## 序幕

## 简介

## 管理种子数据

<!-- todo 确定以下 -->
## 播种数据库

### 清单 1

```golang
func SeedLists(dbc *sqlx.DB) ([]list.List, error) {
    now := time.Now().Truncate(time.Microsecond)

    lists := []list.List{
        {
            Name:     "Grocery",
            Created:  now,
            Modified: now,
        },
        {
            Name:     "To-do",
            Created:  now,
            Modified: now,
        },
        {
            Name:     "Employees",
            Created:  now,
            Modified: now,
        },
    }

    for i := range lists {
        stmt, err := dbc.Prepare("INSERT INTO list (name, created, modified) VALUES ($1, $2, $3) RETURNING list_id;")
        if err != nil {
            return nil, errors.Wrap(err, "prepare list insertion")
        }

        row := stmt.QueryRow(lists[i].Name, lists[i].Created, lists[i].Modified)

        if err = row.Scan(&lists[i].ID); err != nil {
            if err := stmt.Close(); err != nil {
                return nil, errors.Wrap(err, "close psql statement")
            }

            return nil, errors.Wrap(err, "capture list id")
        }

        if err := stmt.Close(); err != nil {
            return nil, errors.Wrap(err, "close psql statement")
        }
    }

    return lists, nil
}
```

### 清单 2

```golang
func SeedItems(dbc *sqlx.DB, lists []list.List) ([]item.Item, error) {
    now := time.Now().Truncate(time.Microsecond)

    items := []item.Item{
        {
            ListID:   lists[0].ID, // Grocery
            Name:     "Chocolate Milk",
            Quantity: 1,
            Created:  now,
            Modified: now,
        },
        {
            ListID:   lists[0].ID, // Grocery
            Name:     "Mac and Cheese",
            Quantity: 2,
            Created:  now,
            Modified: now,
        },
        {
            ListID:   lists[1].ID, // To-do
            Name:     "Write Integration Tests",
            Quantity: 1,
            Created:  now,
            Modified: now,
        },
    }

    for i := range items {
        stmt, err := dbc.Prepare("INSERT INTO item (list_id, name, quantity, created, modified) VALUES ($1, $2, $3, $4, $5) RETURNING item_id;")
        if err != nil {
            return nil, errors.Wrap(err, "prepare item insertion")
        }

        row := stmt.QueryRow(items[i].ListID, items[i].Name, items[i].Quantity, items[i].Created, items[i].Modified)

        if err = row.Scan(&items[i].ID); err != nil {
            if err := stmt.Close(); err != nil {
                return nil, errors.Wrap(err, "close psql statement")
            }

            return nil, errors.Wrap(err, "capture list id")
        }

        if err := stmt.Close(); err != nil {
            return nil, errors.Wrap(err, "close psql statement")
        }
    }

    return items, nil
}
```
### 清单 3

```golang
func Truncate(dbc *sqlx.DB) error {
    stmt := "TRUNCATE TABLE list, item;"

    if _, err := dbc.Exec(stmt); err != nil {
        return errors.Wrap(err, "truncate test database tables")
    }

    return nil
}
```

## 使用 testing.M 创建 TestMain

### 清单 4

```golang
func TestMain(m *testing.M) {
    os.Exit(testMain(m))
}
```

### 清单 5

```golang
func testMain(m *testing.M) int {
    dbc, err := testdb.Open()
    if err != nil {
        log.WithError(err).Info("create test database connection")
        return 1
    }
    defer dbc.Close()

    a = handlers.NewApplication(dbc)

    return m.Run()
}
```

## 编写 Web 服务的集成测试

### 清单 6

```golang
func Test_getItems(t *testing.T) {
    defer func() {
        if err := testdb.Truncate(a.DB); err != nil {
            t.Errorf("error truncating test database tables: %v", err)
        }
    }()

    expectedLists, err := testdb.SeedLists(a.DB)
    if err != nil {
        t.Fatalf("error seeding lists: %v", err)
    }

    expectedItems, err := testdb.SeedItems(a.DB, expectedLists)
    if err != nil {
        t.Fatalf("error seeding items: %v", err)
    }
}
```

### 清单 7

```golang
// Application is the struct that contains the server handler as well as
// any references to services that the application needs.
type Application struct {
    DB      *sqlx.DB
    handler http.Handler
}

// ServeHTTP implements the http.Handler interface for the Application type.
func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    a.handler.ServeHTTP(w, r)
}

```

### 清单 8

```golang
req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/list/%d/item", test.ListID), nil)
if err != nil {
   t.Errorf("error creating request: %v", err)
}

w := httptest.NewRecorder()
a.ServeHTTP(w, req)
```

### 清单 9

```golang
if want, got := http.StatusOK, w.Code; want != got {
    t.Errorf("expected status code: %v, got status code: %v", want, got)
}
```

### 清单 10

```golang
var items []item.Item
resp := web.Response{
    Results: items,
}

if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
    t.Errorf("error decoding response body: %v", err)
}

if d := cmp.Diff(expectedItems, items); d != "" {
    t.Errorf("unexpected difference in response body:\n%v", d)
}
```

### 清单 11

```golang
// Add takes an indefinite amount of operands and adds them together, returning
// the sum of the operation.
func Add(operands ...int) int {
    var sum int

    for _, operand := range operands {
        sum += operand
    }

    return sum
}
```
### 清单 12

```golang
// TestAdd tests the Add function.
func TestAdd(t *testing.T) {
    tt := []struct {
        Name     string
        Operands []int
        Sum      int
    }{
        {
            Name:     "NoOperands",
            Operands: []int{},
            Sum:      0,
        },
        {
            Name:     "OneOperand",
            Operands: []int{10},
            Sum:      10,
        },
        {
            Name:     "TwoOperands",
            Operands: []int{10, 5},
            Sum:      15,
        },
        {
            Name:     "ThreeOperands",
            Operands: []int{10, 5, 4},
            Sum:      19,
        },
    }

    for _, test := range tt {
        fn := func(t *testing.T) {
            if e, a := test.Sum, Add(test.Operands...); e != a {
                t.Errorf("expected sum %d, got sum %d", e, a)
            }
        }

        t.Run(test.Name, fn)
    }
}
```

### 清单 13

```golang
// GenerateTempFile generates a temp file and returns the reference to
// the underlying os.File and an error.
func GenerateTempFile() (*os.File, error) {
    f, err := ioutil.TempFile("", "")
    if err != nil {
        return nil, err
    }

    return f, nil
}
```

### 清单 14

```golang
// GenerateTempFile generates a temp file and returns the reference to
// the underlying os.File.
func GenerateTempFile(t *testing.T) *os.File {
    t.Helper()

    f, err := ioutil.TempFile("", "")
    if err != nil {
        t.Fatalf("unable to generate temp file: %v", err)
    }

    return f
}
```

## 结论

---

via: <https://www.ardanlabs.com/blog/2019/10/integration-testing-in-go-set-up-and-writing-tests.html>

作者：[George Shaw](https://github.com/george-e-shaw-iv/)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[校对者 ID](https://github.com/校对者 ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出