# 如何使用golang处理mysql的binlog
大家好，我是Artem，一名golang开发。我们的团队花费了大量时间训练MySQL binlog。这里整合一些简单用法，不会放过任何隐藏的陷阱。示例代码将在最后显示。

每次从数据库查询的返回结果中拉取用户信息时，主项目中会有高负载模块。此时使用缓存是一个不错的建议，但是什么时候重置缓存呢？这需要由数据来决定更新时间。  
MySQL的主从复制是一个很棒的设计。而我们的守护进程可以视为一个通过binlog获取数据的slave，binlog设置成row格式。这样就能使用所有的数据库命令，但事务下的命令只有在提交后才会记录。在达到内存的使用限制后（默认为1GB），会开启另一个文件，每个新文件的名称后都会有一个增量。  
更多信息查看<https://mariadb.com/kb/en/library/binary-log/>或者<https://dev.mysql.com/doc/refman/8.0/en/binary-log.html>  
本文将分为以下两部分：  
>1、如何处理binlog中的新数据  
>2、如何设置和扩展
  
### part 1.快速运行
我们可以使用这个库<https://github.com/siddontang/go-mysql/>来处理binlog。  
连接到一个新的channel（chanal是一个库的标签）。我们将使用binlog中的row格式<https://mariadb.com/kb/en/library/binary-log-formats/>。  

```
func binLogListener() {
	c, err := getDefaultCanal()
	if err == nil {
		coords, err := c.GetMasterPos()
		if err == nil {
			c.SetEventHandler(&binlogHandler{})
			c.RunFrom(coords)
		}
	}
}
func getDefaultCanal() (*canal.Canal, error) {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d", "127.0.0.1", 3306)
	cfg.User = "root"
	cfg.Password = "root"
	cfg.Flavor = "mysql"
	cfg.Dump.ExecutionPath = ""

	return canal.NewCanal(cfg)
}
```
现在来进行封装

```
type binlogHandler struct {
	canal.DummyEventHandler // Dummy handler from external lib
	BinlogParser // Our custom helper
}
func (h *binlogHandler) OnRow(e *canal.RowsEvent) error {return nil}
func (h *binlogHandler) String() string {return "binlogHandler"}
```

