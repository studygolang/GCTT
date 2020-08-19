# 如何用 Go 编写词法分析器

*词法分析器是所有现代编译器的第一阶段，但是如何编写呢？让我们用 Go 从头开始构建一个。*

![lexer](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200601-How-to-Write-a-Lexer-in-Go/how-to-write-a-lexer-in-go-featured.jpg)

## 什么是词法分析器？

词法分析器有时也称为扫描器，它读取源程序并将输入转换为标记流。这是编译过程中非常重要的一步，因为解析器使用这些标记来创建 AST（抽象语法树）。如果你不熟悉解析器和 AST 也不用担心，这篇文章将只专注于构建词法分析器。

## 词法分析器实战

在我们编写自己的词法分析器之前，让我们看一下 Go 的词法分析器，以便更好地了解“标记”的含义。你可以使用 Go 的 scanner package 来仔细查看 Go 的词法分析器到底发出了什么标记。这个 [文档](https://golang.org/pkg/go/scanner/#Scanner.Scan) 中有实现这一点的示例。这是我们测试 scanner 的程序：

*main.go*
```go
package main

import "fmt"

func main() {
	fmt.Println("Hello world!")
}
```

这是发出的标记：

*输出*
```
1:1     package "package"
1:9     IDENT	"main"
1:13    ;	    "\n"
3:1	    import	"import"
3:8	    STRING  "\"fmt\""
3:13	;		"\n"
5:1	    func	"func"
5:6	    IDENT	"main"
5:10	(		""
5:11	)		""
5:13	{		""
6:2	    IDENT	"fmt"
6:5	    .		""
6:6	    IDENT	"Println"
6:13	(		""
6:14	STRING	"\"Hello world!\""
6:28	)		""
6:29	;		"\n"
7:1	    }		""
7:2	    ;		"\n"
```

第一列包含标记的位置，第二列是标记的类型，最后一列是标记的字面值。这里有一些重要的事情需要注意。首先，词法分析器不发出任何制表符或空格。这是因为 Go 的语法并不依赖于这些东西。另一件需要注意的事情是 **IDENT** 标记。基本上，在 Go 中任何不是关键字的东西都将被标记为标识符，伴随的字符串将被标记为字面值。

如果你想知道这些标记为什么有用，那是因为它们以解析器能够理解的方式表示源程序！

## 语言

现在我们已经在实践中了解了词法分析器，让我们从头开始构建自己的词法分析器！我们首先需要为编程语言定义语法。为简单起见，只包括一些不同的东西：

*program → expr**

*expr → assignment | infixExpr | **int***

*assignment → **id** = expr ;*

*infixExp → expr infixOp expr ;*

*infixOp → **+** | **-** | **\*** | **/***

任何粗体被称为 *terminal*，意味着它不能被进一步扩展。在构建词法分析器时，*terminals* 在构建词法分析器时非常重要，稍后我们将看到这一点。语法可以这样读取：

*程序是零个或多个表达式组成的列表。表达式可以是赋值、中缀表达式或整数等等。*

在构建解析器时，语法变得更加重要，但是现在定义语法很重要，这样才能知道词法分析器应该发出哪些标记！

## 标记

上面的语法使我们可以定义词法分析器在扫描时应发出的标记。标记只是 *terminals*！我们还将包括一个 **EOF** 和 **ILLEGAL** 标记，以便分别表示程序的结束和语言中不合法的字符。

*lexer.go*
```go
type Token int

const (
	EOF = iota
	ILLEGAL
	IDENT
	INT
	SEMI // ;

	// 中缀操作
	ADD // +
	SUB // -
	MUL // *
	DIV // /

	ASSIGN // =
)

var tokens = []string{
	EOF:     "EOF",
	ILLEGAL: "ILLEGAL",
	IDENT:   "IDENT",
	INT:     "INT",
	SEMI:    ";",

	// 中缀操作
	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",

	ASSIGN: "=",
}

func (t Token) String() string {
	return tokens[t]
}
```

## 扫描输入

现在我们准备扫描源程序并发出一些标记！首先，我们将创建一个保留某些状态的 Lexer 结构：

*lexer.go*
```go
type Position struct {
	line   int
	column int
}

type Lexer struct {
	pos    Position
	reader *bufio.Reader
}

func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		pos:    Position{line: 1, column: 0},
		reader: bufio.NewReader(reader),
	}
}
```

调用者需要创建带有适当源文件的 reader，并在创建 Lexer 时将其作为参数传递。

接下来，添加一个每次返回单个标记的 `Lex` 函数。然后，调用者将能够连续调用 `Lex`，直到返回 EOF 标记。首先，处理输入文件末尾的情况。

*lexer.go*
```go
// Lex 扫描输入中的下一个标记。返回标记的位置，标记的类型和字面值。
func (l *Lexer) Lex() (Position, Token, string) {
	// 循环直到返回一个标记为止
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return l.pos, EOF, ""
			}

			// 在这一点上我们无能为力，编译器应该将原始错误返回给用户
			panic(err)
		}
	}
}
```

然后，添加一些逻辑来处理语法中的一些更基本的 *terminals*。我们可以使用 switch 语句来检查是否遇到了以下 *terminals* 之一：

*lexer.go*
```go
func (l *Lexer) Lex() (Position, Token, string) {
	// 循环直到返回一个标记位置
    for {
        …
        // 将列更新为新读取的字符的位置
        l.pos.column++

        switch r {
        case '\n':
            l.resetPosition()
        case ';':
            return l.pos, SEMI, ";"
        case '+':
            return l.pos, ADD, "+"
        case '-':
            return l.pos, SUB, "-"
        case '*':
            return l.pos, MUL, "*"
        case '/':
            return l.pos, DIV, "/"
        case '=':
            return l.pos, ASSIGN, "="
        default:
            if unicode.IsSpace(r) {
                continue
            }
        }
    }
}

func (l *Lexer) resetPosition() {
	l.pos.line++
	l.pos.column = 0
}
```

这使我们可以对除标识符和整数之外的所有内容进行 lex，这非常简洁！接下来我们处理整数。我们需要检测是否看到了数字。如果有的话，扫描后面剩下的数字来标记这个整数。

*lexer.go*
```go
func (l *Lexer) Lex() (Position, Token, string) {
	// 循环直到返回一个标记为止
	for {
		…
		switch r {
		…
		default:
			if unicode.IsSpace(r) {
				continue
			} else if unicode.IsDigit(r) {
				// 备份并让 lexInt 重新扫描 int 的开头
				startPos := l.pos
				l.backup()
				lit := l.lexInt()
				return startPos, INT, lit
			}
        }
    }
}

func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}

	l.pos.column--
}

// lexInt 扫描输入直到整数的结尾，然后返回字面值。
func (l *Lexer) lexInt() string {
	var lit string
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return lit
			}
		}

		l.pos.column++
		if unicode.IsDigit(r) {
			lit = lit + string(r)
		} else {
			// 不是整型
			l.backup()
			return lit
		}
	}
}
```

如你所见，`lexInt` 仅扫描所有连续的数字，然后返回字面值。处理标识符可以用类似的方式完成，但是，我们应该定义标识符中哪些字符有效。对于我们的语言，我们只允许使用大写和小写字母，其他所有内容都应视为 **ILLEGAL** 标记。

*lexer.go*
```go
// Lex 扫描输入中的下一个标记。它返回标记的位置，标记的类型和字面值。
func (l *Lexer) Lex() (Position, Token, string) {
	// 循环直到返回一个标记为止
	for {
		...
		switch r {
		...
		default:
			if unicode.IsSpace(r) {
				continue
			} else if unicode.IsDigit(r) {
				// 备份并让 lexInt 重新扫描 int 的开头
				startPos := l.pos
				l.backup()
				lit := l.lexInt()
				return startPos, INT, lit
			} else if unicode.IsLetter(r) {
				// 备份并让 lexIdent 重新扫描 ident 的开头
				startPos := l.pos
				l.backup()
				lit := l.lexIdent()
				return startPos, IDENT, lit
			} else {
				return l.pos, ILLEGAL, string(r)
			}
		}
	}
}

// lexIdent 扫描输入，直到标识符结尾，然后返回字面值。
func (l *Lexer) lexIdent() string {
	var lit string
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				// 到达标识符末尾
				return lit
			}
		}

        l.pos.column++
		if unicode.IsLetter(r) {
			lit = lit + string(r)
		} else {
			// 扫描到标识符中没有的东西
			l.backup()
			return lit
		}
	}
}
```

请务必注意，词法分析不会捕获诸如未定义的变量或无效的语法之类的错误。它唯一关心的是词法正确性（即我们的语言中允许使用的字符）。下面是运行词法分析器并查看输出的方法：

*lexer.go*
```go
func main() {
	file, err := os.Open("input.test")
	if err != nil {
		panic(err)
	}

	lexer := NewLexer(file)
	for {
		pos, tok, lit := lexer.Lex()
		if tok == EOF {
			break
		}

		fmt.Printf("%d:%d\t%s\t%s\n", pos.line, pos.column, tok, lit)
	}
}
```

如果通过运行 `go run lexer.go` 并使用以下输入运行我们的词法分析器，你将看到发出了标记！

*input.test*
```
a = 5;
b = a + 6;
c + 123;
5+12;
```

*输出*
```
1:1     IDENT	a
1:3	    =		=
1:5	    INT		5
1:6	    ;		;
2:1	    IDENT	b
2:3	    =		=
2:5	    IDENT	a
2:7	    +		+
2:9	    INT		6
2:10	;		;
3:1	    IDENT	c
3:3	    +		+
3:5	    INT		123
3:8	    ;		;
4:1	    INT		5
4:2	    +		+
4:3	    INT		12
4:5	    ;		;
```

希望你能从这篇文章中学到一些东西，我也很乐意在评论区听到你的反馈。这篇文章的所有代码都可以在我的 [GitHub](https://github.com/aaronraff/blog-code/tree/master/how-to-write-a-lexer-in-go) 上找到。里面还包含我未介绍的单元测试。

---
via: https://www.aaronraff.dev/blog/how-to-write-a-lexer-in-go

作者：[Aaron RaffLogo](https://www.aaronraff.dev/)
译者：[alandtsang](https://github.com/alandtsang)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
