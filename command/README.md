#  goravel 项目根据表创建model 命令





## 安装依赖

```go
go get -u github.com/clgwxg/goravel@latest
```



## 注册命令在/app/console/kernel.go文件中

```
import (
    // 导入包--------------------------
	"github.com/clgwxg/goravel/command"
	
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"
)

func (kernel Kernel) Commands() []console.Command {
	return []console.Command{
	    // 注册命令----------------------
		command.NewCreateModelCommand(),
	}
}
```



## 运行创建model命令

```shell
go run . artisan create:model  -t  tableName  // tableName 数据库中表的名字
```