[BinlogParser](https://github.com/JackShadow/go-binlog-example/blob/master/src/parser.go)

然后我们可以在OnRow()方法中添加一些代码逻辑，让他更好用  

```
func (h *binlogHandler) OnRow(e *canal.RowsEvent) error {
	var n int //starting value
	var k int // step
	switch e.Action {
	case canal.DeleteAction:
		return nil // not covered in example
	case canal.UpdateAction:
		n = 1
		k = 2
	case canal.InsertAction:
		n = 0
		k = 1
	}
	for i := n; i < len(e.Rows); i += k {
		key := e.Table.Schema + "." + e.Table.Name
		switch key {
		case User{}.SchemaName() + "." + User{}.TableName():
			/*
			 Real data parsing
			*/
		}
	}
	return nil
}
```
这个包装器的主要逻辑是解析接收到的数据。我们可以通过更新的两个条件获取数据（第一条包含初始数据，第二条则是更新数据），同时也支持多行插入和多行更新。在这种情况下，执行UPDATE操作时，每次都要使用第二个条件。而执行INSERT时，需要操作每一行，为此我们需要使用n和k变量。  
从binlog中获取一个模版，逐行加载数据，每个column都标明注释：

```
type User struct {
	Id      int       `gorm:"column:id"`
	Name    string    `gorm:"column:name"`
	Status  string    `gorm:"column:status"`
	Created time.Time `gorm:"column:created"`
}
func (User) TableName() string {
	return "User"
}
func (User) SchemaName() string {
	return "Test"
}
```

MySQL的表结构体

```
CREATE TABLE Test.User(
id INT AUTO_INCREMENT PRIMARY KEY,
name VARCHAR(40) NULL ,
status ENUM("active","deleted") DEFAULT "active",
created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP
)
ENGINE =InnoDB;
```
现在使用代码的方式实现

```
user := User{}
h.GetBinLogData(&user, e, i)
```
最后，我们可以通过新增加的用户变量获取数据。打印出来让它看起来更美观。

```
if e.Action == canal.UpdateAction {
  oldUser := User{}
  h.GetBinLogData(&oldUser, e, i-1)
  fmt.Printf("User %d is updated from name %s to name %s\n", user.Id, oldUser.Name, user.Name, )
 } else {
  fmt.Printf("User %d is created with name %s\n", user.Id, user.Name, )
}
```
太好了，代码即将实现， “Hello, binlog world”:

```
func main() {
	go binLogListener()
	// placeholder for your handsome code
	time.Sleep(2 * time.Minute)
	fmt.Print("Thx for watching")
}
```
新增和更新用户：

```
INSERT INTO Test.User (`id`,`name`) VALUE (1,"Jack");
UPDATE Test.User SET name="Jonh" WHERE id=1;
```
结果展示：

```
User 1 is created with name Jack
User 1 name changed from Jack to Jonh
```
这段代码通过binlog来解析新增的row，并通过数据表获取我们需要的数据，在结构体中解析数据并输出结果。我没有介绍所有的数据解析器（BinlogParser），这其中还隐藏了一些hydration逻辑模型。  
### part 2.正如cobb所说，我们需要更加深入了解
解析器的隐藏部分是基于反射，可以使用下面这种方式来进行hydration模型。

```
h.GetBinLogData(&user, e, i)
```

使用一些简单的数据类型来处理

```
bool
int
float64
string
time.Time
```
也可以通过json来解析结构体。  
如果你需要更多的数据类型, 或者你只是想知道binlog是如何进行解析工作的, 最好的办法是自己扩展解析类型。  
下面是一个`int`类型的实例:

```
type User struct {
 Id      int       `gorm:"column:id"`
}
```
我们可以通过反射来获取类型名称。parseTagSetting方法可以使注释更便于使用：

```
element := User{} //In common cases we have interface, but here we will start with model
v := reflect.ValueOf(element)
s := reflect.Indirect(v)
t := s.Type()
num := t.NumField()
parsedTag := parseTagSetting(t.Field(k).Tag)
if columnName, ok = parsedTag["COLUMN"]; !ok || columnName == "COLUMN" {
	continue
 }
for k := 0; k < num; k++ {
	name := s.Field(k).Type().Name()
	switch name {
   		case "int":
    // here we deal with an incoming row
	}
}
```
获取了类型名称的同时也可以通过反射来设置它的值

```
func (v Value) SetInt(x int64) {//...

```
解析注释帮助器（从gorm库获取）

```
func parseTagSetting(tags reflect.StructTag) map[string]string {
 setting := map[string]string{}
 for _, str := range []string{tags.Get("sql"), tags.Get("gorm")} {
  tags := strings.Split(str, ";")
  for _, value := range tags {
   v := strings.Split(value, ":")
   k := strings.TrimSpace(strings.ToUpper(v[0]))
   if len(v) >= 2 {
   setting[k] = strings.Join(v[1:], ":")
    } else {
    setting[k] = k
    }
   }
  }
 return setting
}
```
解析器中有int64类型，我们可以创建一个将int64转换为row类型的方法：

```
func (m *BinlogParser) intHelper(e *canal.RowsEvent, n int, columnName string) int64 {
	columnId := m.getBinlogIdByName(e, columnName)
	if e.Table.Columns[columnId].Type != schema.TYPE_NUMBER {
		return 0
	}
	switch e.Rows[n][columnId].(type) {
	case int8:
		return int64(e.Rows[n][columnId].(int8))
	case int32:
		return int64(e.Rows[n][columnId].(int32))
	case int64:
		return e.Rows[n][columnId].(int64)
	case int:
		return int64(e.Rows[n][columnId].(int))
	case uint8:
		return int64(e.Rows[n][columnId].(uint8))
	case uint16:
		return int64(e.Rows[n][columnId].(uint16))
	case uint32:
		return int64(e.Rows[n][columnId].(uint32))
	case uint64:
		return int64(e.Rows[n][columnId].(uint64))
	case uint:
		return int64(e.Rows[n][columnId].(uint))
	}
	return 0
}
```
除了getBinlogIdByName()方法，所有东西看起来都是合理的。  
需要使用trivial帮助器来处理column名而不是它的id，这样可以：  
> 使用gorm注释来处理字段名；  
> 在开头和中间添加字段名时不需要额外修改；  
> 使用字段名处理比column3更方便。
	  
最后，我们加入以下处理：

```
s.Field(k).SetInt(m.intHelper(e, n, columnName))
```
### 还有两个例子
ENUM:我们将获取的值作为索引——所以“active”状态会被设置为0。同样的，我们也需要用enum字符串表示，而不是id，这些可以从字段介绍中获取。重要提示，值中的1描述的是0值索引字段，数组的值是从0开始的。  
Enum的解析如下：

```
func (m *BinlogParser) stringHelper(e *canal.RowsEvent, n int, columnName string) string {
	columnId := m.getBinlogIdByName(e, columnName)
	if e.Table.Columns[columnId].Type == schema.TYPE_ENUM {
		values := e.Table.Columns[columnId].EnumValues //fields value
		if len(values) == 0 || e.Rows[n][columnId] == nil {
			return ""
		}
		return values[e.Rows[n][columnId].(int64)-1] // first id in result is zero one in values
	}
}
```
### 存储JSON
这难道不是个好主意吗？json是MySQL引擎侧的字符串，我们可以将序列化的数据指向解析器。为此，可以添加一个自定义gorm注释——“fromJson”，以下是不同数据之间的例子：

```
type JsonData struct {
	Int        int               `gorm:"column:int"`
	StructData TestData          `gorm:"column:struct_data;fromJson"`
	MapData    map[string]string `gorm:"column:map_data;fromJson"`
	SliceData  []int             `gorm:"column:slice_data;fromJson"`
}
type TestData struct {
	Test string `json:"test"`
	Int  int    `json:"int"`
}
```
虽然可以创造很多条件来实现，但是新增字段会损坏它。上Stack Overflow寻找答案的结果可能是，“如何从未知的json结构体解析?”"不知道你为什么需要这样，但你可以试试..."  
将结构体转换为接口可以实现：

```
if _, ok := parsedTag["FROMJSON"]; ok {
	newObject := reflect.New(s.Field(k).Type()).Interface()
	json := m.stringHelper(e, n, columnName)
	jsoniter.Unmarshal([]byte(json), &newObject)
	s.Field(k).Set(reflect.ValueOf(newObject).Elem().Convert(s.Field(k).Type()))
}
```
如果还有问题、更正或者建议，欢迎提出。此外，需要校对的地方可以在这里提出：<https://github.com/JackShadow/go-binlog-example/blob/master/src/parser_test.go>  
代码示例：<https://github.com/JackShadow/go-binlog-example>  
特别感谢：[Freadm Project.](https://freadm.com/start/?lang=en)

via：<https://blog.golang.org/using-go-modules>  
作者：[Artem Zheltak](https://medium.com/@infinity.jacksparrow)   
译者：[sz233](https://github.com/sz233)  
校对：[校对者ID](https://github.com/校对者ID)  

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉
