首发于：https://studygolang.com/articles/23459

# ORM vs. 非 ORM

我一直很喜欢使用 Go 的 database/sql 包来处理数据库。最近，一些涉及 Gorm 的问题激起了我对 Go 中 `使用 ORM` vs. `直接使用 database/sql` 的好奇心。在 ORM 方面曾有过丰富的经验，所以我决定开始一个实验：利用 Gorm 和 非 ORM 编写同一个简单的应用程序，并比较付诸的努力。

这促使我写下了一些关于 ORM 优缺点的想法。如果您对此感兴趣，请继续阅读！

## 非 ORM vs. ORM 的相关经验

实验中，定义了一个可作为博客引擎子集的简单数据库，同时编写一些操作和查询该数据库的 Go 代码，并比较使用纯 SQL 与使用 ORM 的表现。

数据库表如下：

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/orm-or-not-orm/ormdbschema.png)

尽管很简单，这些表展示了一个惯用的、规范的数据库，基本包含构建简单 wiki 或博客应用程序所需的所有元素：它同时具有一对多的关系（帖子与评论）和多对多关系（帖子与标签）。如果您更喜欢数据库 SQL 语句，这是[代码示例](https://github.com/eliben/code-for-blog/tree/master/2019/orm-vs-no-orm/sql)中的定义：

```sql
create table Post (
    postID integer primary key,
    published date,
    title text,
    content text
);

create table Comment (
    commentID integer primary key,
    postID integer,
    author text,
    published date,
    content text,

    -- One-to-many relationship between Post and Comment; each Comment
    -- references a Post it's logically attached to.
    foreign key(postID) references Post(postID)
);

create table Tag (
    tagID integer primary key,
    name text unique
);

-- Linking table for the many-to-many relationship between Tag and Post
create table PostTag (
    postID integer,
    tagID integer,

    foreign key(postID) references Post(postID),
    foreign key(tagID) references Tag(tagID)
);
```

这个 SQL 用 SQLite 测试过；其他 RDBMS 可能需要进行微调。 使用 Gorm 时，没必要写此 SQL；作为替代，我们定义“对象”（实际上是结构体）, 并附带上 Gorm 的一些魔法 tag：

```go
type Post struct {
  gorm.Model
  Published time.Time
  Title     string
  Content   string
  Comments  []Comment `gorm:"foreignkey:PostID"`
  Tags      []*Tag    `gorm:"many2many:post_tags;"`
}

type Tag struct {
  gorm.Model
  Name  string
  Posts []*Post `gorm:"many2many:post_tags;"`
}

type Comment struct {
  gorm.Model
  Author    string
  Published time.Time
  Content   string
  PostID    int64
}
```

使用此数据库的代码有两种变体：

* no-ORM：通过 database/sql 包使用纯 SQL 查询；
* ORM：使用 Gorm 库进行数据库访问。

示例正在做几件事：

1. 将一些数据（帖子、评论、标签）添加到数据库;
2. 查询带有指定标签的所有帖子;
3. 查询所有帖子详细信息（包括附加到其上的所有评论、标记）。

举个例子，这里是任务（2）的两个变体： 查找带有指定标签的所有帖子（这可能是在博客上填充某种档案列表页面）。

* 首先，no-ORM:

```go
func dbAllPostsInTag(db *sql.DB, tagID int64) ([]post, error) {
  rows, err := db.Query(`
    select Post.postID, Post.published, Post.title, Post.content
    from Post
    inner join PostTag on Post.postID = PostTag.postID
    where PostTag.tagID = ?`, tagID)
  if err != nil {
    return nil, err
  }
  var posts []post
  for rows.Next() {
    var p post
    err = rows.Scan(&p.Id, &p.Published, &p.Title, &p.Content)
    if err != nil {
      return nil, err
    }
    posts = append(posts, p)
  }
  return posts, nil
}
```

如果您了解 SQL，这种方式相当直接。我们需要在 `Post` 和 `PostTag` 之间建立一个内连接,并使用 `tagID` 进行条件过滤; 其余代码仅仅迭代结果。

* 接下来，ORM：

```go
func allPostsInTag(db *gorm.DB, t *Tag) ([]Post, error) {
  var posts []Post
  r := db.Model(t).Related(&posts, "Posts")
  if r.Error != nil {
    return nil, r.Error
  }
  return posts, nil
}
```

在 ORM 代码中，为获得相同的效果, 我们倾向于直接使用对象（此处为 `Tag` ）而非 ID。由 Gorm 生成的 SQL 查询与我在 no-ORM 变体中手动编写的 SQL 查询几乎相同。

除了为我们生成 SQL 之外，Gorm 还提供了一种更简单的方法来填充结果。在使用 database/sql 的代码中，我们显式地迭代结果，将每一行分别扫描到单独的结构体字段中。Gorm 的相关方法（以及其他类似的查询方法）将自动填充结构体，并且还将一次扫描整个结果集。

随意玩代码！令我惊喜的是 Gorm 在此节约代码量（对于 DB 部分的代码，节省约 50％），并且对于这些简单的查询，使用 Gorm 并不难：直接从 API 文档中获取调用方式。我对具体示例的唯一抱怨是，在 Post 和 Tag 之间设置多对多关系有点困难，Gorm 字段的 tag 看起来也很丑陋和魔幻。

## 分层的复杂性让人头疼

像上面那样的简单实验的问题在于，通常很难勾勒出系统的边界。它显然适用于简单的情况，但我有兴趣了解当它被推到极限时会发生什么：它如何处理复杂的查询和数据库模式(schema)？因此我开始浏览 Stack Overflow，那儿有许多与 Gorm 相关的问题，当然足以确信，通常的分层复杂性问题是显而易见的（例 1， 例 2）。让我解释一下我的意思。

当包装层本身很复杂时，任何将复杂功能包含在其中的情况，都有增加整体复杂性的风险。这通常伴随着 `leaky abstractions`: 包裹层无法完成包装底层功能的完美工作，将迫使程序员同时与两个层进行斗争。

不幸的是，Gorm 非常容易受到这个问题的影响。Stack Overflow 提供了无穷无尽的问题，用户最终需应对由 Gorm 本身强加的复杂性，解决其局限性等问题。很少有事情如此让人恼火：确切地知道您想要什么（例如，您想要发出哪个 SQL 查询），但是却无法编写出 Gorm 查询时最终调用的正确代码。

## 使用 ORM 的利弊

从我的实验中可以明显看出使用 ORM 的一个关键优势：它可以节省相当多的繁琐编码。以 DB 为中心的代码节省约 50％ 是非常重要的，这可以为某些应用程序带来真正的改变；另一个不明显的优点是从不同的数据库后端抽象。然而，这在 Go 中可能不是一个问题，因为 database/sql 已经提供了一个很好的可移植层。在缺乏标准化 SQL 访问层的语言中，这种优势更加强大。

至于缺点：

1. 要学习另一层，包括所有特性，特殊语法，魔法标签等。如果您已经熟悉 SQL 本身，那么这主要是一个缺点;
2. 即使您没有 SQL 经验，也有大量的知识库和许多可以帮助解答的人。任何一个 ORM 都是更加晦涩的知识，不为很多人所分享，您将花费大量的时间弄清楚如何使其工作;
3. 调试查询性能具有挑战性，因为我们从 `metal` 进一步抽象了一个级别。有时需要进行相当多的调整才能让 ORM 为您生成正确的查询，当您已经知道需要哪些查询时，这很令人沮丧。

最后，一个缺点只会在长期内变得明显：虽然 SQL 多年来保持相当稳定，但 ORM 是特定于语言的，并且往往会出现和消失。每种流行语言都有各种各样的 ORM 可供选择; 当您从一个团队/公司/项目转移到另一个团队/公司/项目时，您可能需要转换，这是额外的精神负担。或者您可能完全切换语言。SQL 是一个更加稳定的层，可以跨团队/语言/项目与您保持联系。

## 结论

使用原生 SQL 实现了一个简单的应用程序框架，并将其与使用 Gorm 的实现进行了比较后，我可以看到 ORM 在减少格式化代码方面的吸引力。我也记得多年前自己是一个 DB 新手时，使用 Django 及其 ORM 来实现一个应用程序：它很好！我没有必要过多考虑 SQL 或底层数据库，它就可以。但那个用例确实非常简单。

随着我经验越来越丰富，我也看到使用 ORM 的许多缺点。尤其，我不认为在 Go 这种语言中 ORM 对我有用，因为 Go 已经拥有一个很好的 SQL 接口，几乎可以跨数据库后端移植。我宁愿花多一点时间敲代码，但这样可以节省我阅读 ORM 文档、优化查询、尤其是调试的时间。

如果您的工作是编写大量简单的类似 CRUD 的应用程序，那么我可以看到 ORM 在 Go 中仍然有用，其节省的代码量克服了这些缺点。最后，所有这些都归结为这一中心论点即， [Benefits of dependencies in software projects as a function of effort](https://eli.thegreenplace.net/2017/benefits-of-dependencies-in-software-projects-as-a-function-of-effort/)：在我看来，在一个并不属于简单的 CRUD 应用程序上，于 DB 接口相关代码之外花费大量精力，ORM 依赖并不值得。

---

via: https://eli.thegreenplace.net/2019/to-orm-or-not-to-orm/

作者：[Eli Bendersky](https://eli.thegreenplace.net/)
译者：[zhoudingding](https://github.com/dingdingzhou)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
