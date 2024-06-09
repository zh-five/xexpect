// 自动登录ssh，并停留在交互式shell

xe_run(["ssh", "xr@127.0.0.1"]); // 运行命令

// 执行多个匹配，默认命中任意一个就返回
xe_matchs([
    ["yes/no", "yes\n", "C"],       // "C"(continue)标志表示命中后不退出，继续匹配
    ["password", "ya8j3fpa*ed\n"],
]);
xe_matchs([["$", "cd /data/git/\n"]]);

xe_term(); // 停留在交互式终端，若直接结束则可调用 xe_exit()