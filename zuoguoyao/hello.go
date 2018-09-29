package main
import "fmt"
func main() {
	const LENGTH, WIDTH = 1000, 500
	var area int
	const a, b, c = 5, false, "str"
	area = LENGTH * WIDTH
	fmt.Println("面积：", area, a, b, c)
}