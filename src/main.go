// requiz v0.0.0 —— 题库管理系统 CLI
package main

import (
	"fmt"
	"os"
)

const version = "1.3.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	// NOTE: 由于 build 从 src 目录编译，命令默认工作目录即当前目录
	var err error
	switch os.Args[1] {
	case "connect":
		err = cmdConnect(os.Args[2:])
	case "list":
		err = cmdList(os.Args[2:])
	case "read":
		err = cmdRead(os.Args[2:])
	case "view":
		err = cmdView(os.Args[2:])
	case "serve":
		err = cmdServe(os.Args[2:])
	case "version":
		fmt.Printf("requiz v%s\n", version)
	case "help", "-h", "--help":
		usage()
	default:
		usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`requiz v1.0.0 —— 题库管理系统

用法:
  requiz connect <题库目录>      连接题库文件夹（验证 .requiz 配置、读取题目）
  requiz list [题库目录]         列出题库题目（缺省当前目录）
  requiz read <id|文件名> [题库目录]   读取单题全文（含元数据）
  requiz view <id|文件名> [题库目录] [选项]  按需显示题目部分
  requiz serve [题库目录] [-port 端口]  启动 localhost Web 服务（默认 127.0.0.1:8080）

view 选项:
  --a     显示答案
  --e     显示解析
  --n     显示备注
  --yaml  显示元数据
  缺省时只显示题干

示例:
  requiz connect demo/题库A
  requiz list demo/题库A
  requiz read M001 demo/题库A
  requiz view M001 demo/题库A --a --e
  requiz serve demo/题库A -port 8080  （浏览器打开 http://127.0.0.1:8080/）`)
}