# xexpect：跨平台 expect 命令和 go 库

xexpect 命令:
- 支持 linux 和 windows (cmd，powershell，git bash...)
- 支持交互式和非交互式脚本环境
- 内嵌 javascript 引擎(完整支持 ECMAScript 5.1)
- 支持js脚本加密，增加密码安全性

go 库文档：[xexpect 库]([./readme_lib.md](https://github.com/zh-five/xexpect/blob/main/readme_lib.md))

# 功能速览
### 1.命令参数
```shell
xexpect -h           
Usage of xexpect:
  -c string
        code, 运行xexpect js代码
  -e string
        encrypt, 加密xexpect js文件
  -f string
        file, 运行xexpect js文件

```
当没有任何参数时，xexpect 命令将从标准输入读取js代码。

### 2.多种执行方式示例

以下4个命令完全等效：登录ssh后停留在交互式终端
```shell

# 1.指定js文件路径
xexpect -f ssh.js

# 2.重定向文件到命令的标准输入 
xexpect < ssh.js

# 3.在参数里写js代码
xexpect -c 'xe_run(["ssh", "xr@127.0.0.1"]);xe_matchs([["yes/no", "yes\n", "C"],["password", "123456\n"]]);xe_matchs([["$", "cd /data/git/\n"]]);xe_term();'

# 4.shell中编写js代码，然后导入到xexpect命令的标准输入
xexpect <<EOF
xe_run(["ssh", "xr@127.0.0.1"]); // 运行命令

// 执行多个匹配，默认命中任意一个就返回
xe_matchs([
    ["yes/no", "yes\n", "C"],       // "C"(continue)标志表示命中后不退出，继续匹配
    ["password", "123456\n"],
]);
xe_matchs([["$", "cd /data/git/\n"]]); // 登录后打开指定目录

xe_term(); // 停留在交互式终端，若要结束则调用 xe_exit()
EOF

```
 `ssh.js` 文件内容为：
```js
// 自动登录ssh，并停留在交互式shell

xe_run(["ssh", "xr@127.0.0.1"]); // 运行命令

// 执行多个匹配，默认命中任意一个就返回
xe_matchs([
    ["yes/no", "yes\n", "C"],       // "C"(continue)标志表示命中后不退出，继续匹配
    ["password", "123456\n"],
]);
xe_matchs([["$", "cd /data/git/\n"]]); // 登录后打开指定目录

xe_term(); // 停留在交互式终端，若要结束则调用 xe_exit()
```

### 3.加密js文件
```shell
# 加密js文件, 
xexpect -e ssh.js
# output: encrypt to file: ssh.js.xexpect

# 加密后的js文件，只能通过 -f 加载执行
xexpect -f ssh.js.xexpect

```


# 安装


# 使用手册


# 感谢与依赖库
- 跨平台pty的go实现： https://github.com/iyzyi/aiopty
- 纯go实现的js引擎：  https://github.com/dop251/goja